//go:build conformance

package dataholder

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

const inactiveProduct = "2d4f6b8a-0c1e-4a3b-9d5f-7b9c1d3e5f70"

// waitForStatus is the status interval of the data holder in the lab configuration, plus 2 s for the request to the
// Register and the change.
func waitForStatus(t *testing.T) time.Duration {
	t.Helper()

	interval, err := time.ParseDuration(labtest.Env(t, "DATAHOLDER_STATUS_INTERVAL"))
	require.NoError(t, err, "DATAHOLDER_STATUS_INTERVAL is not a duration")

	return interval + 2*time.Second
}

func changeStatus(t *testing.T, productID, action string) {
	t.Helper()

	labtest.ChangeSoftwareProductStatus(t, registrationTestRecipient, productID, action)
}

func customerStatus(t *testing.T, client registeredClient, accessToken string) int {
	t.Helper()

	return getCustomer(t, client.cert, resourceHeader(accessToken)).StatusCode
}

func TestASoftwareProductThatIsNotActiveGetsNoConsentAndNoDataUntilItIsActiveAgain(t *testing.T) {
	client := registered(t, inactiveProduct)
	customer := newEligibleAccount(t)
	tokens := customer.consents(client, customerScope, ninetyDays).tokens(t)
	accessToken := accessTokenOf(t, tokens)
	require.IsType(t, "", tokens["refresh_token"], tokens)
	refreshToken := tokens["refresh_token"].(string)
	notExchanged := customer.consents(client, noAccountScope, ninetyDays)
	spare := customer.consents(client, noAccountScope, ninetyDays).tokens(t)
	flow := client.authorisationURL(t, client.claimsFor(noAccountScope, ninetyDays))
	require.Equal(t, http.StatusOK, customerStatus(t, client, accessToken))

	changeStatus(t, inactiveProduct, "deactivate")

	require.Eventually(t, func() bool { return customerStatus(t, client, accessToken) == http.StatusForbidden },
		waitForStatus(t), time.Second, "the data holder refuses the software product within the interval")

	t.Run("the resource APIs give 403 AdrStatusNotActive with the status", func(t *testing.T) {
		resp := getCustomer(t, client.cert, resourceHeader(accessToken))

		assert.Equal(t, "INACTIVE", requireCDSError(t, resp, http.StatusForbidden,
			"urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive", "ADR Status Is Not Active"))
	})

	t.Run("UserInfo gives 401 invalid_token", func(t *testing.T) {
		resp := userInfo(t, client, client.cert, http.MethodGet, accessToken)

		requireRefusedToken(t, resp, `Bearer error="invalid_token"`)
	})

	t.Run("a token that is not valid gets its own error before the status error", func(t *testing.T) {
		requireRefusedToken(t, client.accountsStatus(t, "not-a-token"), `Bearer error="invalid_token"`)
	})

	t.Run("PAR gives 400 unauthorized_client", func(t *testing.T) {
		requirePushRefused(
			t,
			client.push(t, client.claimsFor(noAccountScope, ninetyDays)),
			http.StatusBadRequest,
			"unauthorized_client",
		)
	})

	t.Run("a flow that started before the change ends with unauthorized_client", func(t *testing.T) {
		claims := client.responseClaims(t, newConsumer(t).open(flow))

		assert.Equal(t, "unauthorized_client", claims["error"])
		assert.NotContains(t, claims, "code")
	})

	t.Run("the authorization_code and refresh_token grants give 400 unauthorized_client", func(t *testing.T) {
		requireTokenError(t, notExchanged.exchange(t), http.StatusBadRequest, "unauthorized_client")
		requireTokenError(t, client.refresh(t, refreshToken), http.StatusBadRequest, "unauthorized_client")
	})

	t.Run("the client credentials grant and the other endpoints work", func(t *testing.T) {
		registrationToken := client.accessToken(t)
		read := client.readRegistration(t, client.cert, client.clientID, registrationToken)
		assert.Equal(t, http.StatusOK, read.status, string(read.body))

		introspected := client.introspect(t, refreshToken)
		require.Equal(t, http.StatusOK, introspected.Status, introspected.Body)
		assert.Equal(t, true, introspected.Body["active"], "the refresh token does not change")

		requireRevocationOK(t, client.revokeToken(t, accessTokenOf(t, spare)))
		assert.Equal(t, http.StatusNoContent, client.revokeArrangementID(t, arrangementIDOf(t, spare)).Status)
	})

	changeStatus(t, inactiveProduct, "activate")

	t.Run("an ACTIVE software product reads the data of its consent again with no new consent", func(t *testing.T) {
		require.Eventually(t, func() bool { return customerStatus(t, client, accessToken) == http.StatusOK },
			waitForStatus(t), time.Second)

		refreshed := client.refresh(t, refreshToken)
		require.Equal(t, http.StatusOK, refreshed.Status, refreshed.Body)
	})
}
