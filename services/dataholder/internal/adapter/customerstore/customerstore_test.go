package customerstore_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

var closedAt = time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

func customer(number string) domain.Customer {
	return domain.Customer{
		CustomerNumber: number,
		FirstName:      "Alex",
		LastName:       "Example",
		EmailAddress:   "alex." + number + "@example.com",
		LastUpdated:    time.Date(2026, 10, 15, 13, 30, 0, 123456000, time.UTC),
	}
}

func account(number string) domain.Account {
	return domain.Account{
		AccountID:       "account-" + number,
		DisplayName:     "Everyday " + number,
		ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName:     "Mock Everyday Account",
		BSB:             "123456",
		AccountNumber:   number,
		CreditLimit:     150_000,
		OpenStatus:      domain.OpenStatusOpen,
		CreatedAt:       time.Date(2026, 10, 15, 13, 30, 0, 123456000, time.UTC),
	}
}

func TestACustomerIsReadWithItsAccountsInTheSequenceOfTheirCreation(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)

	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000002")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("222222")))
	require.NoError(t, store.AddAccount(t.Context(), "00000002", account("333333")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))

	record, err := store.Customer(t.Context(), "00000001")

	require.NoError(t, err)
	assert.Equal(
		t,
		app.CustomerRecord{
			Customer: customer("00000001"),
			Accounts: []domain.Account{account("222222"), account("111111")},
		},
		record,
	)
}

func TestACustomerWithNoFirstNameAndNoAccountsIsRead(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	oneName := customer("00000001")
	oneName.FirstName = ""

	require.NoError(t, store.AddCustomer(t.Context(), oneName))

	record, err := store.Customer(t.Context(), "00000001")

	require.NoError(t, err)
	assert.Equal(t, oneName, record.Customer)
	assert.Empty(t, record.Accounts)
}

func TestACustomerStaysForANewStoreOnTheSameDatabase(t *testing.T) {
	pool := newPool(t)

	require.NoError(t, customerstore.New(pool, nil, nil).AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, customerstore.New(pool, nil, nil).AddAccount(t.Context(), "00000001", account("111111")))

	record, err := customerstore.New(pool, nil, nil).Customer(t.Context(), "00000001")

	require.NoError(t, err)
	assert.Equal(t, customer("00000001"), record.Customer)
	assert.Len(t, record.Accounts, 1)
}

func TestASecondCustomerWithTheSameCustomerNumberIsRefused(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)

	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))

	require.ErrorIs(t, store.AddCustomer(t.Context(), customer("00000001")), app.ErrCustomerNumberUsed)
}

func TestASecondCustomerWithTheSameEmailAddressInADifferentCaseIsRefused(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	second := customer("00000002")
	second.EmailAddress = "Alex.00000001@Example.COM"

	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))

	require.ErrorIs(t, store.AddCustomer(t.Context(), second), app.ErrEmailAddressUsed)
	_, err := store.Customer(t.Context(), "00000002")
	require.ErrorIs(t, err, app.ErrCustomerNotFound)
}

func TestTheCustomerNumberOfAnEmailAddressIsFoundWithNoDifferenceForCase(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	mixedCase := customer("00000001")
	mixedCase.EmailAddress = "Alex.Example@Mail.test"

	require.NoError(t, store.AddCustomer(t.Context(), mixedCase))
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000002")))

	for _, address := range []string{"Alex.Example@Mail.test", "alex.example@mail.test", "ALEX.EXAMPLE@MAIL.TEST"} {
		number, err := store.CustomerNumberOf(t.Context(), address)

		require.NoError(t, err, address)
		assert.Equal(t, "00000001", number, address)
	}

	record, err := store.Customer(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Equal(t, "Alex.Example@Mail.test", record.Customer.EmailAddress, "the address stays as given")
}

func TestAnEmailAddressThatNoCustomerHasGivesNotFound(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))

	_, err := store.CustomerNumberOf(t.Context(), "alex.00000002@example.com")

	require.ErrorIs(t, err, app.ErrCustomerNotFound)
}

