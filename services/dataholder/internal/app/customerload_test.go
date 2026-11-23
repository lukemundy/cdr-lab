package app_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// productsByID gives each product of the brand by its ID, also a product that is not current.
type productsByID map[string]domain.Product

func (p productsByID) Products(context.Context, app.ProductQuery) ([]domain.Product, int, error) {
	return nil, 0, nil
}

func (p productsByID) Product(_ context.Context, productID string) (domain.Product, error) {
	product, ok := p[productID]
	if !ok {
		return domain.Product{}, app.ErrProductNotFound
	}

	return product, nil
}

// loadWrites counts the writes, and writes into the fakes with no transaction.
type loadWrites struct {
	store        *fakeCustomerStore
	transactions *memoryTransactions
	count        int
}

func (l *loadWrites) write(_ context.Context, write func(app.CustomerStore, app.TransactionStore) error) error {
	l.count++

	return write(l.store, l.transactions)
}

func sequentialIDs() func() string {
	next := 0

	return func() string {
		next++

		return "id-" + string(rune('0'+next))
	}
}

func loader(writes *loadWrites, random func() uint64) app.Customers {
	return app.Customers{
		Arrangements: &endings{}, Random: random, NewID: sequentialIDs(), Now: now, BSB: "372100", Writes: writes.write,
		Products: productsByID{"saver": saver()},
	}
}

func newLoadWrites() *loadWrites {
	return &loadWrites{store: newFakeCustomerStore(), transactions: &memoryTransactions{}}
}

func loadFile(t *testing.T, text string) domain.CustomerLoad {
	t.Helper()

	var load domain.CustomerLoad
	require.NoError(t, json.Unmarshal([]byte(text), &load))

	return load
}

const oneCustomer = `{"customers": [{"firstName": "Ada", "lastName": "Lovelace", "emailAddress": "ada@mail.test",
	"accounts": [
		{"productId": "saver", "accountNumber": "11112222", "transactions": [
			{"type": "PAYMENT", "status": "POSTED", "description": "Rent", "amount": "-12.50", "age": "P3DT2H"},
			{"type": "PAYMENT", "status": "PENDING", "description": "Food", "amount": "-1.00", "age": "PT30M"}
		]},
		{"productId": "saver", "accountNumber": "11113333", "displayName": "Old", "openStatus": "CLOSED",
			"transactions": [
				{"type": "FEE", "status": "POSTED", "description": "Fee", "amount": "-2.00", "age": "P10D"}
			]}
	]}]}`

// Verifies: DH-2653, DH-2655, DH-2656, DH-2658, DH-2660, DH-2662
func TestALoadCreatesEachCustomerWithItsAccountsAndTransactionsInTheSequenceOfTheFile(t *testing.T) {
	writes := newLoadWrites()

	results, err := loader(writes, random(1)).Load(t.Context(), loadFile(t, oneCustomer))

	require.NoError(t, err)
	assert.Equal(t, []app.LoadResult{{CustomerNumber: "00000001", Created: true}}, results)
	assert.Equal(t, 1, writes.count, "the load writes in one transaction")

	stored := writes.store.customers["00000001"]
	assert.Equal(t, "ada@mail.test", stored.EmailAddress)

	accounts := writes.store.accounts["00000001"]
	require.Len(t, accounts, 2)
	assert.Equal(t, []string{"11112222", "11113333"}, []string{accounts[0].AccountNumber, accounts[1].AccountNumber})
	assert.Equal(t, "Bonus Saver", accounts[0].DisplayName)
	assert.Equal(t, "Old", accounts[1].DisplayName)

	for _, a := range accounts {
		assert.Equal(t, "Bonus Saver", a.ProductName)
		assert.Equal(t, domain.CategoryTransAndSavingsAccounts, a.ProductCategory)
		assert.Equal(t, "372100", a.BSB)
	}

	assert.Equal(t, domain.OpenStatusOpen, accounts[0].OpenStatus)
	assert.Equal(t, domain.OpenStatusClosed, accounts[1].OpenStatus)

	first := writes.transactions.accounts[accounts[0].AccountID]
	require.Len(t, first, 2)
	assert.Equal(t, []string{"Rent", "Food"}, []string{first[0].Description, first[1].Description})
	assert.Equal(t, now().UTC().Add(-74*time.Hour), *first[0].PostingDateTime)
	assert.Equal(t, now().UTC().Add(-30*time.Minute), *first[1].ExecutionDateTime)
	assert.Len(t, writes.transactions.accounts[accounts[1].AccountID], 1, "the close is after the transactions")
}

