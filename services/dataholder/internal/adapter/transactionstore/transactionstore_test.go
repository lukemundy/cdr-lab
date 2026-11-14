package transactionstore_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

const (
	firstAccount  = "account-1"
	secondAccount = "account-2"
)

var now = time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)

// newStore gives a database with one customer, who has two accounts.
func newStore(t *testing.T) (*transactionstore.Store, *pgxpool.Pool) {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	customers := customerstore.New(pool, nil, nil)
	require.NoError(t, customers.AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: "00000001", LastName: "Example", EmailAddress: "alex@example.com",
	}))

	for i, id := range []string{firstAccount, secondAccount} {
		require.NoError(t, customers.AddAccount(t.Context(), "00000001", domain.Account{
			AccountID: id, DisplayName: "Everyday", ProductCategory: domain.CategoryTransAndSavingsAccounts,
			ProductName: "Mock Everyday Account", BSB: "123456", AccountNumber: []string{"111111", "222222"}[i],
			OpenStatus: domain.OpenStatusOpen,
		}))
	}

	return transactionstore.New(pool), pool
}

func posted(id string, amount domain.Amount, at time.Time) domain.Transaction {
	return domain.Transaction{
		TransactionID: id, Type: domain.TransactionPayment, Status: domain.TransactionPosted,
		Description: "Coffee " + id, Amount: amount, PostingDateTime: &at,
	}
}

func ids(transactions []domain.Transaction) []string {
	found := []string{}
	for _, t := range transactions {
		found = append(found, t.TransactionID)
	}

	return found
}

func query(change func(*app.TransactionQuery)) app.TransactionQuery {
	q := app.TransactionQuery{
		AccountID: firstAccount, Oldest: now.AddDate(0, 0, -90), Newest: now, Page: 1, PageSize: 25,
	}
	if change != nil {
		change(&q)
	}

	return q
}

func TestATransactionIsReadWithEachFieldInUTC(t *testing.T) {
	store, pool := newStore(t)
	perth := time.FixedZone("AWST", 8*60*60)
	executed := now.Add(-time.Hour).In(perth)
	value := now.Add(-2 * time.Hour)
	full := domain.Transaction{
		TransactionID: "transaction-1", Type: domain.TransactionTransferOutgoing, Status: domain.TransactionPosted,
		Description: "Rent", Amount: -120_000, Reference: "October", PostingDateTime: new(now.In(perth)),
		ExecutionDateTime: &executed, ValueDateTime: &value, MerchantName: "Agent", MerchantCategoryCode: "6513",
		BillerCode: "12345", BillerName: "Agent Pty Ltd", CRN: "987654", APCANumber: "123456",
		ExtendedData: &domain.ExtendedData{
			Payee: "Landlord", ExtensionUType: domain.ExtensionNPPPayload,
			NPPPayload: domain.NPPPayload{
				ExtendedDescription: "Rent for October", Service: domain.NPPServiceX2P1, ServiceVersion: "03",
			},
		},
	}
	require.NoError(t, store.AddTransaction(t.Context(), firstAccount, full))

	// A new store on the same database is the data holder after a restart.
	got, err := transactionstore.New(pool).Transactions(t.Context(), firstAccount)

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, time.UTC, got[0].PostingDateTime.Location())
	assert.Equal(t, time.UTC, got[0].ExecutionDateTime.Location())

	full.PostingDateTime, full.ExecutionDateTime = new(now), new(executed.UTC())
	assert.Equal(t, full, got[0])
}

func TestTheTransactionsOfAnAccountAreInTheSequenceOfTheirCreation(t *testing.T) {
	store, _ := newStore(t)

	for _, tr := range []domain.Transaction{
		posted("b", -100, now.Add(-time.Hour)), posted("a", -100, now.Add(-2*time.Hour)),
	} {
		require.NoError(t, store.AddTransaction(t.Context(), firstAccount, tr))
	}

	require.NoError(t, store.AddTransaction(t.Context(), secondAccount, posted("c", -100, now)))

	got, err := store.Transactions(t.Context(), firstAccount)

	require.NoError(t, err)
	assert.Equal(t, []string{"b", "a"}, ids(got))
}