func TestACustomerNumberThatDoesNotExistGivesNotFound(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)

	_, readErr := store.Customer(t.Context(), "00000001")
	addErr := store.AddAccount(t.Context(), "00000001", account("111111"))

	require.ErrorIs(t, readErr, app.ErrCustomerNotFound)
	require.ErrorIs(t, addErr, app.ErrCustomerNotFound)
}

func TestASecondAccountWithTheSameBSBAndAccountNumberIsRefused(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)

	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000002")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))

	second := account("111111")
	second.AccountID = "account-2"

	sameCustomer := store.AddAccount(t.Context(), "00000001", second)
	otherCustomer := store.AddAccount(t.Context(), "00000002", second)

	require.ErrorIs(t, sameCustomer, app.ErrAccountExists)
	require.ErrorIs(t, otherCustomer, app.ErrAccountExists)
}

func TestTheSameAccountNumberWithADifferentBSBIsAccepted(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	otherBSB := account("111111")
	otherBSB.AccountID = "account-2"
	otherBSB.BSB = "654321"

	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))

	require.NoError(t, store.AddAccount(t.Context(), "00000001", otherBSB))
}

func TestASecondAccountWithTheSameIdentifierIsAnErrorThatIsNotAConflictOfAccountNumbers(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	sameID := account("222222")
	sameID.AccountID = account("111111").AccountID

	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))

	err := store.AddAccount(t.Context(), "00000001", sameID)

	require.Error(t, err)
	assert.NotErrorIs(t, err, app.ErrAccountExists)
}

func TestAClosedAccountIsReadAsClosed(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	closed := account("111111")
	closed.OpenStatus = domain.OpenStatusClosed

	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("222222")))

	_, err := store.CloseAccount(t.Context(), closed, closedAt, nil)
	require.NoError(t, err)

	record, err := store.Customer(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Equal(t, []domain.Account{closed, account("222222")}, record.Accounts)
}

func TestTheTimesOfAnAccountAreReadInUTC(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	perth := account("111111")
	perth.CreatedAt = perth.CreatedAt.In(time.FixedZone("AWST", 8*60*60))

	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", perth))

	record, err := store.Customer(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Equal(t, time.UTC, record.Accounts[0].CreatedAt.Location())
}

// hookCalls records each call of the hook, and makes the hook fail with err.
type hookCalls struct {
	mu        sync.Mutex
	customers []string
	at        []time.Time
	err       error
}

func (h *hookCalls) hook(ctx context.Context, tx pgx.Tx, customerNumber string, closedAt time.Time) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// The hook runs in the transaction of the close, so it sees the close.
	var status string
	if err := tx.QueryRow(ctx, "SELECT open_status FROM accounts WHERE customer_number = $1 LIMIT 1",
		customerNumber).Scan(&status); err != nil {
		return err
	}

	if status != string(domain.OpenStatusClosed) {
		return errors.New("the hook does not see the close")
	}

	h.customers, h.at = append(h.customers, customerNumber), append(h.at, closedAt)

	return h.err
}

func TestOnlyTheCloseOfTheLastOpenAccountGivesNoOpenAccountAndCallsTheHook(t *testing.T) {
	calls := &hookCalls{}
	store := customerstore.New(newPool(t), calls.hook, nil)
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("222222")))

	noOpenAccount, err := store.CloseAccount(t.Context(), account("111111"), closedAt, nil)
	require.NoError(t, err)
	assert.False(t, noOpenAccount)
	assert.Empty(t, calls.customers)

	noOpenAccount, err = store.CloseAccount(t.Context(), account("222222"), closedAt, nil)
	require.NoError(t, err)
	assert.True(t, noOpenAccount)
	assert.Equal(t, []string{"00000001"}, calls.customers)
	assert.Equal(t, []time.Time{closedAt}, calls.at)
}

