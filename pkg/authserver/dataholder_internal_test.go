package authserver

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type oneClient struct{}

type allActive struct{}

func (allActive) Active(context.Context, string) (bool, error) { return true, nil }

func (oneClient) Client(context.Context, string) (DataHolderClient, error) {
	return DataHolderClient{ID: "client-1"}, nil
}

type noGrantStore struct{}

func (noGrantStore) SaveGrant(context.Context, Grant) error { return nil }

func (noGrantStore) Grant(context.Context, string) (Grant, error) { return Grant{}, ErrGrantNotFound }

func (noGrantStore) GrantByCode(context.Context, string) (Grant, error) {
	return Grant{}, ErrGrantNotFound
}

func (noGrantStore) GrantByRefreshToken(context.Context, string) (Grant, error) {
	return Grant{}, ErrGrantNotFound
}

type noJTIs struct{}

func (noJTIs) Consume(context.Context, string) error { return nil }

type noArrangements struct{}

func (noArrangements) RevokeArrangement(context.Context, string, string) error {
	return ErrInvalidArrangement
}

type noRevokedTokens struct{}

func (noRevokedTokens) RevokeToken(context.Context, string, time.Time) error { return nil }

func (noRevokedTokens) TokenRevoked(context.Context, string) (bool, error) { return false, nil }

// This test is in the package because only the package can set the clock of a DataHolder.
func TestVerifyAccessTokenRefusesATokenAfterItsExpiry(t *testing.T) {
	key, err := NewSigningKey()
	require.NoError(t, err)

	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Recipient App Web"}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.Key.PublicKey, key.Key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	dataHolder, err := NewDataHolder(DataHolderConfig{
		Issuer:         "https://id.bank.example",
		SecureBaseURL:  "https://mtls.bank.example",
		Scopes:         []string{RegistrationScope},
		Clients:        oneClient{},
		JTIs:           noJTIs{},
		SigningKeys:    func(context.Context) ([]SigningKey, error) { return []SigningKey{key}, nil },
		Pages:          &silentPages{},
		Grants:         noGrantStore{},
		RevokedTokens:  noRevokedTokens{},
		Arrangements:   noArrangements{},
		ClientStatuses: allActive{},
	})
	require.NoError(t, err)

	issued := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.PS256, Key: jose.JSONWebKey{Key: key.Key, KeyID: key.ID}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	require.NoError(t, err)
	token, err := jwt.Signed(signer).Claims(map[string]any{
		"iss": "https://id.bank.example", "client_id": "client-1", "jti": "token-1", "scope": RegistrationScope,
		"exp": issued.Add(AccessTokenLifetime).Unix(),
		"cnf": map[string]string{"x5t#S256": thumbprint(cert)},
	}).Serialize()
	require.NoError(t, err)

	dataHolder.now = func() time.Time { return issued.Add(4 * time.Minute) }
	_, err = dataHolder.VerifyAccessToken(t.Context(), token, cert)
	require.NoError(t, err, "the token is valid before its expiry")

	dataHolder.now = func() time.Time { return issued.Add(5*time.Minute + time.Second) }
	_, err = dataHolder.VerifyAccessToken(t.Context(), token, cert)

	assert.ErrorIs(t, err, ErrInvalidToken)
}

type silentPages struct{}

func (*silentPages) Interact(http.ResponseWriter, *http.Request, Interaction) Outcome {
	return Continue(nil)
}

func (*silentPages) Refused(w http.ResponseWriter, _ *http.Request, status int) {
	w.WriteHeader(status)
}

type consentClient struct{}

func (consentClient) Client(context.Context, string) (DataHolderClient, error) {
	return DataHolderClient{
		ID:           "client-1",
		Scopes:       []string{"openid"},
		GrantTypes:   []string{"authorization_code"},
		RedirectURIs: []string{"https://recipient.example/callback"},
	}, nil
}

// authoriseAt gives the status of a request to the authorisation endpoint at a time after the session started. It is
// in the package because only the package can set the clock of a DataHolder and add a session.
func authoriseAt(t *testing.T, session goidc.AuthnSession, lifetime, after time.Duration, target string) int {
	t.Helper()

	key, err := NewSigningKey()
	require.NoError(t, err)

	dataHolder, err := NewDataHolder(DataHolderConfig{
		Issuer:         "https://id.bank.example",
		SecureBaseURL:  "https://mtls.bank.example",
		Scopes:         []string{"openid"},
		Clients:        consentClient{},
		JTIs:           noJTIs{},
		SigningKeys:    func(context.Context) ([]SigningKey, error) { return []SigningKey{key}, nil },
		Pages:          &silentPages{},
		Grants:         noGrantStore{},
		RevokedTokens:  noRevokedTokens{},
		Arrangements:   noArrangements{},
		ClientStatuses: allActive{},
	})
	require.NoError(t, err)

	start := time.Now()
	session.ClientID = "client-1"
	session.Status = goidc.StatusPending
	session.AuthorizationParameters = goidc.AuthorizationParameters{
		RedirectURI:         "https://recipient.example/callback",
		ResponseType:        goidc.ResponseTypeCode,
		ResponseMode:        goidc.ResponseModeJWT,
		Scopes:              "openid",
		Nonce:               "nonce-1",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: goidc.CodeChallengeMethodSHA256,
	}
	session.ExpiresAt = int(start.Add(lifetime).Unix())
	require.NoError(t, dataHolder.sessions.SaveSession(t.Context(), &session))

	mux := http.NewServeMux()
	dataHolder.RegisterAuthorisation(mux)

	dataHolder.now = func() time.Time { return start.Add(after) }
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))

	return rec.Code
}

// Verifies: AUTH-0066
func TestARequestURIIsValidFor60Seconds(t *testing.T) {
	const (
		lifetime = dataHolderRequestURILifetimeSeconds * time.Second
		target   = "/authorize?client_id=client-1&request_uri=urn:ietf:params:oauth:request_uri:pushed-1"
	)

	session := goidc.AuthnSession{ID: "session-1", PushedAuthReqID: "pushed-1"}

	assert.Equal(t, 60*time.Second, lifetime)
	assert.Equal(t, http.StatusOK, authoriseAt(t, session, lifetime, lifetime-time.Second, target))
	assert.Equal(t, http.StatusBadRequest, authoriseAt(t, session, lifetime, lifetime, target))
}

// Verifies: AUTH-0103
func TestAFlowIsValidFor10Minutes(t *testing.T) {
	const (
		lifetime = dataHolderFlowLifetimeSeconds * time.Second
		target   = "/authorize/flow-1"
	)

	session := goidc.AuthnSession{ID: "flow-1", PolicyID: interactionPolicyID}

	assert.Equal(t, 10*time.Minute, lifetime)
	assert.Equal(t, http.StatusOK, authoriseAt(t, session, lifetime, lifetime-time.Second, target))
	assert.Equal(t, http.StatusBadRequest, authoriseAt(t, session, lifetime, lifetime, target))
}
