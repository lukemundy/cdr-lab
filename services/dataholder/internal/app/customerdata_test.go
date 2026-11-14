package app_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// customerData gives the customer 00000001 with account-1 (credit limit 100.00) and a closed account-2, and a second
// customer 00000002 with account-3. The transactions are on account-1. The first customer has arrangement-1.
func customerData(t *testing.T, transactions ...domain.Transaction) app.CustomerData {
	t.Helper()

	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	store.customers["00000002"] = customer()

	limited := openAccount("account-1")
	limited.CreditLimit = 10_000
	store.accounts["00000001"] = []domain.Account{
		limited, {AccountID: "account-2", OpenStatus: domain.OpenStatusClosed},
	}
	store.accounts["00000002"] = []domain.Account{openAccount("account-3")}

	stored := &memoryTransactions{}
	for _, tr := range transactions {
		require.NoError(t, stored.AddTransaction(t.Context(), "account-1", tr))
	}

	return app.CustomerData{
		Customers:        store,
		TransactionStore: stored,
		Arrangements:     fixedArrangements{"00000001": {current("arrangement-1")}},
		Now:              func() time.Time { return transactionsNow },
	}
}

func TestTheAccountsOfACustomerAreInTheSequenceOfTheirCreationWithTheBalanceOfTheBankingAPIs(t *testing.T) {
	salary := transaction("Salary", transactionsNow)
	salary.Amount = 250_000
	pending := transaction("Coffee", transactionsNow)
	pending.Status, pending.PostingDateTime = domain.TransactionPending, nil
	pending.ExecutionDateTime = &transactionsNow
	data := customerData(t, salary, transaction("Lunch", transactionsNow), pending)

	got, err := data.Accounts(t.Context(), "00000001")

	require.NoError(t, err)
	require.Len(t, got.Accounts, 2)
	assert.Equal(t, "account-1", got.Accounts[0].Account.AccountID)
	assert.Equal(t, domain.Balance{Current: 249_550, Available: 259_100, CreditLimit: 10_000}, got.Accounts[0].Balance)
	assert.Equal(t, "account-2", got.Accounts[1].Account.AccountID, "a closed account is in the list")
	assert.Equal(t, domain.Balance{}, got.Accounts[1].Balance, "an account with no transactions has zero")
}

func TestTheAccountsOfACustomerComeWithTheArrangementsOfTheCustomer(t *testing.T) {
	data := customerData(t)

	got, err := data.Accounts(t.Context(), "00000001")

	require.NoError(t, err)
	assert.Equal(t, []domain.ArrangementState{current("arrangement-1")}, got.Arrangements)
}

func TestACustomerWithNoAccountsHasAnEmptyList(t *testing.T) {
	data := customerData(t)
	data.Customers.(*fakeCustomerStore).customers["00000003"] = customer()

	got, err := data.Accounts(t.Context(), "00000003")

	require.NoError(t, err)
	assert.NotNil(t, got.Accounts)
	assert.Empty(t, got.Accounts)
}

func TestTheBalanceOfAnAccountOfTheCustomerComesFromItsTransactions(t *testing.T) {
	data := customerData(t, transaction("Lunch", transactionsNow))

	got, err := data.Balance(t.Context(), "00000001", "account-1")

	require.NoError(t, err)
	assert.Equal(t, domain.Balance{Current: -450, Available: 9_550, CreditLimit: 10_000}, got)
}

func TestTheTransactionsOfAnAccountOfTheCustomerHaveTheQueryOfGetTransactionsForAccountWithNoFilter(t *testing.T) {
	data := customerData(t, transaction("Coffee", transactionsNow))

	found, total, err := data.Transactions(t.Context(), "00000001", "account-1", 2, 50)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, found, 1)
	assert.Equal(t, "Coffee", found[0].Description)

	query := data.TransactionStore.(*memoryTransactions).query
	assert.Equal(t, app.TransactionQuery{
		AccountID: "account-1",
		Oldest:    transactionsNow.AddDate(0, 0, -90),
		Newest:    transactionsNow,
		Page:      2,
		PageSize:  50,
	}, query)
}

func TestTheDataOfAnAccountOfADifferentCustomerOrOfACustomerThatDoesNotExistIsNotFound(t *testing.T) {
	data := customerData(t)

	_, err := data.Balance(t.Context(), "00000002", "account-1")
	require.ErrorIs(t, err, app.ErrAccountNotFound)

	_, _, err = data.Transactions(t.Context(), "00000002", "account-1", 1, 25)
	require.ErrorIs(t, err, app.ErrAccountNotFound)

	_, err = data.Accounts(t.Context(), "00000009")
	require.ErrorIs(t, err, app.ErrCustomerNotFound)

	_, err = data.Balance(t.Context(), "00000009", "account-1")
	assert.ErrorIs(t, err, app.ErrCustomerNotFound)
}
