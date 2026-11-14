package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type fixedArrangements map[string][]domain.ArrangementState

func (f fixedArrangements) ArrangementsOf(_ context.Context, customerNumber string) ([]domain.ArrangementState, error) {
	return f[customerNumber], nil
}

func (f fixedArrangements) ArrangementsOfSoftwareProduct(
	context.Context,
	string,
) ([]domain.ArrangementState, error) {
	return nil, nil
}

// customerDetails gives the customer 00000001 with account-1 (credit limit 100.00) and account-2, and a second
// customer 00000002 with no accounts.
func customerDetails(t *testing.T, transactions ...domain.Transaction) app.CustomerDetails {
	t.Helper()

	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	store.customers["00000002"] = customer()

	limited := openAccount("account-1")
	limited.CreditLimit = 10_000
	store.accounts["00000001"] = []domain.Account{limited, openAccount("account-2")}

	stored := &memoryTransactions{}
	for _, tr := range transactions {
		require.NoError(t, stored.AddTransaction(t.Context(), "account-1", tr))
	}

	return app.CustomerDetails{
		Customers:    store,
		Transactions: stored,
		Arrangements: fixedArrangements{"00000001": {
			{Created: true, Arrangement: domain.Arrangement{ArrangementID: "arrangement-1", ClientID: "client-1"}},
			{Created: true, Arrangement: domain.Arrangement{ArrangementID: "arrangement-2", ClientID: "client-2"}},
		}},
		Registrations: fakeRegistrations{"client-1": {ClientID: "client-1"}},
	}
}

func TestTheDetailOfACustomerHasEachAccountWithTheBalanceOfTheBankingAPIsAndEachArrangement(t *testing.T) {
	salary := transaction("Salary", transactionsNow)
	salary.Amount = 250_000
	pending := transaction("Coffee", transactionsNow)
	pending.Status, pending.PostingDateTime = domain.TransactionPending, nil
	pending.ExecutionDateTime = &transactionsNow
	details := customerDetails(t, salary, transaction("Lunch", transactionsNow), pending)

	detail, err := details.Detail(t.Context(), "00000001")

	require.NoError(t, err)
	assert.Equal(t, customer(), detail.Customer)
	require.Len(t, detail.Accounts, 2)
	assert.Equal(t, "account-1", detail.Accounts[0].Account.AccountID)
	assert.Equal(t, domain.Balance{Current: 249_550, Available: 259_100, CreditLimit: 10_000},
		detail.Accounts[0].Balance)
	assert.Equal(t, domain.Balance{}, detail.Accounts[1].Balance, "an account with no transactions has zero")
	require.Len(t, detail.Arrangements, 2)
	assert.Equal(t, "arrangement-1", detail.Arrangements[0].Arrangement.ArrangementID)
	assert.True(t, detail.Arrangements[0].ClientRegistered)
	assert.False(t, detail.Arrangements[1].ClientRegistered, "the client of arrangement-2 has no registration")
}

func TestTheDetailOfACustomerWithNoAccountsOrArrangementsHasEmptyLists(t *testing.T) {
	detail, err := customerDetails(t).Detail(t.Context(), "00000002")

	require.NoError(t, err)
	assert.Empty(t, detail.Accounts)
	assert.Empty(t, detail.Arrangements)
}

func TestTheDetailOfACustomerThatDoesNotExistIsNotFound(t *testing.T) {
	_, err := customerDetails(t).Detail(t.Context(), "00000009")

	assert.ErrorIs(t, err, app.ErrCustomerNotFound)
}

func TestTheTransactionPageOfAnAccountHasNoTimeLimitAndThePageOfTheRequest(t *testing.T) {
	details := customerDetails(t, transaction("Coffee", transactionsNow))

	found, total, err := details.TransactionPage(t.Context(), "00000001", "account-1", 2, 50)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, found, 1)
	assert.Equal(t, "Coffee", found[0].Description)

	query := details.Transactions.(*memoryTransactions).query
	assert.Equal(t, "account-1", query.AccountID)
	assert.True(t, query.Oldest.IsZero(), "no oldest limit")
	assert.True(t, query.Newest.After(transactionsNow.AddDate(1000, 0, 0)), "no newest limit")
	assert.Equal(t, 2, query.Page)
	assert.Equal(t, 50, query.PageSize)
}

func TestTheTransactionPageNeedsAnAccountOfTheCustomer(t *testing.T) {
	details := customerDetails(t)

	_, _, err := details.TransactionPage(t.Context(), "00000009", "account-1", 1, 25)
	require.ErrorIs(t, err, app.ErrCustomerNotFound)

	_, _, err = details.TransactionPage(t.Context(), "00000002", "account-1", 1, 25)
	assert.ErrorIs(t, err, app.ErrAccountNotFound)
}
