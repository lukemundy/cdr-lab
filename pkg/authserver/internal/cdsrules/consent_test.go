package cdsrules_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sharing = cdsClaims(map[string]any{"sharing_duration": 3600})

func requireInvalidGrant(t *testing.T, r reply) {
	t.Helper()
	require.Equal(t, http.StatusBadRequest, r.status, r.body)
	assert.Equal(t, "invalid_grant", r.body["error"])
}

// Verifies: AUTH-0205
func TestTheTokenResponsesHaveTheArrangementID(t *testing.T) {
	_, recipient := newAuthServer(t)

	tokens := recipient.exchange(recipient.authorise(sharing))
	refreshed := recipient.refreshAs(clientID, tokens["refresh_token"].(string))

	assert.Equal(t, arrangementID, tokens["cdr_arrangement_id"])
	require.Equal(t, http.StatusOK, refreshed.status, refreshed.body)
	assert.Equal(t, arrangementID, refreshed.body["cdr_arrangement_id"])
	assert.NotEmpty(t, refreshed.body["access_token"])
	assert.NotEqual(t, tokens["access_token"], refreshed.body["access_token"])
}

func TestAClientCredentialsResponseHasNoArrangementID(t *testing.T) {
	_, recipient := newAuthServer(t)
	form := recipient.clientAssertion(clientID)
	form.Set("grant_type", "client_credentials")
	form.Set("scope", "cdr:registration")

	resp := recipient.post("/token", form)

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.NotContains(t, resp.body, "cdr_arrangement_id")
}

// Verifies: AUTH-0210, AUTH-0215, AUTH-0216
func TestTheIDTokenHasTheSubjectAndTheAuthenticationOfTheConsent(t *testing.T) {
	server, recipient := newAuthServer(t)

	tokens := recipient.exchange(recipient.authorise(sharing))
	refreshed := recipient.refreshAs(clientID, tokens["refresh_token"].(string))

	for name, idToken := range map[string]any{"exchange": tokens["id_token"], "refresh": refreshed.body["id_token"]} {
		t.Run(name, func(t *testing.T) {
			claims := server.claims(t, idToken.(string))

			assert.Equal(t, pairwiseSub, claims["sub"])
			assert.Equal(t, "urn:cds.au:cdr:2", claims["acr"])
			assert.EqualValues(t, server.authenticatedAt.Unix(), claims["auth_time"])
			assert.Equal(t, "nonce-1", claims["nonce"])
		})
	}
}

// Verifies: AUTH-0209
func TestTheRefreshTokenExpiresWhenTheSharingOfTheConsentEnds(t *testing.T) {
	server, recipient := newAuthServer(t)

	tokens := recipient.exchange(recipient.authorise(sharing))

	grant, err := server.store.GrantByRefreshToken(t.Context(), tokens["refresh_token"].(string))
	require.NoError(t, err)
	assert.Equal(t, grant.Store["sharing_ends_at"], grant.RefreshTokenExpiresAt)
}

// Verifies: AUTH-0211
func TestARefreshDoesNotChangeTheRefreshToken(t *testing.T) {
	_, recipient := newAuthServer(t)
	refreshToken := recipient.exchange(recipient.authorise(sharing))["refresh_token"].(string)

	first := recipient.refreshAs(clientID, refreshToken)
	second := recipient.refreshAs(clientID, refreshToken)

	require.Equal(t, http.StatusOK, first.status, first.body)
	require.Equal(t, http.StatusOK, second.status, second.body)

	if issued, ok := first.body["refresh_token"]; ok {
		assert.Equal(t, refreshToken, issued)
	}
}

// Verifies: AUTH-0202, AUTH-0203
func TestEachLaterUseOfACodeGivesInvalidGrantAndEndsItsRefreshToken(t *testing.T) {
	_, recipient := newAuthServer(t)
	a := recipient.authorise(sharing)
	tokens := recipient.exchange(a)

	requireInvalidGrant(t, recipient.exchangeAs(clientID, a))
	requireInvalidGrant(t, recipient.exchangeAs(clientID, a))
	requireInvalidGrant(t, recipient.refreshAs(clientID, tokens["refresh_token"].(string)))
}

// Verifies: AUTH-0204
func TestAnExchangeThatFailsEndsTheCode(t *testing.T) {
	_, recipient := newAuthServer(t)
	a := recipient.authorise(sharing)
	wrong := a
	wrong.verifier = "a-verifier-that-does-not-match-the-code-challenge"

	requireInvalidGrant(t, recipient.exchangeAs(clientID, wrong))
	requireInvalidGrant(t, recipient.exchangeAs(clientID, a))
}

// Verifies: AUTH-0202
func TestACodeOfADifferentClientGivesInvalidGrant(t *testing.T) {
	_, recipient := newAuthServer(t)

	requireInvalidGrant(t, recipient.exchangeAs(otherClientID, recipient.authorise(sharing)))
}

