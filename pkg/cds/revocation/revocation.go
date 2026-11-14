// Package revocation checks the revocation status of a certificate with the CRL that the certificate gives.
package revocation

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrRevoked = errors.New("the certificate is revoked")
	// ErrStatusUnknown means that the caller must refuse the certificate: no
	// current CRL from its issuer says that it is not revoked.
	ErrStatusUnknown = errors.New("the revocation status of the certificate is not known")
)

const (
	// DefaultMaxAge is the time between two loads of a CRL: short, because a revocation must take effect quickly,
	// and long enough that a burst of connections causes one load.
	DefaultMaxAge = time.Minute
	// RetryInterval is the time after a failed load before the next load, also when the maximum age is shorter, so
	// that a CRL server that does not reply does not delay each handshake.
	RetryInterval = time.Minute
)

const (
	maxCRLBytes    = 10 << 20
	requestTimeout = 5 * time.Second
)

type Checker struct {
	Now func() time.Time
	// With a MaxAge of 0, each check loads the CRL.
	MaxAge time.Duration

	issuers []*x509.Certificate
	client  *http.Client
	mu      sync.Mutex
	cache   map[string]cached
}

type cached struct {
	crl     *x509.RevocationList
	checked time.Time
	// failed is true when the load at checked failed, and crl is the last CRL.
	failed bool
}

// NewChecker accepts only CRLs that one of the issuers signed.
func NewChecker(issuers []*x509.Certificate) *Checker {
	return &Checker{
		Now:     time.Now,
		MaxAge:  DefaultMaxAge,
		issuers: issuers,
		client:  &http.Client{Timeout: requestTimeout},
		cache:   map[string]cached{},
	}
}

func (c *Checker) Check(ctx context.Context, cert *x509.Certificate) error {
	issuer := c.issuer(cert)
	if issuer == nil {
		return fmt.Errorf("%w: no trusted issuer signed it", ErrStatusUnknown)
	}

	var errs []error

	for _, crlURL := range cert.CRLDistributionPoints {
		if !strings.HasPrefix(crlURL, "http://") && !strings.HasPrefix(crlURL, "https://") {
			continue
		}

		crl, err := c.currentCRL(ctx, crlURL, issuer)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for _, entry := range crl.RevokedCertificateEntries {
			if entry.SerialNumber.Cmp(cert.SerialNumber) == 0 {
				return ErrRevoked
			}
		}

		return nil
	}

	if len(errs) == 0 {
		return fmt.Errorf("%w: it gives no HTTP CRL URL", ErrStatusUnknown)
	}

	return fmt.Errorf("%w: %w", ErrStatusUnknown, errors.Join(errs...))
}

func (c *Checker) issuer(cert *x509.Certificate) *x509.Certificate {
	for _, issuer := range c.issuers {
		if cert.CheckSignatureFrom(issuer) == nil {
			return issuer
		}
	}

	return nil
}

// currentCRL holds the lock during the request, so that concurrent handshakes load the CRL one time.
// When a load fails, it uses the last CRL until the next update of that CRL (RFC 5280, section 5.1.2.5).
func (c *Checker) currentCRL(
	ctx context.Context,
	crlURL string,
	issuer *x509.Certificate,
) (*x509.RevocationList, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.Now()
	last, found := c.cache[crlURL]
	current := found && now.Before(last.crl.NextUpdate)

	age := now.Sub(last.checked)
	if current && (age < c.MaxAge || (last.failed && age < RetryInterval)) {
		return last.crl, nil
	}

	crl, err := c.load(ctx, crlURL, issuer, now)
	if err != nil {
		if !current {
			return nil, err
		}

		slog.WarnContext(ctx, "the CRL could not be loaded again; the last CRL is in use",
			slog.String("url.full", crlURL), slog.Any("error", err))
		c.cache[crlURL] = cached{crl: last.crl, checked: now, failed: true}

		return last.crl, nil
	}

	c.cache[crlURL] = cached{crl: crl, checked: now}

	return crl, nil
}

func (c *Checker) load(
	ctx context.Context,
	crlURL string,
	issuer *x509.Certificate,
	now time.Time,
) (*x509.RevocationList, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, crlURL, nil)
	if err != nil {
		return nil, fmt.Errorf("make the CRL request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get the CRL: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the CRL server gave status %d", resp.StatusCode)
	}

	der, err := io.ReadAll(io.LimitReader(resp.Body, maxCRLBytes))
	if err != nil {
		return nil, fmt.Errorf("read the CRL: %w", err)
	}

	crl, err := x509.ParseRevocationList(der)
	if err != nil {
		return nil, fmt.Errorf("parse the CRL: %w", err)
	}

	if err := crl.CheckSignatureFrom(issuer); err != nil {
		return nil, fmt.Errorf("check the CRL signature: %w", err)
	}

	if !now.Before(crl.NextUpdate) {
		return nil, errors.New("the CRL is after its next update")
	}

	return crl, nil
}
