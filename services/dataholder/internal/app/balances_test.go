package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// consentWithBalances gives the consent data of a consent of account-2 and account-1. Only account-1 has transactions,
// and it is closed.
func consentWithBalances(t *testing.T) (app.ConsentData, string) {
	t.Helper()

	salary := transaction("Salary", transactionsNow)
	salary.Amount = 250_000
	pending := transaction("Coffee", transactionsNow)
	pending.Status, pending.PostingDateTime = domain.TransactionPending, nil
	pending.ExecutionDateTime = &transactionsNow

	data, arrangementID, _ := consentWithTransactions(t, salary, transaction("Lunch", transactionsNow), pending)
	_, err := app.Customers{Store: data.Customers, Now: now}.CloseAccount(t.Context(), consentCustomer, "account-1")
	require.NoError(t, err)

	return data, arrangementID
}

func balanceIDs(balances []app.ConsentBalance) []string {
	found := []string{}
	for _, b := range balances {
		found = append(found, b.ID)
	}

	return found
}

// Verifies: DH-1013, DH-1014
func TestTheBalanceOfAnAccountOfTheConsentComesFromItsTransactions(t *testing.T) {
	data, arrangementID := consentWithBalances(t)

	got, err := data.Balance(t.Context(), arrangementID, idOfAccount1)

	require.NoError(t, err)
	assert.Equal(t, app.ConsentBalance{
		ID:      idOfAccount1,
		Balance: domain.Balance{Current: 249_550, Available: 249_100},
	}, got)
}

func TestAnAccountWithNoTransactionsHasABalanceOfZero(t *testing.T) {
	data, arrangementID := consentWithBalances(t)

	got, err := data.Balance(t.Context(), arrangementID, idOfAccount2)

	require.NoError(t, err)
	assert.Equal(t, domain.Balance{}, got.Balance)
}

func TestTheBalanceOfAnAccountThatTheConsentDoesNotGiveIsNotFound(t *testing.T) {
	data, arrangementID := consentWithBalances(t)

	_, err := data.Balance(t.Context(), arrangementID, "account-1")

	require.ErrorIs(t, err, app.ErrAccountNotFound)
}

func TestTheBalancesOfAConsentMatchTheFiltersOfTheAccountsAndArePaginated(t *testing.T) {
	data, arrangementID := consentWithBalances(t)

	all, total, err := data.Balances(t.Context(), arrangementID, app.AccountQuery{Page: 1, PageSize: 25})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, []string{idOfAccount1, idOfAccount2}, balanceIDs(all))
	assert.Equal(t, domain.Amount(249_550), all[0].Balance.Current)

	open, total, err := data.Balances(t.Context(), arrangementID, app.AccountQuery{
		OpenStatus: domain.OpenStatusOpen, Page: 1, PageSize: 25,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, []string{idOfAccount2}, balanceIDs(open))

	second, total, err := data.Balances(t.Context(), arrangementID, app.AccountQuery{Page: 2, PageSize: 1})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, []string{idOfAccount2}, balanceIDs(second))
}

// Verifies: DH-1074
func TestTheBalancesOfSpecificAccountsAreInTheSequenceOfTheConsentWithOneForEachAccount(t *testing.T) {
	data, arrangementID := consentWithBalances(t)

	tests := map[string]struct {
		ids  []string
		want []string
	}{
		"each account":        {ids: []string{idOfAccount2, idOfAccount1}, want: []string{idOfAccount1, idOfAccount2}},
		"one account":         {ids: []string{idOfAccount2}, want: []string{idOfAccount2}},
		"the same ID 2 times": {ids: []string{idOfAccount1, idOfAccount1}, want: []string{idOfAccount1}},
		"no IDs":              {ids: []string{}, want: []string{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, total, err := data.Balances(t.Context(), arrangementID, app.AccountQuery{
				IDs: tt.ids, Page: 1, PageSize: 25,
			})

			require.NoError(t, err)
			assert.Equal(t, tt.want, balanceIDs(got))
			assert.Equal(t, len(tt.want), total)
		})
	}
}

func TestSpecificAccountsThatTheConsentDoesNotGiveAreEachNamedOneTime(t *testing.T) {
	data, arrangementID := consentWithBalances(t)

	_, _, err := data.Balances(t.Context(), arrangementID, app.AccountQuery{
		IDs: []string{"unknown-2", idOfAccount2, "account-1", "unknown-2"}, Page: 1, PageSize: 25,
	})

	var unknown app.UnknownAccountsError
	require.ErrorAs(t, err, &unknown)
	assert.Equal(t, []string{"unknown-2", "account-1"}, unknown.IDs)
}
