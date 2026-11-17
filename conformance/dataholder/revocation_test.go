//go:build conformance

package dataholder

import (
	"crypto/rand"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

func revocationEndpoint(t *testing.T, client registeredClient) string {
	t.Helper()

	endpoint := openIDConfiguration(t, client.issuer)["revocation_endpoint"]
	require.IsType(t, "", endpoint)

	return endpoint.(string)
}

func (c registeredClient) revoke(t *testing.T, form url.Values) labtest.TokenReply {
	t.Helper()

	return c.tokenRequest(t, revocationEndpoint(t, c), form)
}

func (c registeredClient) revokeToken(t *testing.T, token string) labtest.TokenReply {
	t.Helper()

	return c.revoke(t, url.Values{"token": {token}})
}

// requireRevocationOK checks the response of RFC 7009, section 2.2: 200 with no body.
func requireRevocationOK(t *testing.T, r labtest.TokenReply) {
	t.Helper()
	require.Equal(t, http.StatusOK, r.Status, r.Body)
	assert.Empty(t, r.Body)
}

func (c registeredClient) accountsStatus(t *testing.T, token string) response {
	t.Helper()

	return bankingResource(t, c.cert, bankingAccountsPath, "3", token)
}

func accessTokenOf(t *testing.T, tokens map[string]any) string {
	t.Helper()
	require.IsType(t, "", tokens["access_token"], tokens)

	return tokens["access_token"].(string)
}

// This test is the full sequence of the revocations of one consent: a different client cannot revoke its tokens, the
// software product revokes access tokens and keeps the consent, and then it revokes the refresh token.
func TestASoftwareProductRevokesItsTokens(t *testing.T) {
	client := registered(t, allScopesProduct)
	other := registered(t, registrationTestProduct)
	customer := newAccount(t)
	own := customer.addAccount("Everyday", "70000001")

	tokens := customer.consents(client, "openid bank:accounts.basic:read", ninetyDays, own.AccountID).tokens(t)
	accessToken := accessTokenOf(t, tokens)
	require.IsType(t, "", tokens["refresh_token"], tokens)
	refreshToken := tokens["refresh_token"].(string)

	t.Run("the OpenID Provider configuration gives the revocation endpoint on the secure base URI", func(t *testing.T) {
		assert.Equal(t, secureBaseURI(t)+"/revoke", revocationEndpoint(t, client))
	})

	t.Run("a token of a different client gives unauthorized_client and does not change", func(t *testing.T) {
		for _, token := range []string{accessToken, refreshToken} {
			requireTokenError(t, other.revokeToken(t, token), http.StatusBadRequest, "unauthorized_client")
		}

		assert.Equal(t, http.StatusOK, client.accountsStatus(t, accessToken).StatusCode)
		assert.Equal(t, true, client.introspect(t, refreshToken).Body["active"])
	})

	t.Run(
		"a revoked access token is refused, and a new access token from the refresh token is valid",
		func(t *testing.T) {
			requireRevocationOK(t, client.revokeToken(t, accessToken))

			requireRefusedToken(t, client.accountsStatus(t, accessToken), `Bearer error="invalid_token"`)

			refreshed := client.refresh(t, refreshToken)
			require.Equal(t, http.StatusOK, refreshed.Status, refreshed.Body)
			assert.Equal(t, http.StatusOK, client.accountsStatus(t, accessTokenOf(t, refreshed.Body)).StatusCode)
		},
	)

	t.Run("the token type hint has no effect", func(t *testing.T) {
		refreshed := client.refresh(t, refreshToken)
		require.Equal(t, http.StatusOK, refreshed.Status, refreshed.Body)
		hinted := accessTokenOf(t, refreshed.Body)

		requireRevocationOK(t, client.revoke(t, url.Values{"token": {hinted}, "token_type_hint": {"refresh_token"}}))

		requireRefusedToken(t, client.accountsStatus(t, hinted), `Bearer error="invalid_token"`)
		assert.Equal(t, true, client.introspect(t, refreshToken).Body["active"])
	})

	t.Run("a value that is not an active token of the client gives 200 and changes nothing", func(t *testing.T) {
		for _, token := range []string{rand.Text(), tokens["id_token"].(string), accessToken} {
			requireRevocationOK(t, client.revokeToken(t, token))
		}

		assert.Equal(t, true, client.introspect(t, refreshToken).Body["active"])
	})

	t.Run("a revoked refresh token ends each token of its grant", func(t *testing.T) {
		refreshed := client.refresh(t, refreshToken)
		require.Equal(t, http.StatusOK, refreshed.Status, refreshed.Body)

		requireRevocationOK(t, client.revokeToken(t, refreshToken))

		requireTokenError(t, client.refresh(t, refreshToken), http.StatusBadRequest, "invalid_grant")
		requireNotActive(t, client.introspect(t, refreshToken))
		requireRefusedToken(
			t,
			client.accountsStatus(t, accessTokenOf(t, refreshed.Body)),
			`Bearer error="invalid_token"`,
		)
	})

	// Verifies: AUTH-0223, DH-0910
	t.Run("the consent of more than 24 hours does not end", func(t *testing.T) {
		assert.Nil(t, arrangementOf(t, customer, arrangementIDOf(t, tokens)).EndedAt)
	})
}

// Verifies: DH-0910, AUTH-0223
func TestARevokedRefreshTokenOfAOneTimeCollectionEndsItsConsent(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newEligibleAccount(t)
	tokens := customer.consents(client, noAccountScope, oneHour).tokens(t)
	require.IsType(t, "", tokens["refresh_token"], tokens)
	arrangementID := arrangementIDOf(t, tokens)

	requireRevocationOK(t, client.revokeToken(t, tokens["refresh_token"].(string)))

	ended := arrangementOf(t, customer, arrangementID)
	assert.NotNil(t, ended.EndedAt)
	assert.Equal(t, "REVOKED_BY_DATA_RECIPIENT", ended.EndReason)
	assert.Empty(t, ended.NoticeStatus, "the consent has no notice")
	requireInvalidArrangement(t, client.revokeArrangementID(t, arrangementID), arrangementID)
}

func TestARevokedClientCredentialsTokenIsRefusedByRegistrationManagement(t *testing.T) {
	client := registered(t, allScopesProduct)
	other := registered(t, registrationTestProduct)
	token := client.accessToken(t)

	requireTokenError(t, other.revokeToken(t, token), http.StatusBadRequest, "unauthorized_client")
	require.Equal(t, http.StatusOK, client.readRegistration(t, client.cert, client.clientID, token).status)

	requireRevocationOK(t, client.revokeToken(t, token))

	read := client.readRegistration(t, client.cert, client.clientID, token)
	assert.Equal(t, http.StatusUnauthorized, read.status, string(read.body))
	assert.Equal(t, `Bearer error="invalid_token"`, read.header.Get("WWW-Authenticate"))
}

func TestTokenRevocationRefusesARequestThatIsNotValid(t *testing.T) {
	client := registered(t, allScopesProduct)

	t.Run("no token", func(t *testing.T) {
		requireTokenError(t, client.revoke(t, url.Values{}), http.StatusBadRequest, "invalid_request")
	})

	t.Run("client assertion that is not valid", func(t *testing.T) {
		form := url.Values{"token": {rand.Text()}, "client_assertion": {"not-a-jwt"}}

		requireTokenError(t, client.revoke(t, form), http.StatusUnauthorized, "invalid_client")
	})
}
