// Package recipientnotice sends the notice of the end of an arrangement to the CDR arrangement revocation endpoint
// of a software product.
package recipientnotice

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

const (
	revocationPath  = "/arrangements/revoke"
	jwtLifetime     = 5 * time.Minute
	responseTimeout = 10 * time.Second
)

var _ app.NoticeSender = Sender{}

// NewClient gives a client that does not follow a redirect, because a redirect could send the notice to a host that
// the registration does not name.
func NewClient(transport http.RoundTripper) *http.Client {
	return &http.Client{
		Transport:     otelhttp.NewTransport(transport),
		Timeout:       responseTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

type Sender struct {
	BrandID string
	// Keys gives the keys of the data holder JWKS. The first key signs.
	Keys   func(context.Context) ([]authserver.SigningKey, error)
	Client *http.Client
	NewID  func() string
	Now    func() time.Time
}

type claims struct {
	jwt.Claims

	ArrangementID string `json:"cdr_arrangement_id,omitempty"`
}

func (s Sender) Send(ctx context.Context, notice app.Notice) (int, error) {
	endpoint := strings.TrimSuffix(notice.RecipientBaseURI, "/") + revocationPath

	signer, err := s.signer(ctx)
	if err != nil {
		return 0, err
	}

	bearer, err := jwt.Signed(signer).Claims(s.claims(endpoint, "")).Serialize()
	if err != nil {
		return 0, fmt.Errorf("sign the bearer JWT: %w", err)
	}

	arrangementJWT, err := jwt.Signed(signer).Claims(s.claims(endpoint, notice.ArrangementID)).Serialize()
	if err != nil {
		return 0, fmt.Errorf("sign the cdr_arrangement_jwt: %w", err)
	}

	body := url.Values{"cdr_arrangement_jwt": {arrangementJWT}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create the request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return 0, classify(err)
	}

	_ = resp.Body.Close()

	return resp.StatusCode, nil
}

func (s Sender) signer(ctx context.Context) (jose.Signer, error) {
	keys, err := s.Keys(ctx)
	if err != nil {
		return nil, fmt.Errorf("read the signing keys: %w", err)
	}

	if len(keys) == 0 {
		return nil, errors.New("the data holder has no signing key")
	}

	return jose.NewSigner(
		jose.SigningKey{Algorithm: jose.PS256, Key: jose.JSONWebKey{Key: keys[0].Key, KeyID: keys[0].ID}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
}

func (s Sender) claims(endpoint, arrangementID string) claims {
	now := s.Now()

	return claims{
		Issuer:        s.BrandID,
		Subject:       s.BrandID,
		Audience:      jwt.Audience{endpoint},
		ID:            s.NewID(),
		IssuedAt:      jwt.NewNumericDate(now),
		Expiry:        jwt.NewNumericDate(now.Add(jwtLifetime)),
		ArrangementID: arrangementID,
	}
}

func classify(err error) error {
	var (
		netErr     net.Error
		verifyErr  *tls.CertificateVerificationError
		unknownCA  x509.UnknownAuthorityError
		hostErr    x509.HostnameError
		invalidErr x509.CertificateInvalidError
		recordErr  tls.RecordHeaderError
	)

	switch {
	case errors.As(err, &netErr) && netErr.Timeout(), errors.Is(err, context.DeadlineExceeded):
		return app.ErrNoticeTimeout
	case errors.As(err, &verifyErr), errors.As(err, &unknownCA), errors.As(err, &hostErr),
		errors.As(err, &invalidErr), errors.As(err, &recordErr):
		return app.ErrNoticeTLS
	default:
		return app.ErrNoticeNoConnection
	}
}
