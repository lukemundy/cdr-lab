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

// The data holder sends the notices of removedProduct to the tests, so that a test can see that none comes.
const removedProduct = "4f6a8c0e-2b4d-4c6e-8f0a-9b1c3d5e7f81"

func TestTheRemovalOfASoftwareProductEndsItsConsentsAndItsRegistrationWithNoNotice(t *testing.T) {
	client := registered(t, removedProduct)
	customer := newEligibleAccount(t)
	tokens := customer.consents(client, noAccountScope, ninetyDays).tokens(t)
	accessToken := accessTokenOf(t, tokens)
	require.IsType(t, "", tokens["refresh_token"], tokens)
	refreshToken := tokens["refresh_token"].(string)
	arrangementID := arrangementIDOf(t, tokens)
	// The Register gives an SSA only to an ACTIVE software product, so the SSA of the new registration comes first.
	again := newRegistrant(t, removedProduct)

	changeStatus(t, removedProduct, "remove")

	require.Eventually(t, func() bool {
		return arrangementOf(t, customer, arrangementID).EndReason == "SOFTWARE_PRODUCT_REMOVED"
	}, waitForStatus, time.Second, "the data holder ends the arrangement within the interval")

	t.Run("the arrangement ends with no notice", func(t *testing.T) {
		got := arrangementOf(t, customer, arrangementID)

		require.NotNil(t, got.EndedAt)
		assert.Empty(t, got.NoticeStatus, "the end needs no notice")
		assert.Empty(t, labtest.NoticesFor(arrangementID))
	})

	t.Run("the tokens give 401 invalid_token", func(t *testing.T) {
		requireRefusedToken(
			t,
			userInfo(t, client, client.cert, http.MethodGet, accessToken),
			`Bearer error="invalid_token"`,
		)
	})

	t.Run("the client authentication fails", func(t *testing.T) {
		requireTokenError(t, client.requestToken(t, client.assertion(t, tokenEndpoint(t, client))),
			http.StatusUnauthorized, "invalid_client")
		requireTokenError(t, client.refresh(t, refreshToken), http.StatusUnauthorized, "invalid_client")
	})

	t.Run("a new registration with an SSA that is not expired is refused", func(t *testing.T) {
		requireRegistrationError(t, again.register(t, again.claims()), "unapproved_software_statement")
	})
}