// Verifies: DH-2661, DH-2662
func TestALoadChangesNothingForACustomerWhoseEmailAddressTheBrandHas(t *testing.T) {
	writes := newLoadWrites()
	writes.store.customers["00000009"] = domain.Customer{
		CustomerNumber: "00000009", LastName: "Other Name", EmailAddress: "ADA@Mail.test",
	}

	results, err := loader(writes, random(1)).Load(t.Context(), loadFile(t, oneCustomer))

	require.NoError(t, err)
	assert.Equal(t, []app.LoadResult{{CustomerNumber: "00000009", Created: false}}, results)
	assert.Len(t, writes.store.customers, 1)
	assert.Empty(t, writes.store.accounts)
	assert.Empty(t, writes.transactions.accounts)
}

// Verifies: DH-2659, DH-2665
func TestALoadWithAnInvalidRecordGivesTheErrorsOfAllRecordsAtTheirPositionsAndWritesNothing(t *testing.T) {
	writes := newLoadWrites()
	// The existing customer has invalid records too, so that its records are also checked.
	writes.store.customers["00000009"] = domain.Customer{CustomerNumber: "00000009", EmailAddress: "ada@mail.test"}

	_, err := loader(writes, random(1)).Load(t.Context(), loadFile(t, `{"customers": [
		{"lastName": "Lovelace", "emailAddress": "ada@mail.test", "accounts": [
			{"productId": "saver", "accountNumber": "12345"}]},
		{"lastName": "", "emailAddress": "grace@mail.test", "accounts": [
			{"productId": "saver", "accountNumber": "22223333", "transactions": [
				{"type": "FEE", "status": "POSTED", "description": "Fee", "amount": "-2.00", "age": "P1W"}]}]}
	]}`))

	assert.ElementsMatch(t, []string{
		"customers[0].accounts[0].accountNumber", "customers[1].lastName",
		"customers[1].accounts[0].transactions[0].age",
	}, fieldsOf(t, err))
	assert.Contains(t, err.Error(), "customers[1].accounts[0].transactions[0]: age must be")
	assert.Zero(t, writes.count)
}

// Verifies: DH-2654
func TestALoadAccountOfAProductThatTheBrandDoesNotHaveIsAnErrorAtItsPosition(t *testing.T) {
	writes := newLoadWrites()

	_, err := loader(writes, random(1)).Load(t.Context(), loadFile(t, `{"customers": [
		{"lastName": "Lovelace", "emailAddress": "ada@mail.test", "accounts": [
			{"productId": "saver", "accountNumber": "11112222"},
			{"productId": "wombat-gold", "accountNumber": "11113333"}]}]}`))

	assert.Equal(t, []string{"customers[0].accounts[1].productId"}, fieldsOf(t, err))
	assert.Equal(t, "customers[0].accounts[1]: the data holder has no product with this productId", err.Error())
	assert.Zero(t, writes.count)
}

// Verifies: DH-2663
func TestALoadWithTwoCustomersOfTheSameEmailAddressIsNotValid(t *testing.T) {
	writes := newLoadWrites()

	_, err := loader(writes, random(1)).Load(t.Context(), loadFile(t, `{"customers": [
		{"lastName": "Lovelace", "emailAddress": "ada@mail.test"},
		{"lastName": "Hopper", "emailAddress": "grace@mail.test"},
		{"lastName": "Byron", "emailAddress": "ADA@mail.TEST"}]}`))

	assert.Equal(t, []string{"customers[2].emailAddress"}, fieldsOf(t, err))
	assert.Equal(t, "customers[2]: emailAddress is also the emailAddress of customers[0]", err.Error())
	assert.Zero(t, writes.count)
}

// Verifies: DH-2650
func TestALoadFileMustHaveCustomers(t *testing.T) {
	writes := newLoadWrites()

	_, err := loader(writes, random(1)).Load(t.Context(), loadFile(t, `{}`))

	assert.Equal(t, []string{"customers"}, fieldsOf(t, err))

	results, err := loader(writes, random(1)).Load(t.Context(), loadFile(t, `{"customers": []}`))

	require.NoError(t, err)
	assert.Equal(t, []app.LoadResult{}, results)
}

// Verifies: DH-2664
func TestALoadAccountThatACreateRefusesIsAnErrorAtItsPosition(t *testing.T) {
	writes := newLoadWrites()
	customers := loader(writes, random(1))
	customers.Writes = func(ctx context.Context, write func(app.CustomerStore, app.TransactionStore) error) error {
		return write(numbersInUse{writes.store, map[string]bool{"11113333": true}}, writes.transactions)
	}

	_, err := customers.Load(t.Context(), loadFile(t, oneCustomer))

	var record app.LoadRecordError
	require.ErrorAs(t, err, &record)
	assert.Equal(t, "customers[0].accounts[1]", record.Position)
	assert.ErrorIs(t, err, app.ErrAccountExists)
}
