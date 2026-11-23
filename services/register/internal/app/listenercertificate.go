package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

const listenerCertificateRenewalWindow = 30 * 24 * time.Hour

var registerSubject = pkix.Name{
	Country:            []string{"AU"},
	Organization:       []string{"Mock CDR Register"},
	OrganizationalUnit: []string{"Consumer Data Right"},
}

type ListenerCertificate interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	Replace(ctx context.Context, cert tls.Certificate) error
}

// ListenerCertificates records no event for a certificate of the mTLS listener.
type ListenerCertificates struct {
	intermediate ca.Authority
	dnsNames     []string
	now          func() time.Time
}

func NewListenerCertificates(intermediate ca.Authority, dnsNames []string, now func() time.Time) ListenerCertificates {
	return ListenerCertificates{intermediate: intermediate, dnsNames: dnsNames, now: now}
}

// Issue makes a new key, which is only in the memory of the process.
func (c ListenerCertificates) Issue() (tls.Certificate, error) {
	if len(c.dnsNames) == 0 {
		return tls.Certificate{}, errors.New("the certificate of the mTLS listener needs at least one DNS name")
	}

	key, err := rsa.GenerateKey(rand.Reader, ca.LeafKeyBits)
	if err != nil {
		return tls.Certificate{}, err
	}

	subject := registerSubject
	subject.CommonName = c.dnsNames[0]

	// A CSR gives the certificate the checks of the CDS CSR profile.
	der, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject:            subject,
		DNSNames:           c.dnsNames,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	csr, err := x509.ParseCertificateRequest(der)
	if err != nil {
		return tls.Certificate{}, err
	}

	leaf, err := c.intermediate.Issue(csr, ca.Server, c.now())
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("issue the certificate of the mTLS listener: %w", err)
	}

	return tls.Certificate{Certificate: [][]byte{leaf.Raw}, PrivateKey: key, Leaf: leaf}, nil
}

func (c ListenerCertificates) Renew(ctx context.Context, listener ListenerCertificate) error {
	current, err := listener.GetCertificate(nil)
	if err != nil {
		return err
	}

	if c.now().Before(current.Leaf.NotAfter.Add(-listenerCertificateRenewalWindow)) {
		return nil
	}

	cert, err := c.Issue()
	if err != nil {
		return err
	}

	return listener.Replace(ctx, cert)
}

// KeepRenewed compares wall-clock times, so the first check after a suspend of the host renews a due certificate.
func (c ListenerCertificates) KeepRenewed(ctx context.Context, listener ListenerCertificate, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Renew(ctx, listener); err != nil {
				slog.ErrorContext(ctx, "the certificate of the mTLS listener was not renewed", slog.Any("error", err))
			}
		}
	}
}
