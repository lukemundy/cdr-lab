// Package registrationjwt verifies the JWTs of a registration request.
package registrationjwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var (
	// The Register signs an SSA with PS256.
	ssaAlgs     = []jose.SignatureAlgorithm{jose.PS256}
	requestAlgs = []jose.SignatureAlgorithm{jose.PS256, jose.ES256}
)

type Verifier struct {
	registerKeys *RegisterKeys
	// client gets the JWKS of a software product.
	client *http.Client
}

var _ app.RegistrationJWTs = (*Verifier)(nil)

func NewVerifier(registerKeys *RegisterKeys, transport http.RoundTripper) *Verifier {
	return &Verifier{registerKeys: registerKeys, client: authserver.NewJWKSClient(transport)}
}

func (v *Verifier) SoftwareStatement(requestJWT string) (string, error) {
	parsed, err := jwt.ParseSigned(requestJWT, requestAlgs)
	if err != nil {
		return "", app.ErrInvalidJWT
	}

	var claims struct {
		SoftwareStatement string `json:"software_statement"`
	}
	if err := parsed.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return "", app.ErrInvalidJWT
	}

	return claims.SoftwareStatement, nil
}

func (v *Verifier) VerifySSA(ctx context.Context, ssa string) (domain.SSA, error) {
	parsed, err := jwt.ParseSigned(ssa, ssaAlgs)
	if err != nil {
		return domain.SSA{}, app.ErrInvalidJWT
	}

	key, err := v.registerKeys.key(ctx, parsed.Headers[0].KeyID)
	if errors.Is(err, errKeyNotFound) {
		return domain.SSA{}, app.ErrInvalidJWT
	}

	if err != nil {
		return domain.SSA{}, err
	}

	var claims domain.SSA
	if err := parsed.Claims(key, &claims); err != nil {
		return domain.SSA{}, app.ErrInvalidJWT
	}

	return claims, nil
}

// requestClaims is domain.RegistrationRequest with an aud that can be one string or a list.
type requestClaims struct {
	domain.RegistrationRequest

	Audience jwt.Audience `json:"aud"`
}

func (v *Verifier) VerifyRequest(ctx context.Context, requestJWT, jwksURI string) (domain.RegistrationRequest, error) {
	parsed, err := jwt.ParseSigned(requestJWT, requestAlgs)
	if err != nil {
		return domain.RegistrationRequest{}, app.ErrInvalidJWT
	}

	keys, err := getJWKS(ctx, v.client, jwksURI)
	if err != nil {
		return domain.RegistrationRequest{}, fmt.Errorf("%w: %w", app.ErrKeysNotAvailable, err)
	}

	found := keys.Key(parsed.Headers[0].KeyID)
	if len(found) != 1 {
		return domain.RegistrationRequest{}, app.ErrInvalidJWT
	}

	var claims requestClaims
	if err := parsed.Claims(found[0], &claims); err != nil {
		return domain.RegistrationRequest{}, app.ErrInvalidJWT
	}

	claims.RegistrationRequest.Audience = claims.Audience

	return claims.RegistrationRequest, nil
}
