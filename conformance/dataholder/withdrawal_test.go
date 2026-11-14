//go:build conformance

package dataholder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// The operator interface acts for the consumer, who asks the bank to withdraw the consent.
func TestAWithdrawalByTheConsumerStopsEachTokenOfTheConsent(t *testing.T) {
	client := registered(t, allScopesProduct)
	customer := newAccount(t)
	own := customer.addAccount("Everyday", "70000003")

	tokens := customer.consents(client, "openid bank:accounts.basic:read", ninetyDays, own.AccountID).tokens(t)
	accessToken := accessTokenOf(t, tokens)
	require.IsType(t, "", tokens["refresh_token"], tokens)
	refreshToken := tokens["refresh_token"].(string)

	reply := withdrawConsent(t, customer.customer.CustomerNumber, arrangementIDOf(t, tokens))
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))

	requireRefusedToken(t, client.accountsStatus(t, accessToken), `Bearer error="invalid_token"`)
	requireTokenError(t, client.refresh(t, refreshToken), http.StatusBadRequest, "invalid_grant")
}
