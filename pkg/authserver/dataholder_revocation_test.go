package authserver_test

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

type revokedTokenStore struct {
	mu       sync.Mutex
	expiries map[string]time.Time
	failure  error
}

func (s *revokedTokenStore) RevokeToken(_ context.Context, id string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.failure != nil {
		return s.failure
	}

	s.expiries[id] = expiresAt

	return nil
}

func (s *revokedTokenStore) TokenRevoked(_ context.Context, id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.failure != nil {
		return false, s.failure
	}

	_, ok := s.expiries[id]

	return ok, nil
}

// revoke sends the request as the client with the ID clientID.
func (b *bank) revoke(clientID, token string) reply {
	b.t.Helper()

	a := b.assertion()
	a.claims["iss"], a.claims["sub"] = clientID, clientID
	form := b.form(a)
	form.Del("scope")
	form.Del("grant_type")
	form.Set("token", token)

	return b.post(b.revocationURL, form)
}

// signedByAnotherIssuer gives an access token of a different authorisation server for the registered client.
func signedByAnotherIssuer(t *testing.T) string {
	t.Helper()

	key, err := authserver.NewSigningKey()
	require.NoError(t, err)

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.PS256, Key: jose.JSONWebKey{Key: key.Key, KeyID: key.ID}},
		(&jose.SignerOptions{}).WithType("at+jwt"),
	)
	require.NoError(t, err)

	now := time.Now()
	token, err := jwt.Signed(signer).Claims(map[string]any{
		"iss": "https://id.other.example", "client_id": registeredClientID, "jti": rand.Text(),
		"iat": now.Unix(), "exp": now.Add(time.Minute).Unix(),
	}).Serialize()
	require.NoError(t, err)

	return token
}

func requireRevoked(t *testing.T, r reply) {
	t.Helper()
	require.Equal(t, http.StatusOK, r.status, r.body)
	assert.Empty(t, r.body)
}

func (b *bank) consentTokenValid(token string) bool {
	b.t.Helper()

	_, err := b.dataHolder.VerifyConsentToken(b.t.Context(), token, leaf(b.t, b.cert))
	if errors.Is(err, authserver.ErrInvalidToken) {
		return false
	}

	require.NoError(b.t, err)

	return true
}

// Verifies: AUTH-0224
func TestARevokedAccessTokenOfAConsentIsRefusedAndANewAccessTokenFromTheRefreshTokenIsValid(t *testing.T) {
	b := newBank(t)
	tokens := b.exchange(b.authorise(hour))
	accessToken := tokens["access_token"].(string)

	requireRevoked(t, b.revoke(registeredClientID, accessToken))

	assert.False(t, b.consentTokenValid(accessToken))

	r := b.refresh(tokens["refresh_token"].(string))
	require.Equal(t, http.StatusOK, r.status, r.body)
	assert.True(t, b.consentTokenValid(r.body["access_token"].(string)))
}

// Verifies: AUTH-0224
func TestARevokedClientCredentialsTokenIsRefused(t *testing.T) {
	b := newBank(t)
	accessToken := b.accessToken()

	requireRevoked(t, b.revoke(registeredClientID, accessToken))

	_, err := b.dataHolder.VerifyAccessToken(t.Context(), accessToken, leaf(t, b.cert))
	require.ErrorIs(t, err, authserver.ErrInvalidToken)
}

// Verifies: AUTH-0223
func TestARevokedRefreshTokenEndsEachTokenOfItsGrant(t *testing.T) {
	b := newBank(t)
	tokens := b.exchange(b.authorise(hour))
	refreshToken := tokens["refresh_token"].(string)
	refreshed := b.refresh(refreshToken)
	require.Equal(t, http.StatusOK, refreshed.status, refreshed.body)

	requireRevoked(t, b.revoke(registeredClientID, refreshToken))

	requireTokenError(t, b.refresh(refreshToken), "invalid_grant")
	assert.Equal(t, map[string]any{"active": false}, b.introspect(refreshToken).body)
	assert.False(t, b.consentTokenValid(tokens["access_token"].(string)))
	assert.False(t, b.consentTokenValid(refreshed.body["access_token"].(string)))
}

