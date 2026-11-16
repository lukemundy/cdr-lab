package app_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// memoryTransactions keeps the transactions of each account, and the last query.
type memoryTransactions struct {
	accounts map[string][]domain.Transaction
	query    app.TransactionQuery
}

func (m *memoryTransactions) AddTransaction(_ context.Context, accountID string, t domain.Transaction) error {
	if m.accounts == nil {
		m.accounts = map[string][]domain.Transaction{}
	}

	m.accounts[accountID] = append(m.accounts[accountID], t)

	return nil
}

func (m *memoryTransactions) Transactions(_ context.Context, accountID string) ([]domain.Transaction, error) {
	return m.accounts[accountID], nil
}

// FindTransactions applies only the time limits, and gives the transactions in the sequence of their creation.
func (m *memoryTransactions) FindTransactions(
	_ context.Context,
	q app.TransactionQuery,
) ([]domain.Transaction, int, error) {
	m.query = q
	found := slices.DeleteFunc(slices.Clone(m.accounts[q.AccountID]), func(t domain.Transaction) bool {
		return t.EffectiveTime().Before(q.Oldest) || t.EffectiveTime().After(q.Newest)
	})

	return found, len(found), nil
}

func (m *memoryTransactions) Totals(
	_ context.Context,
	accountIDs []string,
) (map[string]domain.TransactionTotals, error) {
	totals := map[string]domain.TransactionTotals{}

	for _, id := range accountIDs {
		for _, t := range m.accounts[id] {
			sum := totals[id]

			switch {
			case t.Status == domain.TransactionPosted:
				sum.Posted += t.Amount
			case t.Amount < 0:
				sum.PendingDebits += t.Amount
			}

			totals[id] = sum
		}
	}

	return totals, nil
}

var transactionsNow = time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)

func transaction(description string, at time.Time) domain.Transaction {
	return domain.Transaction{
		Type: domain.TransactionPayment, Status: domain.TransactionPosted, Description: description, Amount: -450,
		PostingDateTime: &at,
	}
}

func withDetail(t domain.Transaction) domain.Transaction {
	t.ExtendedData = &domain.ExtendedData{
		Payee: "Cafe", ExtensionUType: domain.ExtensionNPPPayload,
		NPPPayload: domain.NPPPayload{
			Service: domain.NPPServiceX2P1, ServiceVersion: "03", ExtendedDescription: "Coffee",
		},
	}

	return t
}

// operatorCustomers gives a customer with an open account and a closed account, and a second customer.
func operatorCustomers() (app.Customers, *memoryTransactions) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	store.customers["00000002"] = customer()
	store.accounts["00000001"] = []domain.Account{
		openAccount("account-1"), {AccountID: "account-2", OpenStatus: domain.OpenStatusClosed},
	}
	transactions := &memoryTransactions{}
	next := 0

	return app.Customers{
		Store: store, Transactions: transactions, Now: now,
		NewID: func() string {
			next++
			return "transaction-" + string(rune('0'+next))
		},
	}, transactions
}

func TestTheOperatorAddsATransactionInUTCWithANewIdentifier(t *testing.T) {
	customers, transactions := operatorCustomers()
	perth := transactionsNow.In(time.FixedZone("AWST", 8*60*60))
	requested := transaction("Coffee", perth)
	requested.TransactionID = "of-the-request"

	created, err := customers.AddTransaction(t.Context(), "00000001", "account-1", requested)

	require.NoError(t, err)
	assert.Equal(t, "transaction-1", created.TransactionID)
	assert.Equal(t, time.UTC, created.PostingDateTime.Location())
	assert.Equal(t, []domain.Transaction{created}, transactions.accounts["account-1"])

	read, err := customers.AccountTransactions(t.Context(), "00000001", "account-1")
	require.NoError(t, err)
	assert.Equal(t, []domain.Transaction{created}, read)
}

func TestTheOperatorCannotAddATransactionThatBreaksARule(t *testing.T) {
	customers, transactions := operatorCustomers()

	_, err := customers.AddTransaction(t.Context(), "00000001", "account-1", transaction("", transactionsNow))

	assert.Equal(t, []string{"description"}, fieldsOf(t, err))
	assert.Empty(t, transactions.accounts)
}

func TestTheOperatorCannotAddATransactionForAnAccountThatCannotHaveIt(t *testing.T) {
	tests := map[string]struct {
		customerNumber, accountID string
		want                      error
	}{
		"closed account": {
			customerNumber: "00000001",
			accountID:      "account-2",
			want:           domain.ErrAccountClosed,
		},
		"account of a different customer": {
			customerNumber: "00000002",
			accountID:      "account-1",
			want:           app.ErrAccountNotFound,
		},
		"no customer": {
			customerNumber: "99999999",
			accountID:      "account-1",
			want:           app.ErrCustomerNotFound,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			customers, transactions := operatorCustomers()

			_, err := customers.AddTransaction(
				t.Context(), tt.customerNumber, tt.accountID, transaction("Coffee", transactionsNow),
			)

			require.ErrorIs(t, err, tt.want)
			assert.Empty(t, transactions.accounts)
		})
	}
}

// consentWithTransactions gives the consent data of a consent of account-2 and account-1 of the consent customer,
// with transactions on account-1.
func consentWithTransactions(t *testing.T, transactions ...domain.Transaction) (app.ConsentData, string, string) {
	t.Helper()

	data, arrangementID := consentWithAccounts(t)
	store := &memoryTransactions{}

	for i, tr := range transactions {
		tr.TransactionID = "transaction-" + string(rune('1'+i))
		require.NoError(t, store.AddTransaction(t.Context(), "account-1", tr))
	}

	data.TransactionStore = store
	data.Now = func() time.Time { return transactionsNow }

	return data, arrangementID, idOfAccount1
}

