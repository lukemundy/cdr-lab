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

func arrangementRevocationEndpoint(t *testing.T, client registeredClient) string {
	t.Helper()

	endpoint := openIDConfiguration(t, client.issuer)["cdr_arrangement_revocation_endpoint"]
	require.IsType(t, "", endpoint)

	return endpoint.(string)
}

func (c registeredClient) revokeArrangement(t *testing.T, form url.Values) labtest.TokenReply {
	t.Helper()

	return c.tokenRequest(t, arrangementRevocationEndpoint(t, c), form)
}

func (c registeredClient) revokeArrangementID(t *testing.T, arrangementID string) labtest.TokenReply {
	t.Helper()

	return c.revokeArrangement(t, url.Values{"cdr_arrangement_id": {arrangementID}})
}

// requireArrangementError checks the error structure of the CDS Payload Conventions.
func requireArrangementError(t *testing.T, r labtest.TokenReply, status int, code, detail string) {
	t.Helper()
	require.Equal(t, status, r.Status, r.Body)

	want := map[string]any{"code": code, "title": errorTitles[code], "detail": detail}
	assert.Equal(t, map[string]any{"errors": []any{want}}, r.Body)
}

func requireInvalidArrangement(t *testing.T, r labtest.TokenReply, arrangementID string) {
	t.Helper()
	requireArrangementError(t, r, http.StatusUnprocessableEntity,
		"urn:au-cds:error:cds-all:Authorisation/InvalidArrangement", arrangementID)
}

func arrangementIDOf(t *testing.T, tokens map[string]any) string {
	t.Helper()
	require.IsType(t, "", tokens["cdr_arrangement_id"], tokens)

	return tokens["cdr_arrangement_id"].(string)
}

// This test is the full sequence of a revocation: a different client cannot end the consent, the software product
// ends it, each token of the consent stops, and a second revocation is refused.
func TestASoftwareProductEndsAConsent(t *testing.T) {
	client := registered(t, allScopesProduct)
	other := registered(t, registrationTestProduct)
	customer := newAccount(t)
	own := customer.addAccount("Everyday", "70000002")

	tokens := customer.consents(client, "openid bank:accounts.basic:read", ninetyDays, own.AccountID).tokens(t)
	accessToken := accessTokenOf(t, tokens)
	require.IsType(t, "", tokens["refresh_token"], tokens)
	refreshToken := tokens["refresh_token"].(string)
	arrangementID := arrangementIDOf(t, tokens)

	t.Run("the OpenID Provider configuration gives the endpoint on the secure base URI", func(t *testing.T) {
		assert.Equal(t, secureBaseURI(t)+"/arrangements/revoke", arrangementRevocationEndpoint(t, client))
	})

	t.Run("an arrangement of a different client gives InvalidArrangement and does not change", func(t *testing.T) {
		requireInvalidArrangement(t, other.revokeArrangementID(t, arrangementID), arrangementID)

		assert.Equal(t, http.StatusOK, client.accountsStatus(t, accessToken).StatusCode)
	})

	t.Run("a revocation gives 204 with no body", func(t *testing.T) {
		r := client.revokeArrangementID(t, arrangementID)

		require.Equal(t, http.StatusNoContent, r.Status, r.Body)
		assert.Empty(t, r.Body)
	})

	t.Run("after the response, each token of the consent stops", func(t *testing.T) {
		requireRefusedToken(t, client.accountsStatus(t, accessToken), `Bearer error="invalid_token"`)
		requireRefusedToken(t, userInfo(t, client, client.cert, http.MethodGet, accessToken),
			`Bearer error="invalid_token"`)
		requireTokenError(t, client.refresh(t, refreshToken), http.StatusBadRequest, "invalid_grant")
		requireNotActive(t, client.introspect(t, refreshToken))
	})

	t.Run("a second revocation gives InvalidArrangement", func(t *testing.T) {
		requireInvalidArrangement(t, client.revokeArrangementID(t, arrangementID), arrangementID)
	})
}

func TestARevocationOfAnArrangementThatDoesNotExistGivesInvalidArrangement(t *testing.T) {
	client := registered(t, allScopesProduct)
	unknown := rand.Text()

	requireInvalidArrangement(t, client.revokeArrangementID(t, unknown), unknown)
}

// A once-off arrangement is current until its access token expires, so the software product can revoke it, and the
// revocation stops its access token.
//
// Verifies: DH-0911, AUTH-0208
func TestASoftwareProductEndsAOnceOffArrangement(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newEligibleAccount(t)

	tokens := customer.consents(client, noAccountScope, 0).tokens(t)
	accessToken := accessTokenOf(t, tokens)
	arrangementID := arrangementIDOf(t, tokens)

	t.Run("the access token expires in 5 minutes at most", func(t *testing.T) {
		require.IsType(t, float64(0), tokens["expires_in"], tokens)
		assert.LessOrEqual(t, tokens["expires_in"], float64(300))
		assert.Equal(t, http.StatusOK, userInfo(t, client, client.cert, http.MethodGet, accessToken).StatusCode)
	})

	t.Run("a revocation gives 204 and the access token stops", func(t *testing.T) {
		r := client.revokeArrangementID(t, arrangementID)

		require.Equal(t, http.StatusNoContent, r.Status, r.Body)
		requireRefusedToken(t, userInfo(t, client, client.cert, http.MethodGet, accessToken),
			`Bearer error="invalid_token"`)
	})

	t.Run("a second revocation gives InvalidArrangement", func(t *testing.T) {
		requireInvalidArrangement(t, client.revokeArrangementID(t, arrangementID), arrangementID)
	})
}

func TestARevocationThatIsNotValidIsRefused(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newEligibleAccount(t)
	tokens := customer.consents(client, noAccountScope, ninetyDays).tokens(t)
	arrangementID := arrangementIDOf(t, tokens)

	for name, form := range map[string]url.Values{
		"no cdr_arrangement_id":      {},
		"empty cdr_arrangement_id":   {"cdr_arrangement_id": {""}},
		"only a cdr_arrangement_jwt": {"cdr_arrangement_jwt": {"eyJhbGciOiJQUzI1NiJ9.e30.c2ln"}},
	} {
		t.Run(name, func(t *testing.T) {
			requireArrangementError(t, client.revokeArrangement(t, form), http.StatusBadRequest,
				"urn:au-cds:error:cds-all:Field/Missing", "cdr_arrangement_id")
		})
	}

	t.Run("client assertion that is not valid", func(t *testing.T) {
		form := url.Values{"cdr_arrangement_id": {arrangementID}, "client_assertion": {"not-a-jwt"}}

		requireTokenError(t, client.revokeArrangement(t, form), http.StatusUnauthorized, "invalid_client")
	})

	assert.Equal(t, true, client.introspect(t, tokens["refresh_token"].(string)).Body["active"],
		"no refused request ends the consent")
}

func TestAnAccessTokenDoesNotExpireAfterTheEndOfTheSharing(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newEligibleAccount(t)

	tokens := customer.consents(client, noAccountScope, 120).tokens(t)

	require.IsType(t, float64(0), tokens["expires_in"], tokens)
	assert.LessOrEqual(t, tokens["expires_in"].(float64), float64(120))

	info := client.introspect(t, tokens["refresh_token"].(string))
	require.IsType(t, float64(0), info.Body["exp"], info.Body)
	assert.LessOrEqual(t, client.signedClaims(t, accessTokenOf(t, tokens))["exp"], info.Body["exp"])
}
