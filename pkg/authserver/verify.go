package authserver

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// ErrInvalidToken wraps each cause that is the fault of the caller, and other errors are server failures.
var ErrInvalidToken = errors.New("the access token is not valid")

type AccessToken struct {
	ClientID string
	Scopes   []string
}

type accessTokenClaims struct {
	ID           string `json:"jti"`
	ClientID     string `json:"client_id"`
	Scope        string `json:"scope"`
	GrantID      string `json:"grant_id"`
	Confirmation struct {
		Thumbprint string `json:"x5t#S256"`
	} `json:"cnf"`
	ExpiresAt time.Time `json:"-"`
}

// VerifyAccessToken accepts a token only with the client certificate that the token is bound to (RFC 8705).
func (r *Register) VerifyAccessToken(
	ctx context.Context,
	token string,
	clientCert *x509.Certificate,
) (AccessToken, error) {
	claims, err := verifyAccessToken(ctx, token, clientCert, r.issuer, r.cfg.SigningKeys, time.Now())
	if err != nil {
		return AccessToken{}, err
	}

	return claims.accessToken(), nil
}

func (c accessTokenClaims) accessToken() AccessToken {
	return AccessToken{ClientID: c.ClientID, Scopes: strings.Fields(c.Scope)}
}

func verifyAccessToken(
	ctx context.Context,
	token string,
	clientCert *x509.Certificate,
	issuer string,
	signingKeys func(context.Context) ([]SigningKey, error),
	now time.Time,
) (accessTokenClaims, error) {
	claims, err := parseAccessToken(ctx, token, issuer, signingKeys, now)
	if err != nil {
		return accessTokenClaims{}, err
	}

	if clientCert == nil {
		return accessTokenClaims{}, fmt.Errorf("%w: the request has no client certificate", ErrInvalidToken)
	}

	if subtle.ConstantTimeCompare([]byte(thumbprint(clientCert)), []byte(claims.Confirmation.Thumbprint)) != 1 {
		return accessTokenClaims{}, fmt.Errorf("%w: it is bound to a different client certificate", ErrInvalidToken)
	}

	return claims, nil
}

// parseAccessToken does not check the client certificate.
func parseAccessToken(
	ctx context.Context,
	token string,
	issuer string,
	signingKeys func(context.Context) ([]SigningKey, error),
	now time.Time,
) (accessTokenClaims, error) {
	parsed, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.PS256})
	if err != nil || len(parsed.Headers) != 1 {
		return accessTokenClaims{}, fmt.Errorf("%w: it is not a PS256 JWT", ErrInvalidToken)
	}

	keys, err := signingKeys(ctx)
	if err != nil {
		return accessTokenClaims{}, err
	}

	var (
		standard jwt.Claims
		claims   accessTokenClaims
	)

	verified := false

	for _, k := range keys {
		if k.ID == parsed.Headers[0].KeyID && parsed.Claims(&k.Key.PublicKey, &standard, &claims) == nil {
			verified = true
			break
		}
	}

	if !verified {
		return accessTokenClaims{}, fmt.Errorf(
			"%w: no signing key of the authorisation server signed it",
			ErrInvalidToken,
		)
	}

	if standard.Expiry == nil {
		return accessTokenClaims{}, fmt.Errorf("%w: it has no expiry", ErrInvalidToken)
	}

	if err := standard.ValidateWithLeeway(jwt.Expected{Issuer: issuer, Time: now}, 0); err != nil {
		return accessTokenClaims{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if claims.ClientID == "" {
		return accessTokenClaims{}, fmt.Errorf("%w: it has no client_id", ErrInvalidToken)
	}

	// A revocation identifies the token by its jti.
	if claims.ID == "" {
		return accessTokenClaims{}, fmt.Errorf("%w: it has no jti", ErrInvalidToken)
	}

	claims.ExpiresAt = standard.Expiry.Time().UTC()

	return claims, nil
}

// thumbprint is the x5t#S256 value of RFC 8705.
func thumbprint(cert *x509.Certificate) string {
	sum := sha256.Sum256(cert.Raw)

	return base64.RawURLEncoding.EncodeToString(sum[:])
}
