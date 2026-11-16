package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type fakeCustomerStore struct {
	customers map[string]domain.Customer
	accounts  map[string][]domain.Account
	// transactions are by account ID.
	transactions map[string][]domain.Transaction
	closedAt     time.Time
	err          error
}

func newFakeCustomerStore() *fakeCustomerStore {
	return &fakeCustomerStore{customers: map[string]domain.Customer{}, accounts: map[string][]domain.Account{}}
}

func (f *fakeCustomerStore) AddCustomer(ctx context.Context, customer domain.Customer) error {
	if f.err != nil {
		return f.err
	}

	if _, ok := f.customers[customer.CustomerNumber]; ok {
		return app.ErrCustomerNumberUsed
	}

	if _, err := f.CustomerNumberOf(ctx, customer.EmailAddress); err == nil {
		return app.ErrEmailAddressUsed
	}

	f.customers[customer.CustomerNumber] = customer

	return nil
}

func (f *fakeCustomerStore) AddAccount(_ context.Context, customerNumber string, account domain.Account) error {
	if f.err != nil {
		return f.err
	}

	f.accounts[customerNumber] = append(f.accounts[customerNumber], account)

	return nil
}

func (f *fakeCustomerStore) AddAccountWithTransaction(
	ctx context.Context,
	customerNumber string,
	account domain.Account,
	transaction domain.Transaction,
) error {
	if err := f.AddAccount(ctx, customerNumber, account); err != nil {
		return err
	}

	if f.transactions == nil {
		f.transactions = map[string][]domain.Transaction{}
	}

	f.transactions[account.AccountID] = append(f.transactions[account.AccountID], transaction)

	return nil
}

func (f *fakeCustomerStore) CloseAccount(
	ctx context.Context,
	account domain.Account,
	closedAt time.Time,
	check app.CloseCheck,
) (bool, error) {
	if f.err != nil {
		return false, f.err
	}

	f.closedAt = closedAt

	for number, accounts := range f.accounts {
		for i := range accounts {
			if accounts[i].AccountID != account.AccountID {
				continue
			}

			if accounts[i].OpenStatus == domain.OpenStatusClosed {
				return false, domain.ErrAccountClosed
			}

			noOpenAccount := len(domain.OpenAccounts(f.accounts[number])) == 1
			if check != nil {
				if err := check(ctx, noOpenAccount); err != nil {
					return false, err
				}
			}

			f.accounts[number][i].OpenStatus = domain.OpenStatusClosed

			return noOpenAccount, nil
		}
	}

	return false, nil
}

func (f *fakeCustomerStore) Customer(_ context.Context, customerNumber string) (app.CustomerRecord, error) {
	customer, ok := f.customers[customerNumber]
	if !ok {
		return app.CustomerRecord{}, app.ErrCustomerNotFound
	}

	return app.CustomerRecord{Customer: customer, Accounts: f.accounts[customerNumber]}, nil
}

func (f *fakeCustomerStore) CustomerNumberOf(_ context.Context, emailAddress string) (string, error) {
	for number, customer := range f.customers {
		if strings.EqualFold(customer.EmailAddress, emailAddress) {
			return number, nil
		}
	}

	return "", app.ErrCustomerNotFound
}

// otherCustomer has a different email address from customer.
func otherCustomer() domain.Customer {
	return domain.Customer{LastName: "Other", EmailAddress: "sam@example.com"}
}

func random(values ...uint64) func() uint64 {
	return func() uint64 {
		next := values[0]
		values = values[1:]

		return next
	}
}

func customer() domain.Customer {
	return domain.Customer{LastName: "Example", EmailAddress: "alex@example.com"}
}

func account() domain.Account {
	return domain.Account{
		DisplayName:     "Everyday",
		ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName:     "Mock Everyday Account",
		BSB:             "123456",
		AccountNumber:   "12345678",
	}
}

// now is a time with a time zone that is not UTC.
func now() time.Time {
	return time.Date(2026, 10, 15, 21, 30, 0, 0, time.FixedZone("AWST", 8*60*60))
}