func TestAnErrorOfTheHookUndoesTheClose(t *testing.T) {
	calls := &hookCalls{err: errors.New("insert failed")}
	store := customerstore.New(newPool(t), calls.hook, nil)
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))

	_, err := store.CloseAccount(t.Context(), account("111111"), closedAt, nil)

	require.ErrorIs(t, err, calls.err)

	record, err := store.Customer(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Equal(t, domain.OpenStatusOpen, record.Accounts[0].OpenStatus)
}

func TestAClosedAccountDoesNotCloseAgain(t *testing.T) {
	calls := &hookCalls{}
	store := customerstore.New(newPool(t), calls.hook, nil)
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))
	_, err := store.CloseAccount(t.Context(), account("111111"), closedAt, nil)
	require.NoError(t, err)

	_, err = store.CloseAccount(t.Context(), account("111111"), closedAt, nil)

	require.ErrorIs(t, err, domain.ErrAccountClosed)
	assert.Len(t, calls.customers, 1)
}

func TestTwoClosesAtTheSameTimeOfTheLastTwoOpenAccountsGiveNoOpenAccountOneTime(t *testing.T) {
	calls := &hookCalls{}
	store := customerstore.New(newPool(t), calls.hook, nil)

	for _, number := range []string{"00000001", "00000002", "00000003", "00000004", "00000005"} {
		require.NoError(t, store.AddCustomer(t.Context(), customer(number)))
		first, second := account(number+"1"), account(number+"2")
		require.NoError(t, store.AddAccount(t.Context(), number, first))
		require.NoError(t, store.AddAccount(t.Context(), number, second))

		var (
			closes        sync.WaitGroup
			noOpenAccount [2]bool
			errs          [2]error
		)

		for i, closing := range []domain.Account{first, second} {
			closes.Go(func() { noOpenAccount[i], errs[i] = store.CloseAccount(t.Context(), closing, closedAt, nil) })
		}

		closes.Wait()
		require.NoError(t, errors.Join(errs[:]...))
		assert.ElementsMatch(t, []bool{true, false}, noOpenAccount[:], number)
	}

	assert.Len(t, calls.customers, 5)
}

var errLastClose = errors.New("the check refuses the close of the last open account")

func refuseLastClose(_ context.Context, noOpenAccount bool) error {
	if noOpenAccount {
		return errLastClose
	}

	return nil
}

func TestARefusalOfTheCheckUndoesTheCloseAndCallsNoHook(t *testing.T) {
	calls := &hookCalls{}
	store := customerstore.New(newPool(t), calls.hook, nil)
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))

	_, err := store.CloseAccount(t.Context(), account("111111"), closedAt, refuseLastClose)

	require.ErrorIs(t, err, errLastClose)
	assert.Empty(t, calls.customers)

	record, err := store.Customer(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Equal(t, domain.OpenStatusOpen, record.Accounts[0].OpenStatus)
}

func TestTheCheckOfOneOfTwoClosesAtTheSameTimeSeesTheOtherClose(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)

	for _, number := range []string{"00000001", "00000002", "00000003", "00000004", "00000005"} {
		require.NoError(t, store.AddCustomer(t.Context(), customer(number)))
		first, second := account(number+"1"), account(number+"2")
		require.NoError(t, store.AddAccount(t.Context(), number, first))
		require.NoError(t, store.AddAccount(t.Context(), number, second))

		var (
			closes sync.WaitGroup
			errs   [2]error
		)

		for i, closing := range []domain.Account{first, second} {
			closes.Go(func() { _, errs[i] = store.CloseAccount(t.Context(), closing, closedAt, refuseLastClose) })
		}

		closes.Wait()

		refused := 0

		for _, err := range errs {
			if errors.Is(err, errLastClose) {
				refused++
			} else {
				require.NoError(t, err, number)
			}
		}

		assert.Equal(t, 1, refused, "the check of the second close sees the first close: %s", number)
	}
}

