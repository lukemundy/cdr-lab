package authserver_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

func (l *lab) issuedToken() string {
	l.t.Helper()
	r := l.token(l.form(l.assertion()))
	require.Equal(l.t, http.StatusOK, r.status, r.body)

	return r.body["access_token"].(string)
}

// craftedToken is a token that the token endpoint does not issue, signed as the test chooses.
func (l *lab) craftedToken(
	alg jose.SignatureAlgorithm,
	key *rsa.PrivateKey,
	kid string,
	change func(claims map[string]any),
) string {
	l.t.Helper()

	now := time.Now()

	claims := map[string]any{
		"iss":       issuer,
		"sub":       productID,
		"client_id": productID,
		"jti":       rand.Text(),
		"scope":     "cdr-register:read",
		"iat":       now.Unix(),
		"exp":       now.Add(5 * time.Minute).Unix(),
		"cnf":       map[string]any{"x5t#S256": thumbprint(l.cert)},
	}
	if change != nil {
		change(claims)
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: alg, Key: jose.JSONWebKey{Key: key, KeyID: kid}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	require.NoError(l.t, err)
	token, err := jwt.Signed(signer).Claims(claims).Serialize()
	require.NoError(l.t, err)

	return token
}

// Verifies: AUTH-0055
func TestAnIssuedTokenIsVerifiedWithTheCertificateThatItIsBoundTo(t *testing.T) {
	l := newLab(t)

	got, err := l.register.VerifyAccessToken(t.Context(), l.issuedToken(), l.cert.Leaf)

	require.NoError(t, err)
	assert.Equal(t, authserver.AccessToken{ClientID: productID, Scopes: []string{"cdr-register:read"}}, got)
}

func TestATokenIsNotValid(t *testing.T) {
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	l := newLab(t)
	otherCert := clientCertificate(t, "Recipient App Web")
	valid := func(change func(map[string]any)) string {
		return l.craftedToken(jose.PS256, l.signingKey, "register-1", change)
	}

	tests := map[string]struct {
		token string
		cert  *x509.Certificate
	}{
		"with a different client certificate": {token: l.issuedToken(), cert: otherCert.Leaf},
		"when it is expired": {
			token: valid(func(c map[string]any) { c["exp"] = time.Now().Add(-time.Second).Unix() }),
			cert:  l.cert.Leaf,
		},
		"when it has no expiry": {
			token: valid(func(c map[string]any) { delete(c, "exp") }),
			cert:  l.cert.Leaf,
		},
		"from a different issuer": {
			token: valid(func(c map[string]any) { c["iss"] = "https://holder.example" }),
			cert:  l.cert.Leaf,
		},
		"with no certificate binding": {
			token: valid(func(c map[string]any) { delete(c, "cnf") }),
			cert:  l.cert.Leaf,
		},
		"with no client_id": {
			token: valid(func(c map[string]any) { delete(c, "client_id") }),
			cert:  l.cert.Leaf,
		},
		"signed with a different key": {
			token: l.craftedToken(jose.PS256, otherKey, "register-1", nil),
			cert:  l.cert.Leaf,
		},
		"signed with an unknown key ID": {
			token: l.craftedToken(jose.PS256, l.signingKey, "unknown", nil),
			cert:  l.cert.Leaf,
		},
		"signed with RS256": {
			token: l.craftedToken(jose.RS256, l.signingKey, "register-1", nil),
			cert:  l.cert.Leaf,
		},
		"when it is not a JWT": {token: "not-a-token", cert: l.cert.Leaf},
		"when it is empty":     {token: "", cert: l.cert.Leaf},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := l.register.VerifyAccessToken(t.Context(), tt.token, tt.cert)

			assert.ErrorIs(t, err, authserver.ErrInvalidToken)
		})
	}
}

func TestACraftedTokenWithValidClaimsIsAccepted(t *testing.T) {
	l := newLab(t)

	_, err := l.register.VerifyAccessToken(
		t.Context(),
		l.craftedToken(jose.PS256, l.signingKey, "register-1", nil),
		l.cert.Leaf,
	)

	assert.NoError(t, err, "the refusals of the crafted tokens come from their changes only")
}

func TestATokenWithNoClientCertificateIsNotValid(t *testing.T) {
	l := newLab(t)

	_, err := l.register.VerifyAccessToken(t.Context(), l.issuedToken(), nil)

	assert.ErrorIs(t, err, authserver.ErrInvalidToken)
}

func TestASigningKeyFailureIsNotAnInvalidToken(t *testing.T) {
	l := newLab(t)
	token := l.issuedToken()
	failure := errors.New("database failure")
	l.signingKeysErr = failure

	_, err := l.register.VerifyAccessToken(t.Context(), token, l.cert.Leaf)

	require.ErrorIs(t, err, failure)
	assert.NotErrorIs(t, err, authserver.ErrInvalidToken)
}

// Verifies: AUTH-0010, AUTH-0056
func TestDuringAKeyChangeTheFirstKeySignsAndATokenOfEachKeyIsValid(t *testing.T) {
	l := newLab(t)
	next, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	l.otherKeys = []authserver.SigningKey{{ID: "register-2", Key: next}}

	issued := l.token(l.form(l.assertion()))
	_, fromNext := l.register.VerifyAccessToken(
		t.Context(),
		l.craftedToken(jose.PS256, next, "register-2", nil),
		l.cert.Leaf,
	)
	_, body := l.getPublic("/idp/.well-known/openid-configuration/jwks")

	l.accessTokenClaims(issued)
	require.NoError(t, fromNext, "a token of the other key in the key set is valid")
	assert.Len(t, body["keys"], 2, "the JWKS has the two keys")
}
