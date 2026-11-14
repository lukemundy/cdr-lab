// Package ssa signs Software Statement Assertions.
package ssa

import (
	"context"
	"errors"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type Signer struct {
	// Keys gives the keys that Get JWKS publishes, and the first key signs.
	Keys func(context.Context) ([]authserver.SigningKey, error)
}

var _ app.SSASigner = Signer{}

func (s Signer) SignSSA(ctx context.Context, claims domain.SSAClaims) (string, error) {
	keys, err := s.Keys(ctx)
	if err != nil {
		return "", err
	}

	if len(keys) == 0 {
		return "", errors.New("the Register has no SSA signing key")
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.PS256, Key: jose.JSONWebKey{Key: keys[0].Key, KeyID: keys[0].ID}},
		(&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return "", err
	}

	return jwt.Signed(signer).Claims(claims).Serialize()
}
