package ssa_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/ssa"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func key(t *testing.T, id string) authserver.SigningKey {
	t.Helper()

	k, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return authserver.SigningKey{ID: id, Key: k}
}

func signerWith(keys []authserver.SigningKey, err error) ssa.Signer {
	return ssa.Signer{Keys: func(context.Context) ([]authserver.SigningKey, error) { return keys, err }}
}

func TestTheSSAIsAPS256JWTSignedWithTheFirstKey(t *testing.T) {
	newest, older := key(t, "ssa-2"), key(t, "ssa-1")

	token, err := signerWith(
		[]authserver.SigningKey{newest, older},
		nil,
	).SignSSA(t.Context(), domain.SSAClaims{Issuer: "cdr-register", SoftwareID: "sp-1"})

	require.NoError(t, err)
	parsed, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(t, err)
	require.Len(t, parsed.Headers, 1)
	assert.Equal(t, "ssa-2", parsed.Headers[0].KeyID)
	assert.Equal(t, "JWT", parsed.Headers[0].ExtraHeaders[jose.HeaderType])

	var claims map[string]any
	require.NoError(t, parsed.Claims(&newest.Key.PublicKey, &claims))
	assert.Equal(t, "cdr-register", claims["iss"])
	assert.Equal(t, "sp-1", claims["software_id"])
	assert.Error(t, parsed.Claims(&older.Key.PublicKey, &claims), "the older key did not sign it")
}

func TestSignSSAFailsWithoutAKey(t *testing.T) {
	failure := errors.New("database failure")

	_, noKey := signerWith(nil, nil).SignSSA(t.Context(), domain.SSAClaims{})
	_, readFailure := signerWith(nil, failure).SignSSA(t.Context(), domain.SSAClaims{})

	require.Error(t, noKey)
	assert.ErrorIs(t, readFailure, failure)
}