// Verifies: AUTH-0202
func TestAnExpiredCodeGivesInvalidGrant(t *testing.T) {
	server, recipient := newAuthServer(t)
	a := recipient.authorise(sharing)
	grant, err := server.store.GrantByAuthCode(t.Context(), a.jarm["code"].(string))
	require.NoError(t, err)

	grant.AuthCodeExpiresAt = int(time.Now().Add(-time.Second).Unix())
	require.NoError(t, server.store.SaveGrant(t.Context(), grant))

	requireInvalidGrant(t, recipient.exchangeAs(clientID, a))
}

// Verifies: AUTH-0212
func TestARefreshTokenThatTheClientCannotUseGivesInvalidGrant(t *testing.T) {
	server, recipient := newAuthServer(t)
	refreshToken := recipient.exchange(recipient.authorise(sharing))["refresh_token"].(string)

	requireInvalidGrant(t, recipient.refreshAs(otherClientID, refreshToken))
	requireInvalidGrant(t, recipient.refreshAs(clientID, "a-refresh-token-that-does-not-exist"))

	grant, err := server.store.GrantByRefreshToken(t.Context(), refreshToken)
	require.NoError(t, err)

	grant.RefreshTokenExpiresAt = int(time.Now().Add(-time.Second).Unix())
	require.NoError(t, server.store.SaveGrant(t.Context(), grant))

	requireInvalidGrant(t, recipient.refreshAs(clientID, refreshToken))
}

// Verifies: AUTH-0213
func TestARefreshWithANewCertificateGivesATokenThatIsBoundToIt(t *testing.T) {
	server, recipient := newAuthServer(t)
	tokens := recipient.exchange(recipient.authorise(sharing))
	renewed := recipient.withNewCertificate()

	refreshed := renewed.refreshAs(clientID, tokens["refresh_token"].(string))

	require.Equal(t, http.StatusOK, refreshed.status, refreshed.body)

	before := server.claims(t, tokens["access_token"].(string))["cnf"]
	after := server.claims(t, refreshed.body["access_token"].(string))["cnf"]
	require.NotEmpty(t, after)
	assert.NotEqual(t, before, after)

	again := recipient.refreshAs(clientID, tokens["refresh_token"].(string))
	require.Equal(t, http.StatusOK, again.status, again.body)
	assert.Equal(t, before, server.claims(t, again.body["access_token"].(string))["cnf"])
}

// Verifies: AUTH-0219
func TestIntrospectionOfARefreshTokenOfTheClientGivesTheArrangement(t *testing.T) {
	server, recipient := newAuthServer(t)
	refreshToken := recipient.exchange(recipient.authorise(sharing))["refresh_token"].(string)
	grant, err := server.store.GrantByRefreshToken(t.Context(), refreshToken)
	require.NoError(t, err)

	resp := recipient.introspectAs(clientID, refreshToken)

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.Equal(t, map[string]any{
		"active":             true,
		"exp":                float64(grant.RefreshTokenExpiresAt),
		"scope":              "openid " + dataScope,
		"cdr_arrangement_id": arrangementID,
	}, resp.body)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))
}

// Verifies: AUTH-0220
func TestIntrospectionOfEachOtherTokenGivesOnlyNotActive(t *testing.T) {
	server, recipient := newAuthServer(t)
	tokens := recipient.exchange(recipient.authorise(sharing))
	refreshToken := tokens["refresh_token"].(string)

	expired := recipient.exchange(recipient.authorise(sharing))["refresh_token"].(string)
	grant, err := server.store.GrantByRefreshToken(t.Context(), expired)
	require.NoError(t, err)

	grant.RefreshTokenExpiresAt = int(time.Now().Add(-time.Second).Unix())
	require.NoError(t, server.store.SaveGrant(t.Context(), grant))

	tests := map[string]struct{ client, token string }{
		"access token":                        {clientID, tokens["access_token"].(string)},
		"ID token":                            {clientID, tokens["id_token"].(string)},
		"refresh token of a different client": {otherClientID, refreshToken},
		"expired refresh token":               {clientID, expired},
		"unknown value":                       {clientID, "a-value-that-is-not-a-token"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := recipient.introspectAs(tt.client, tt.token)

			require.Equal(t, http.StatusOK, resp.status, resp.body)
			assert.Equal(t, map[string]any{"active": false}, resp.body)
		})
	}
}

// Verifies: AUTH-0221
func TestIntrospectionRefusesARequestThatIsNotValid(t *testing.T) {
	_, recipient := newAuthServer(t)
	refreshToken := recipient.exchange(recipient.authorise(sharing))["refresh_token"].(string)

	noToken := recipient.post("/introspect", recipient.clientAssertion(clientID))
	form := recipient.clientAssertion(clientID)
	form.Set("token", refreshToken)
	form.Set("client_assertion", "not-a-jwt")
	noClient := recipient.post("/introspect", form)

	assert.Equal(t, http.StatusBadRequest, noToken.status, noToken.body)
	assert.Equal(t, "invalid_request", noToken.body["error"])
	assert.Equal(t, http.StatusUnauthorized, noClient.status, noClient.body)
	assert.Equal(t, "invalid_client", noClient.body["error"])
}
