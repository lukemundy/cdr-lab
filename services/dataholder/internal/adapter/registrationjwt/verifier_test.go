package registrationjwt_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationjwt"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

// jwksServer serves a key set that a test can change, and counts the requests.
type jwksServer struct {
	*httptest.Server

	keys     atomic.Pointer[jose.JSONWebKeySet]
	requests atomic.Int32
	status   atomic.Int32
}

func newJWKSServer(t *testing.T, keys ...jose.JSONWebKey) *jwksServer {
	t.Helper()

	s := &jwksServer{}
	s.publish(keys...)
	s.status.Store(http.StatusOK)
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.requests.Add(1)
		w.WriteHeader(int(s.status.Load()))
		_ = json.NewEncoder(w).Encode(s.keys.Load())
	}))
	t.Cleanup(s.Close)

	return s
}

func (s *jwksServer) publish(keys ...jose.JSONWebKey) {
	set := jose.JSONWebKeySet{}
	for _, k := range keys {
		set.Keys = append(set.Keys, k.Public())
	}

	s.keys.Store(&set)
}

func rsaKey(t *testing.T, kid string) jose.JSONWebKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return jose.JSONWebKey{Key: key, KeyID: kid, Algorithm: "PS256", Use: "sig"}
}

func ecKey(t *testing.T, kid string) jose.JSONWebKey {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	return jose.JSONWebKey{Key: key, KeyID: kid, Algorithm: "ES256", Use: "sig"}
}

func sign(t *testing.T, alg jose.SignatureAlgorithm, key jose.JSONWebKey, claims map[string]any) string {
	t.Helper()

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: alg, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))
	require.NoError(t, err)
	signed, err := jwt.Signed(signer).Claims(claims).Serialize()
	require.NoError(t, err)

	return signed
}

type fixture struct {
	registerKey   jose.JSONWebKey
	recipientKey  jose.JSONWebKey
	register      *jwksServer
	recipient     *jwksServer
	registerKeys  *registrationjwt.RegisterKeys
	verifier      *registrationjwt.Verifier
	now           time.Time
	recipientJWKS string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	f := &fixture{
		registerKey:  rsaKey(t, "register-1"),
		recipientKey: rsaKey(t, "recipient-1"),
		now:          time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC),
	}
	f.register = newJWKSServer(t, f.registerKey)
	f.recipient = newJWKSServer(t, f.recipientKey)
	f.recipientJWKS = f.recipient.URL
	f.registerKeys = registrationjwt.NewRegisterKeys(f.register.URL, http.DefaultTransport)
	f.registerKeys.Now = func() time.Time { return f.now }
	f.verifier = registrationjwt.NewVerifier(f.registerKeys, http.DefaultTransport)

	return f
}

func TestSoftwareStatementGivesTheSSAOfARequestThatIsNotVerified(t *testing.T) {
	f := newFixture(t)
	request := sign(t, jose.PS256, rsaKey(t, "unknown"), map[string]any{"software_statement": "a.b.c"})

	statement, err := f.verifier.SoftwareStatement(request)

	require.NoError(t, err)
	assert.Equal(t, "a.b.c", statement)
}

func TestSoftwareStatementRefusesABodyThatIsNotAJWT(t *testing.T) {
	f := newFixture(t)

	bodies := map[string]string{"text": "not a JWT", "empty": "", "JSON": `{"software_statement": "a.b.c"}`}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			_, err := f.verifier.SoftwareStatement(body)

			assert.ErrorIs(t, err, app.ErrInvalidJWT)
		})
	}
}

// Verifies: DH-0610
func TestVerifySSAGivesTheClaimsOfAnSSAThatTheRegisterSigned(t *testing.T) {
	f := newFixture(t)
	ssa := sign(t, jose.PS256, f.registerKey, map[string]any{
		"iss": "cdr-register", "iat": 1, "exp": 2, "jti": "ssa-1", "software_id": "software-1",
		"redirect_uris": []string{"https://recipient.example/cb"}, "scope": "openid cdr:registration",
		"a_claim_that_the_data_holder_does_not_know": true,
	})

	claims, err := f.verifier.VerifySSA(t.Context(), ssa)

	require.NoError(t, err)
	assert.Equal(t, "cdr-register", claims.Issuer)
	assert.Equal(t, "software-1", claims.SoftwareID)
	assert.Equal(t, int64(2), claims.ExpiresAt)
	assert.Equal(t, []string{"https://recipient.example/cb"}, claims.RedirectURIs)
}

// Verifies: DH-0610
func TestVerifySSARefusesAnSSAThatTheRegisterDidNotSign(t *testing.T) {
	f := newFixture(t)
	other := rsaKey(t, "register-1")
	claims := map[string]any{"iss": "cdr-register"}

	tests := map[string]string{
		"a key that is not in the Register JWKS":      sign(t, jose.PS256, rsaKey(t, "unknown"), claims),
		"a different key with the same key ID":        sign(t, jose.PS256, other, claims),
		"an algorithm that the Register does not use": sign(t, jose.RS256, f.registerKey, claims),
		"not a JWT": "not a JWT",
	}
	for name, ssa := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := f.verifier.VerifySSA(t.Context(), ssa)

			assert.ErrorIs(t, err, app.ErrInvalidJWT)
		})
	}
}

