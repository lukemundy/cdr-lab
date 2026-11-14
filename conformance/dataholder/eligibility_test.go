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

// secondNoticeProduct is a second software product whose recipient base URI is the JWKS server of the tests.
const secondNoticeProduct = "7a9c1e3b-5d7f-4a9b-8c1d-3e5f7a9b1c2d"

type consentTokens struct {
	arrangementID, accessToken, refreshToken string
}

func tokensOf(t *testing.T, tokens map[string]any) consentTokens {
	t.Helper()

	require.IsType(t, "", tokens["refresh_token"], tokens)

	return consentTokens{
		arrangementID: arrangementIDOf(t, tokens),
		accessToken:   accessTokenOf(t, tokens),
		refreshToken:  tokens["refresh_token"].(string),
	}
}

func arrangementOf(t *testing.T, customer account, arrangementID string) operatorArrangement {
	t.Helper()

	for _, a := range arrangementsOf(t, customer.customer.CustomerNumber) {
		if a.ArrangementID == arrangementID {
			return a
		}
	}

	t.Fatalf("the customer has no arrangement %s", arrangementID)

	return operatorArrangement{}
}

// This test is the full sequence of the end of eligibility: a close that leaves an open account ends nothing, and
// the close of the last open account ends each arrangement that did not end, with one notice for each.
func TestTheCloseOfTheLastOpenAccountEndsEachArrangementOfTheCustomer(t *testing.T) {
	client := registered(t, noticeProduct)
	other := registered(t, secondNoticeProduct)
	customer := newAccount(t)
	everyday := customer.addAccount("Everyday", newAccountNumber())
	savings := customer.addAccount("Savings", newAccountNumber())
	number := customer.customer.CustomerNumber

	onlyEveryday := tokensOf(t, customer.consents(client, "openid bank:accounts.basic:read", ninetyDays,
		everyday.AccountID).tokens(t))
	noAccount := tokensOf(t, customer.consents(other, noAccountScope, ninetyDays).tokens(t))
	revoked := tokensOf(t, customer.consents(client, noAccountScope, ninetyDays).tokens(t))
	require.Equal(t, http.StatusNoContent, client.revokeArrangementID(t, revoked.arrangementID).Status)
	revokedBefore := arrangementOf(t, customer, revoked.arrangementID)

	t.Run("a close while the customer has a different open account ends no arrangement", func(t *testing.T) {
		closeAccount(t, number, everyday.AccountID)

		resp := client.accountsStatus(t, onlyEveryday.accessToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
		assert.Empty(t, arrangementOf(t, customer, onlyEveryday.arrangementID).EndReason,
			"the arrangement has only the closed account")
		assert.Empty(t, arrangementOf(t, customer, noAccount.arrangementID).EndReason)
	})

	closing := time.Now().Add(-time.Second)

	closeAccount(t, number, savings.AccountID)

	closed := time.Now()

	t.Run("after the close of the last open account, each token of its arrangements stops", func(t *testing.T) {
		for name, tc := range map[string]struct {
			client registeredClient
			tokens consentTokens
		}{
			"arrangement with the closed account":                         {client, onlyEveryday},
			"arrangement of a different software product with no account": {other, noAccount},
		} {
			t.Run(name, func(t *testing.T) {
				requireRefusedToken(
					t,
					tc.client.accountsStatus(t, tc.tokens.accessToken),
					`Bearer error="invalid_token"`,
				)
				requireTokenError(
					t,
					tc.client.refresh(t, tc.tokens.refreshToken),
					http.StatusBadRequest,
					"invalid_grant",
				)
			})
		}
	})

	t.Run("each arrangement that did not end ends with CONSUMER_NOT_ELIGIBLE at the close", func(t *testing.T) {
		for _, id := range []string{onlyEveryday.arrangementID, noAccount.arrangementID} {
			got := arrangementOf(t, customer, id)

			assert.Equal(t, "CONSUMER_NOT_ELIGIBLE", got.EndReason, id)
			require.NotNil(t, got.EndedAt, id)
			assert.WithinRange(t, *got.EndedAt, closing, closed, id)
		}
	})

	t.Run("each software product gets one notice for each arrangement", func(t *testing.T) {
		for _, id := range []string{onlyEveryday.arrangementID, noAccount.arrangementID} {
			requireNoticeStatus(t, customer, id, "COMPLETE")
			assert.Len(t, labtest.NoticesFor(id), 1, id)
		}
	})

	t.Run("an arrangement that ended before the close does not change and gets no notice", func(t *testing.T) {
		assert.Equal(t, revokedBefore, arrangementOf(t, customer, revoked.arrangementID))
		assert.Empty(t, labtest.NoticesFor(revoked.arrangementID))
	})
}