// Verifies: DH-1201
func TestANewCustomerGetsACustomerNumberAndTheTimeOfItsLastChange(t *testing.T) {
	store := newFakeCustomerStore()
	customers := app.Customers{Store: store, Now: now, Random: random(1)}

	created, err := customers.Create(t.Context(), customer())

	require.NoError(t, err)
	assert.Equal(t, "00000001", created.CustomerNumber)
	assert.Equal(t, time.Date(2026, 10, 15, 13, 30, 0, 0, time.UTC), created.LastUpdated)
	assert.Equal(t, created, store.customers["00000001"])
}

func TestACustomerNumberOfTheRequestHasNoEffect(t *testing.T) {
	requested := customer()
	requested.CustomerNumber = "99999999"

	created, err := app.Customers{Store: newFakeCustomerStore(), Now: now, Random: random(1)}.
		Create(t.Context(), requested)

	require.NoError(t, err)
	assert.Equal(t, "00000001", created.CustomerNumber)
}

func TestACustomerNumberThatIsInUseIsReplaced(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = otherCustomer()
	customers := app.Customers{Store: store, Now: now, Random: random(1, 2)}

	created, err := customers.Create(t.Context(), customer())

	require.NoError(t, err)
	assert.Equal(t, "00000002", created.CustomerNumber)
}

func TestACustomerIsNotCreatedWhenEachNumberIsInUse(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = otherCustomer()
	customers := app.Customers{Store: store, Now: now, Random: random(1, 1, 1)}

	_, err := customers.Create(t.Context(), customer())

	require.ErrorIs(t, err, app.ErrNoFreeCustomerNumber)
	require.NotErrorIs(t, err, app.ErrCustomerNumberUsed)
	assert.Len(t, store.customers, 1)
}

func TestACustomerWithAnEmailAddressInUseIsNotCreatedAndTriesNoOtherNumber(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	inUse := otherCustomer()
	inUse.EmailAddress = "ALEX@example.com"
	customers := app.Customers{Store: store, Now: now, Random: random(2)}

	_, err := customers.Create(t.Context(), inUse)

	require.ErrorIs(t, err, app.ErrEmailAddressUsed)
	assert.Len(t, store.customers, 1)
}

// Verifies: DH-1200, DH-1203, DH-1206
func TestACustomerIsIdentifiedByItsCustomerNumberOrItsEmailAddressInAnyCase(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	store.customers["00000002"] = otherCustomer()
	customers := app.Customers{Store: store}

	for _, identifier := range []string{"00000001", " 00000001 ", "alex@example.com", "  Alex@Example.COM\t"} {
		record, err := customers.Identify(t.Context(), identifier)

		require.NoError(t, err, identifier)
		assert.Equal(t, customer(), record.Customer, identifier)
	}
}

func TestAnIdentifierThatNoCustomerHasGivesNotFound(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	customers := app.Customers{Store: store}

	for _, identifier := range []string{"00000002", "alex@example.org", "alex", ""} {
		_, err := customers.Identify(t.Context(), identifier)

		require.ErrorIs(t, err, app.ErrCustomerNotFound, identifier)
	}
}

func TestACustomerThatIsNotValidIsNotStored(t *testing.T) {
	store := newFakeCustomerStore()

	_, err := app.Customers{Store: store, Now: now, Random: random(1)}.
		Create(t.Context(), domain.Customer{})

	assert.Equal(t, []string{"lastName", "emailAddress"}, fieldsOf(t, err))
	assert.Empty(t, store.customers)
}

func TestAnAccountThatIsNotValidIsNotStored(t *testing.T) {
	store := newFakeCustomerStore()

	_, err := app.Customers{Store: store}.AddAccount(t.Context(), "00000001", domain.Account{})

	assert.Equal(
		t, []string{"displayName", "productCategory", "productName", "bsb", "accountNumber"}, fieldsOf(t, err),
	)
	assert.Empty(t, store.accounts)
}