// Verifies: DH-1040
func TestTheTransactionsOfAnAccountOfTheConsentHaveTheIDsOfTheSoftwareProductAndCustomer(t *testing.T) {
	data, arrangementID, accountID := consentWithTransactions(t, transaction("Coffee", transactionsNow))

	got, total, err := data.Transactions(t.Context(), arrangementID, accountID, app.TransactionFilter{
		Page: 1, PageSize: 25,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, got, 1)
	assert.Equal(t, domain.ResourceID("software-1", consentCustomer, "transaction-1"), got[0].ID)
	assert.Equal(t, "Coffee", got[0].Transaction.Description)
}

// Verifies: DH-1085, DH-1088, DH-1089
func TestTheTimeLimitsOfTransactionsHaveTheirDefaultsAndTheTwoYearLimit(t *testing.T) {
	newest := transactionsNow.Add(-time.Hour)
	oldest := transactionsNow.AddDate(-3, 0, 0)

	tests := map[string]struct {
		filter                 app.TransactionFilter
		wantOldest, wantNewest time.Time
	}{
		"no limits": {
			wantOldest: transactionsNow.AddDate(0, 0, -90), wantNewest: transactionsNow,
		},
		"newest only": {
			filter:     app.TransactionFilter{Newest: &newest},
			wantOldest: newest.AddDate(0, 0, -90), wantNewest: newest,
		},
		"oldest before the two years": {
			filter:     app.TransactionFilter{Oldest: &oldest},
			wantOldest: transactionsNow.AddDate(-2, 0, 0), wantNewest: transactionsNow,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			data, arrangementID, accountID := consentWithTransactions(t)
			tt.filter.Page, tt.filter.PageSize = 1, 25

			_, _, err := data.Transactions(t.Context(), arrangementID, accountID, tt.filter)

			require.NoError(t, err)

			query := data.TransactionStore.(*memoryTransactions).query
			assert.Equal(t, tt.wantOldest, query.Oldest)
			assert.Equal(t, tt.wantNewest, query.Newest)
			assert.Equal(t, "account-1", query.AccountID)
		})
	}
}

// Verifies: DH-1084
func TestTheOtherFiltersOfTransactionsGoToTheQuery(t *testing.T) {
	data, arrangementID, accountID := consentWithTransactions(t)
	minAmount, maxAmount := domain.Amount(-100), domain.Amount(100)

	_, _, err := data.Transactions(t.Context(), arrangementID, accountID, app.TransactionFilter{
		MinAmount: &minAmount, MaxAmount: &maxAmount, Text: "coffee", Page: 2, PageSize: 10,
	})

	require.NoError(t, err)

	query := data.TransactionStore.(*memoryTransactions).query
	assert.Equal(t, &minAmount, query.MinAmount)
	assert.Equal(t, &maxAmount, query.MaxAmount)
	assert.Equal(t, "coffee", query.Text)
	assert.Equal(t, 2, query.Page)
	assert.Equal(t, 10, query.PageSize)
}

// Verifies: DH-1091
func TestATransactionWithDetailIsFoundByItsIDInTheTwoYears(t *testing.T) {
	future := transactionsNow.Add(24 * time.Hour)
	data, arrangementID, accountID := consentWithTransactions(t,
		withDetail(transaction("Recent", transactionsNow.AddDate(-1, 0, 0))),
		withDetail(transaction("Future", future)),
	)

	for i, want := range []string{"Recent", "Future"} {
		id := domain.ResourceID("software-1", consentCustomer, "transaction-"+string(rune('1'+i)))

		got, err := data.Transaction(t.Context(), arrangementID, accountID, id)

		require.NoError(t, err, want)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, want, got.Transaction.Description)
	}
}

// Verifies: DH-1091
func TestATransactionThatHasNoDetailOrIsNotInTheTwoYearsOrIsNotOfTheAccountIsNotFound(t *testing.T) {
	data, arrangementID, accountID := consentWithTransactions(t,
		transaction("No detail", transactionsNow),
		withDetail(transaction("Too old", transactionsNow.AddDate(-2, 0, 0).Add(-time.Second))),
	)

	for _, id := range []string{
		domain.ResourceID("software-1", consentCustomer, "transaction-1"),
		domain.ResourceID("software-1", consentCustomer, "transaction-2"),
		domain.ResourceID("software-2", consentCustomer, "transaction-1"),
		"transaction-1",
	} {
		_, err := data.Transaction(t.Context(), arrangementID, accountID, id)

		assert.ErrorIs(t, err, app.ErrTransactionNotFound, id)
	}
}

func TestTheTransactionsOfAnAccountThatTheConsentDoesNotGiveAreNotFound(t *testing.T) {
	data, arrangementID, _ := consentWithTransactions(t)

	_, _, listErr := data.Transactions(t.Context(), arrangementID, "account-1", app.TransactionFilter{
		Page: 1, PageSize: 25,
	})
	_, detailErr := data.Transaction(t.Context(), arrangementID, "account-1", "transaction-1")

	require.ErrorIs(t, listErr, app.ErrAccountNotFound)
	require.ErrorIs(t, detailErr, app.ErrAccountNotFound)
	assert.NotContains(t, listErr.Error(), "account-1", "the error has no identifier")
}