// inserts keeps each transaction that the inserter gets, and sees the account in the same database transaction.
type inserts struct {
	accountIDs   []string
	transactions []domain.Transaction
	err          error
}

func (i *inserts) insert(ctx context.Context, tx pgx.Tx, accountID string, transaction domain.Transaction) error {
	var found string
	if err := tx.QueryRow(ctx, "SELECT account_id FROM accounts WHERE account_id = $1", accountID).
		Scan(&found); err != nil {
		return err
	}

	i.accountIDs, i.transactions = append(i.accountIDs, accountID), append(i.transactions, transaction)

	return i.err
}

func drawdown() domain.Transaction {
	return domain.Drawdown(2_000_000, closedAt)
}

func TestAnAccountWithATransactionAddsTheAccountAndTheTransactionInOneDatabaseTransaction(t *testing.T) {
	calls := &inserts{}
	store := customerstore.New(newPool(t), nil, calls.insert)
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))

	require.NoError(t, store.AddAccountWithTransaction(t.Context(), "00000001", account("111111"), drawdown()))

	record, err := store.Customer(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Equal(t, []domain.Account{account("111111")}, record.Accounts)
	assert.Equal(t, []string{"account-111111"}, calls.accountIDs)
	assert.Equal(t, []domain.Transaction{drawdown()}, calls.transactions)
}

func TestAnErrorOfTheInserterUndoesTheAccount(t *testing.T) {
	calls := &inserts{err: errors.New("insert failed")}
	store := customerstore.New(newPool(t), nil, calls.insert)
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))

	err := store.AddAccountWithTransaction(t.Context(), "00000001", account("111111"), drawdown())

	require.ErrorIs(t, err, calls.err)

	record, err := store.Customer(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Empty(t, record.Accounts)
}

func TestAnAccountWithATransactionGivesTheErrorsOfAnAccount(t *testing.T) {
	calls := &inserts{}
	store := customerstore.New(newPool(t), nil, calls.insert)
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("111111")))

	sameNumber := account("111111")
	sameNumber.AccountID = "account-other"

	sameNumberErr := store.AddAccountWithTransaction(t.Context(), "00000001", sameNumber, drawdown())
	noCustomerErr := store.AddAccountWithTransaction(t.Context(), "00000009", account("222222"), drawdown())

	require.ErrorIs(t, sameNumberErr, app.ErrAccountExists)
	require.ErrorIs(t, noCustomerErr, app.ErrCustomerNotFound)
	assert.Empty(t, calls.transactions)
}

// Verifies: DH-2665
func TestAWriteInATransactionThatGivesAnErrorLeavesNothing(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	errStop := errors.New("stop")

	err := store.InTransaction(t.Context(), func(inTx *customerstore.Store, _ pgx.Tx) error {
		require.NoError(t, inTx.AddCustomer(t.Context(), customer("00000001")))
		require.NoError(t, inTx.AddAccount(t.Context(), "00000001", account("111111")))

		_, err := inTx.Customer(t.Context(), "00000001")
		require.NoError(t, err, "the transaction sees its own writes")

		return errStop
	})

	require.ErrorIs(t, err, errStop)

	_, err = store.Customer(t.Context(), "00000001")
	assert.ErrorIs(t, err, app.ErrCustomerNotFound)
}

// Verifies: DH-1201
func TestACustomerNumberInUseInATransactionLetsTheTransactionTryADifferentNumber(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	require.NoError(t, store.AddCustomer(t.Context(), customer("00000001")))

	err := store.InTransaction(t.Context(), func(inTx *customerstore.Store, _ pgx.Tx) error {
		used := customer("00000001")
		used.EmailAddress = "other@example.com"
		require.ErrorIs(t, inTx.AddCustomer(t.Context(), used), app.ErrCustomerNumberUsed)

		used.CustomerNumber = "00000002"

		return inTx.AddCustomer(t.Context(), used)
	})

	require.NoError(t, err)

	_, err = store.Customer(t.Context(), "00000002")
	assert.NoError(t, err)
}