// Verifies: DH-0614
func TestVerifySSAGetsTheRegisterJWKSAgainForAKeyIDThatItDoesNotHave(t *testing.T) {
	f := newFixture(t)
	_, err := f.verifier.VerifySSA(t.Context(), sign(t, jose.PS256, f.registerKey, map[string]any{}))
	require.NoError(t, err)

	next := rsaKey(t, "register-2")
	f.register.publish(f.registerKey, next)
	f.now = f.now.Add(registrationjwt.MinRefreshInterval)

	_, err = f.verifier.VerifySSA(t.Context(), sign(t, jose.PS256, next, map[string]any{}))

	require.NoError(t, err)
	assert.EqualValues(t, 2, f.register.requests.Load())
}

func TestVerifySSAGetsTheRegisterJWKSOneTimeForKeysThatItHas(t *testing.T) {
	f := newFixture(t)

	for range 3 {
		_, err := f.verifier.VerifySSA(t.Context(), sign(t, jose.PS256, f.registerKey, map[string]any{}))
		require.NoError(t, err)
	}

	assert.EqualValues(t, 1, f.register.requests.Load())
}

// Verifies: DH-0614
func TestVerifySSADoesNotGetTheRegisterJWKSAgainBeforeTheMinimumInterval(t *testing.T) {
	f := newFixture(t)
	unknown := sign(t, jose.PS256, rsaKey(t, "invented"), map[string]any{})

	for range 3 {
		_, err := f.verifier.VerifySSA(t.Context(), unknown)
		require.ErrorIs(t, err, app.ErrInvalidJWT)
	}

	assert.EqualValues(t, 1, f.register.requests.Load())
}

func TestVerifySSAGivesAServerFailureWhenTheRegisterJWKSIsNotAvailable(t *testing.T) {
	f := newFixture(t)
	f.register.status.Store(http.StatusServiceUnavailable)

	_, err := f.verifier.VerifySSA(t.Context(), sign(t, jose.PS256, f.registerKey, map[string]any{}))

	require.Error(t, err)
	assert.NotErrorIs(t, err, app.ErrInvalidJWT)
}

// Verifies: DH-0622
func TestVerifyRequestGivesTheClaimsOfARequestThatTheSoftwareProductSigned(t *testing.T) {
	f := newFixture(t)
	ec := ecKey(t, "recipient-ec")
	f.recipient.publish(f.recipientKey, ec)

	tests := map[string]struct {
		alg jose.SignatureAlgorithm
		key jose.JSONWebKey
		aud any
	}{
		"PS256, with one audience": {jose.PS256, f.recipientKey, "https://id.bank.example"},
		"ES256, with a list of audiences": {
			jose.ES256,
			ec,
			[]string{"https://id.bank.example", "https://other.example"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			//nolint:gosec // G101 reads the protocol field id_token_encrypted_response_alg as a credential
			request := sign(t, tc.alg, tc.key, map[string]any{
				"iss": "software-1", "aud": tc.aud, "exp": 5, "jti": "request-1",
				"grant_types": []string{"client_credentials"}, "software_statement": "a.b.c",
				"id_token_encrypted_response_alg": "RSA-OAEP",
			})

			claims, err := f.verifier.VerifyRequest(t.Context(), request, f.recipientJWKS)

			require.NoError(t, err)
			assert.Equal(t, "software-1", claims.Issuer)
			assert.Contains(t, claims.Audience, "https://id.bank.example")
			assert.Equal(t, []string{"client_credentials"}, claims.GrantTypes)
			assert.Equal(t, "a.b.c", claims.SoftwareStatement)
		})
	}
}

// Verifies: DH-0622
func TestVerifyRequestRefusesARequestThatTheSoftwareProductDidNotSign(t *testing.T) {
	f := newFixture(t)
	claims := map[string]any{"iss": "software-1"}

	tests := map[string]string{
		"a key that is not at the jwks_uri":     sign(t, jose.PS256, rsaKey(t, "unknown"), claims),
		"a different key with the same key ID":  sign(t, jose.PS256, rsaKey(t, "recipient-1"), claims),
		"an algorithm that is not permitted":    sign(t, jose.RS256, f.recipientKey, claims),
		"a claim with a type that is not valid": sign(t, jose.PS256, f.recipientKey, map[string]any{"exp": "soon"}),
	}
	for name, request := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := f.verifier.VerifyRequest(t.Context(), request, f.recipientJWKS)

			assert.ErrorIs(t, err, app.ErrInvalidJWT)
		})
	}
}

func TestVerifyRequestGivesKeysNotAvailableWhenItCannotGetTheJWKS(t *testing.T) {
	f := newFixture(t)
	request := sign(t, jose.PS256, f.recipientKey, map[string]any{})

	t.Run("an error status", func(t *testing.T) {
		f.recipient.status.Store(http.StatusNotFound)

		_, err := f.verifier.VerifyRequest(t.Context(), request, f.recipientJWKS)

		assert.ErrorIs(t, err, app.ErrKeysNotAvailable)
	})

	t.Run("no server", func(t *testing.T) {
		f.recipient.Close()

		_, err := f.verifier.VerifyRequest(t.Context(), request, f.recipientJWKS)

		assert.ErrorIs(t, err, app.ErrKeysNotAvailable)
	})
}