// Verifies: AUTH-0225
func TestATokenOfADifferentClientGivesUnauthorizedClientAndDoesNotChange(t *testing.T) {
	b := newBank(t)
	tokens := b.exchange(b.authorise(hour))
	refreshToken := tokens["refresh_token"].(string)
	clientCredentialsToken := b.accessToken()

	for _, token := range []string{tokens["access_token"].(string), refreshToken, clientCredentialsToken} {
		requireTokenError(t, b.revoke(secondClientID, token), "unauthorized_client")
	}

	assert.True(t, b.consentTokenValid(tokens["access_token"].(string)))
	assert.Equal(t, true, b.introspect(refreshToken).body["active"])

	_, err := b.dataHolder.VerifyAccessToken(t.Context(), clientCredentialsToken, leaf(t, b.cert))
	require.NoError(t, err)
}

// Verifies: AUTH-0226
func TestRevocationOfAValueThatIsNotAnActiveTokenOfTheClientGivesOKAndChangesNothing(t *testing.T) {
	b := newBank(t)
	tokens := b.exchange(b.authorise(hour))
	revokedAccessToken := b.accessToken()
	requireRevoked(t, b.revoke(registeredClientID, revokedAccessToken))
	revoked := len(b.revoked.expiries)

	for name, token := range map[string]string{
		"unknown value":         rand.Text(),
		"ID token":              tokens["id_token"].(string),
		"revoked access token":  revokedAccessToken,
		"token of another bank": signedByAnotherIssuer(t),
	} {
		requireRevoked(t, b.revoke(registeredClientID, token))
		assert.Len(t, b.revoked.expiries, revoked, name)
	}

	assert.True(t, b.consentTokenValid(tokens["access_token"].(string)))
	assert.Equal(t, true, b.introspect(tokens["refresh_token"].(string)).body["active"])
}

// Verifies: AUTH-0222
func TestTheTokenTypeHintHasNoEffect(t *testing.T) {
	b := newBank(t)
	tokens := b.exchange(b.authorise(hour))
	refreshToken := tokens["refresh_token"].(string)

	a := b.assertion()
	form := b.form(a)
	form.Del("scope")
	form.Del("grant_type")
	form.Set("token", tokens["access_token"].(string))
	form.Set("token_type_hint", "refresh_token")
	requireRevoked(t, b.post(b.revocationURL, form))

	assert.False(t, b.consentTokenValid(tokens["access_token"].(string)))
	assert.Equal(t, true, b.introspect(refreshToken).body["active"])
}

// Verifies: AUTH-0032, AUTH-0227
func TestRevocationRefusesARequestThatIsNotValid(t *testing.T) {
	b := newBank(t)

	requireTokenError(t, b.post(b.revocationURL, b.assertionForm()), "invalid_request")

	r := b.post(b.revocationURL, url.Values{"token": {b.accessToken()}})
	assert.Equal(t, http.StatusUnauthorized, r.status, r.body)
	assert.Equal(t, "invalid_client", r.body["error"])
}

func TestARevokedAccessTokenStaysRevokedAfterARestart(t *testing.T) {
	b := newBank(t)
	accessToken := b.exchange(b.authorise(hour))["access_token"].(string)
	requireRevoked(t, b.revoke(registeredClientID, accessToken))

	b.restart()

	assert.False(t, b.consentTokenValid(accessToken))
}

func TestTheStoreKeepsARevokedAccessTokenUntilItExpires(t *testing.T) {
	b := newBank(t)
	accessToken := b.accessToken()
	expiry := time.Unix(int64(b.claims(accessToken)["exp"].(float64)), 0).UTC()

	requireRevoked(t, b.revoke(registeredClientID, accessToken))

	require.Len(t, b.revoked.expiries, 1)

	for id, expiresAt := range b.revoked.expiries {
		assert.Equal(t, b.claims(accessToken)["jti"], id)
		assert.Equal(t, expiry, expiresAt)
	}
}

func TestAStoreFailureGivesAServerErrorAndNoAccess(t *testing.T) {
	b := newBank(t)
	accessToken := b.accessToken()
	b.revoked.failure = errors.New("database failure")

	r := b.revoke(registeredClientID, accessToken)
	assert.Equal(t, http.StatusInternalServerError, r.status, r.body)
	assert.Equal(t, "server_error", r.body["error"])

	_, err := b.dataHolder.VerifyAccessToken(t.Context(), accessToken, leaf(t, b.cert))
	require.Error(t, err)
	assert.NotErrorIs(t, err, authserver.ErrInvalidToken)
}