func TestFindTransactionsGivesTheEffectiveTimeInDescendingOrderAndThenTheLastCreatedFirst(t *testing.T) {
	store, _ := newStore(t)
	pending := posted("pending", -100, now)
	pending.Status, pending.PostingDateTime, pending.ExecutionDateTime = domain.TransactionPending, nil,
		new(now.Add(-time.Minute))

	for _, tr := range []domain.Transaction{
		posted("old", -100, now.Add(-time.Hour)),
		posted("same-1", -100, now.Add(-30*time.Minute)),
		posted("same-2", -100, now.Add(-30*time.Minute)),
		pending,
	} {
		require.NoError(t, store.AddTransaction(t.Context(), firstAccount, tr))
	}

	require.NoError(t, store.AddTransaction(t.Context(), secondAccount, posted("other-account", -100, now)))

	got, total, err := store.FindTransactions(t.Context(), query(nil))

	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Equal(t, []string{"pending", "same-2", "same-1", "old"}, ids(got))
}

func TestFindTransactionsMatchesEachFilter(t *testing.T) {
	store, _ := newStore(t)
	withReference := posted("reference", 2_500, now.Add(-2*time.Hour))
	withReference.Reference = "Pay 100%_OCTOBER"

	for _, tr := range []domain.Transaction{
		posted("oldest-limit", -1_000, now.AddDate(0, 0, -90)),
		posted("too-old", -1_000, now.AddDate(0, 0, -90).Add(-time.Second)),
		posted("newest-limit", -500, now),
		posted("too-new", -500, now.Add(time.Second)),
		withReference,
	} {
		require.NoError(t, store.AddTransaction(t.Context(), firstAccount, tr))
	}

	tests := map[string]struct {
		change func(*app.TransactionQuery)
		want   []string
	}{
		"inclusive time limits": {want: []string{"newest-limit", "reference", "oldest-limit"}},
		"min-amount": {
			change: func(q *app.TransactionQuery) { q.MinAmount = new(domain.Amount(-500)) },
			want:   []string{"newest-limit", "reference"},
		},
		"max-amount": {
			change: func(q *app.TransactionQuery) { q.MaxAmount = new(domain.Amount(-500)) },
			want:   []string{"newest-limit", "oldest-limit"},
		},
		"description with other case": {
			change: func(q *app.TransactionQuery) { q.Text = "COFFEE NEWEST" },
			want:   []string{"newest-limit"},
		},
		"reference with the characters of a LIKE pattern": {
			change: func(q *app.TransactionQuery) { q.Text = "100%_oct" },
			want:   []string{"reference"},
		},
		"each filter": {
			change: func(q *app.TransactionQuery) {
				q.MinAmount, q.Text = new(domain.Amount(0)), "coffee"
			},
			want: []string{"reference"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, total, err := store.FindTransactions(t.Context(), query(tt.change))

			require.NoError(t, err)
			assert.Equal(t, tt.want, ids(got))
			assert.Equal(t, len(tt.want), total)
		})
	}
}

func TestTotalsSumThePostedAmountsAndTheNegativePendingAmountsOfEachAccount(t *testing.T) {
	store, _ := newStore(t)
	pending := func(id string, amount domain.Amount) domain.Transaction {
		tr := posted(id, amount, now)
		tr.Status, tr.PostingDateTime, tr.ExecutionDateTime = domain.TransactionPending, nil, &now

		return tr
	}

	for _, tr := range []domain.Transaction{
		posted("salary", 250_000, now.AddDate(-3, 0, 0)),
		posted("rent", -120_000, now),
		pending("coffee", -450),
		pending("refund", 2_000),
	} {
		require.NoError(t, store.AddTransaction(t.Context(), firstAccount, tr))
	}

	got, err := store.Totals(t.Context(), []string{firstAccount, secondAccount})

	require.NoError(t, err)
	assert.Equal(t, map[string]domain.TransactionTotals{
		firstAccount: {Posted: 130_000, PendingDebits: -450},
	}, got)
}

func TestFindTransactionsGivesOnePage(t *testing.T) {
	store, _ := newStore(t)

	for i, id := range []string{"a", "b", "c"} {
		require.NoError(t, store.AddTransaction(
			t.Context(), firstAccount, posted(id, -100, now.Add(-time.Duration(i)*time.Hour)),
		))
	}

	got, total, err := store.FindTransactions(t.Context(), query(func(q *app.TransactionQuery) {
		q.Page, q.PageSize = 2, 2
	}))

	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, []string{"c"}, ids(got))
}

func TestAnInsertInADatabaseTransactionAddsTheTransactionAtItsCommit(t *testing.T) {
	store, pool := newStore(t)

	require.NoError(t, pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error {
		return transactionstore.Insert(t.Context(), tx, firstAccount, posted("t1", -100, now))
	}))

	got, err := store.Transactions(t.Context(), firstAccount)
	require.NoError(t, err)
	assert.Equal(t, []string{"t1"}, ids(got))
}