func fieldsOf(t *testing.T, err error) []string {
	t.Helper()

	invalid, ok := errors.AsType[domain.ValidationError](err)
	require.True(t, ok, err)

	fields := make([]string, 0, len(invalid.Errors))
	for _, f := range invalid.Errors {
		fields = append(fields, f.Field)
	}

	return fields
}

func newID() string { return "account-1" }

func TestAValidAccountIsStoredForItsCustomerOpenWithANewIdentifierAndItsTimeOfCreation(t *testing.T) {
	store := newFakeCustomerStore()
	want := account()
	want.AccountID = "account-1"
	want.OpenStatus = domain.OpenStatusOpen
	want.CreatedAt = time.Date(2026, 10, 15, 13, 30, 0, 0, time.UTC)

	created, err := app.Customers{Store: store, NewID: newID, Now: now}.AddAccount(t.Context(), "00000001", account())

	require.NoError(t, err)
	assert.Equal(t, want, created)
	assert.Equal(t, []domain.Account{want}, store.accounts["00000001"])
}

func TestTheIdentifierAndTheStatusOfAnAccountRequestAreNotUsed(t *testing.T) {
	requested := account()
	requested.AccountID = "account-of-the-request"
	requested.OpenStatus = domain.OpenStatusClosed

	created, err := app.Customers{Store: newFakeCustomerStore(), NewID: newID, Now: now}.
		AddAccount(t.Context(), "00000001", requested)

	require.NoError(t, err)
	assert.Equal(t, "account-1", created.AccountID)
	assert.Equal(t, domain.OpenStatusOpen, created.OpenStatus)
}

func TestAnOpenAccountOfTheCustomerCloses(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	customers := app.Customers{Store: store, Arrangements: &endings{}, NewID: newID, Now: now}
	created, err := customers.AddAccount(t.Context(), "00000001", account())
	require.NoError(t, err)

	closed, err := customers.CloseAccount(t.Context(), "00000001", created.AccountID)

	require.NoError(t, err)
	assert.Equal(t, domain.OpenStatusClosed, closed.OpenStatus)
	assert.Equal(t, []domain.Account{closed}, store.accounts["00000001"])

	_, err = customers.CloseAccount(t.Context(), "00000001", created.AccountID)

	require.ErrorIs(t, err, domain.ErrAccountClosed)
}

// endings records each end of the arrangements of a customer who is not eligible.
type endings struct {
	// current are the arrangements that a loss of eligibility ends.
	current   []domain.ArrangementState
	customers []string
	at        []time.Time
	err       error
}

func (e *endings) IneligibilityEnds(
	context.Context,
	string,
	time.Time,
) ([]domain.ArrangementState, error) {
	return e.current, nil
}

func (e *endings) EndForIneligibleCustomer(_ context.Context, customerNumber string, at time.Time) error {
	e.customers, e.at = append(e.customers, customerNumber), append(e.at, at)

	return e.err
}

func TestTheCloseOfTheLastOpenAccountEndsTheArrangementsOfTheCustomerAtTheTimeOfTheClose(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	ends := &endings{}
	store.accounts["00000001"] = []domain.Account{openAccount("account-1"), openAccount("account-2")}
	customers := app.Customers{Store: store, Arrangements: ends, Now: now}

	_, err := customers.CloseAccount(t.Context(), "00000001", "account-1")
	require.NoError(t, err)
	assert.Empty(t, ends.customers, "the customer has a different open account")

	_, err = customers.CloseAccount(t.Context(), "00000001", "account-2")
	require.NoError(t, err)
	assert.Equal(t, []string{"00000001"}, ends.customers)
	assert.Equal(t, []time.Time{now().UTC()}, ends.at)
	assert.Equal(t, now().UTC(), store.closedAt)
}

func TestTheCloseOfTheLastOpenAccountSucceedsWhenTheEndOfTheArrangementsFails(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	ends := &endings{err: errors.New("database failure")}
	customers := app.Customers{Store: store, Arrangements: ends, NewID: newID, Now: now}
	created, err := customers.AddAccount(t.Context(), "00000001", account())
	require.NoError(t, err)

	closed, err := customers.CloseAccount(t.Context(), "00000001", created.AccountID)

	require.NoError(t, err, "the job of the close ends the arrangements later")
	assert.Equal(t, domain.OpenStatusClosed, closed.OpenStatus)
}

func TestAnAccountThatTheCustomerDoesNotHaveDoesNotClose(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	store.customers["00000002"] = customer()
	customers := app.Customers{Store: store, NewID: newID, Now: now}
	created, err := customers.AddAccount(t.Context(), "00000002", account())
	require.NoError(t, err)

	_, otherCustomer := customers.CloseAccount(t.Context(), "00000001", created.AccountID)
	_, unknownCustomer := customers.CloseAccount(t.Context(), "99999999", created.AccountID)

	require.ErrorIs(t, otherCustomer, app.ErrAccountNotFound)
	require.ErrorIs(t, unknownCustomer, app.ErrCustomerNotFound)
	assert.Equal(t, domain.OpenStatusOpen, store.accounts["00000002"][0].OpenStatus)
}

func TestAStoreErrorHasItsOperation(t *testing.T) {
	store := newFakeCustomerStore()
	store.err = errors.New("connection lost")
	customers := app.Customers{Store: store, Now: now, Random: random(1), NewID: newID}

	_, createErr := customers.Create(t.Context(), customer())
	_, accountErr := customers.AddAccount(t.Context(), "00000001", account())

	require.ErrorIs(t, createErr, store.err)
	require.ErrorContains(t, createErr, "add the customer")
	require.ErrorIs(t, accountErr, store.err)
	require.ErrorContains(t, accountErr, "add the account")
}

func current(id string) domain.ArrangementState {
	return domain.ArrangementState{Created: true, Arrangement: domain.Arrangement{ArrangementID: id}}
}

func TestACloseThatEndsTheArrangementsOfTheRequestCloses(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()
	store.accounts["00000001"] = []domain.Account{openAccount("account-1"), openAccount("account-2")}
	ends := &endings{current: []domain.ArrangementState{current("arrangement-1"), current("arrangement-2")}}
	customers := app.Customers{Store: store, Arrangements: ends, Now: now}

	_, err := customers.CloseAccountThatEnds(t.Context(), "00000001", "account-1", []string{})
	require.NoError(t, err, "a close that is not the last ends no arrangement")

	_, err = customers.CloseAccountThatEnds(
		t.Context(), "00000001", "account-2", []string{"arrangement-2", "arrangement-1"},
	)
	require.NoError(t, err, "the sequence of the IDs has no effect")
	assert.Equal(t, []string{"00000001"}, ends.customers)
}

func TestACloseThatEndsDifferentArrangementsFromTheRequestChangesNothing(t *testing.T) {
	tests := map[string]struct {
		accounts []domain.Account
		ends     []string
	}{
		"a close that a different close made the last": {
			accounts: []domain.Account{openAccount("account-1")},
			ends:     []string{},
		},
		"a last close with an arrangement that is not in the request": {
			accounts: []domain.Account{openAccount("account-1")},
			ends:     []string{"arrangement-1"},
		},
		"a close that is not the last with an arrangement in the request": {
			accounts: []domain.Account{openAccount("account-1"), openAccount("account-2")},
			ends:     []string{"arrangement-1"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store := newFakeCustomerStore()
			store.customers["00000001"] = customer()
			store.accounts["00000001"] = tc.accounts
			ends := &endings{current: []domain.ArrangementState{current("arrangement-1"), current("arrangement-2")}}
			customers := app.Customers{Store: store, Arrangements: ends, Now: now}

			_, err := customers.CloseAccountThatEnds(t.Context(), "00000001", "account-1", tc.ends)

			require.ErrorIs(t, err, app.ErrEndsChanged)
			assert.Equal(t, domain.OpenStatusOpen, store.accounts["00000001"][0].OpenStatus)
			assert.Empty(t, ends.customers)
		})
	}
}
