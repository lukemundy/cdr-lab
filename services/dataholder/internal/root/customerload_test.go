package root

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/productstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// All customer, account and transaction data in this file is invented. The brand of the test configuration has the
// bsb 376100.

// seedFile has a customer with an open account and a closed account, and a customer with no accounts. The last
// transaction is older than the default window of Get Transactions.
const seedFile = `{"customers": [
	{"firstName": "Ada", "lastName": "Lovelace", "emailAddress": "ada@mail.test", "accounts": [
		{"productId": "everyday", "accountNumber": "11112222", "creditLimit": "500.00", "transactions": [
			{"type": "PAYMENT", "status": "POSTED", "description": "Rent", "amount": "-12.50", "age": "P3DT2H"},
			{"type": "PAYMENT", "status": "PENDING", "description": "Food", "amount": "-1.00", "age": "PT30M"},
			{"type": "FEE", "status": "POSTED", "description": "Old fee", "amount": "-3.00", "age": "P91D"}
		]},
		{"productId": "everyday", "accountNumber": "11113333", "displayName": "Old Saver", "openStatus": "CLOSED",
			"transactions": [
				{"type": "TRANSFER_INCOMING", "status": "POSTED", "description": "Pay", "amount": "100.00",
					"age": "P10D"}
			]}
	]},
	{"lastName": "Hopper", "emailAddress": "grace@mail.test"}
]}`

func newLoadLab(t *testing.T) operatorLab {
	t.Helper()

	l := newOperatorLab(t)
	product := domain.Product{
		ProductID: "everyday", ProductCategory: domain.CategoryTransAndSavingsAccounts, Name: "Everyday Account",
		Description: "An account for each day.", Brand: "Coastal Bank", LastUpdated: time.Now().Add(-time.Hour),
	}
	require.NoError(t, productstore.New(l.pool).ImportProducts(t.Context(), []domain.Product{product}))

	return l
}

type loadReply struct {
	Customers []app.LoadResult `json:"customers"`
}

func (l operatorLab) load(t *testing.T, file string) operatorReply {
	t.Helper()

	return l.operate(t, http.MethodPost, "/v1/customer-loads", json.RawMessage(file))
}

func (l operatorLab) loaded(t *testing.T, file string) []app.LoadResult {
	t.Helper()

	reply := l.load(t, file)
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))

	var results loadReply
	reply.decode(t, &results)

	return results.Customers
}

type rowCounts struct{ customers, accounts, transactions int }

func (l operatorLab) rows(t *testing.T) rowCounts {
	t.Helper()

	var c rowCounts
	require.NoError(t, l.pool.QueryRow(t.Context(), `SELECT
		(SELECT count(*) FROM customers), (SELECT count(*) FROM accounts), (SELECT count(*) FROM transactions)`,
	).Scan(&c.customers, &c.accounts, &c.transactions))

	return c
}

func (l operatorLab) transactionsOf(t *testing.T, customerNumber, accountID string) []operatorTransaction {
	t.Helper()

	reply := l.operate(t, http.MethodGet, transactionsPath(customerNumber, accountID), nil)
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))

	var transactions []operatorTransaction
	reply.decode(t, &transactions)

	return transactions
}

// Verifies: DH-2649, DH-2653, DH-2655, DH-2656, DH-2660, DH-2662
func TestALoadCreatesTheCustomersWithTheAccountsOfTheProductsOfTheBrandAndTheirTransactions(t *testing.T) {
	l := newLoadLab(t)

	results := l.loaded(t, seedFile)

	require.Len(t, results, 2)
	assert.True(t, results[0].Created)
	assert.True(t, results[1].Created)

	var ada operatorCustomer
	l.customer(t, results[0].CustomerNumber).decode(t, &ada)
	assert.Equal(t, "Ada", ada.FirstName)
	assert.Equal(t, "ada@mail.test", ada.EmailAddress)
	require.Len(t, ada.Accounts, 2)

	open, closed := ada.Accounts[0], ada.Accounts[1]
	assert.Equal(t, operatorAccount{
		AccountID: open.AccountID, DisplayName: "Everyday Account", ProductCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
		ProductName: "Everyday Account", BSB: "376100", AccountNumber: "11112222", CreditLimit: "500.00",
		OpenStatus: "OPEN",
	}, open)
	assert.Equal(t, "Old Saver", closed.DisplayName)
	assert.Equal(t, "CLOSED", closed.OpenStatus)

	descriptions := []string{}
	for _, transaction := range l.transactionsOf(t, ada.CustomerNumber, open.AccountID) {
		descriptions = append(descriptions, transaction.Description)
	}

	assert.Equal(t, []string{"Rent", "Food", "Old fee"}, descriptions)
	assert.Len(t, l.transactionsOf(t, ada.CustomerNumber, closed.AccountID), 1, "the close is after the transactions")

	var grace operatorCustomer
	l.customer(t, results[1].CustomerNumber).decode(t, &grace)
	assert.Equal(t, "Hopper", grace.LastName)
}

// Verifies: DH-2658
func TestTheTransactionsOfALoadWithAnAgeInTheDefaultWindowAreInGetTransactions(t *testing.T) {
	l := newLoadLab(t)
	results := l.loaded(t, seedFile)

	customers := customerstore.New(l.pool, nil, nil)
	record, err := customers.Customer(t.Context(), results[0].CustomerNumber)
	require.NoError(t, err)

	data := app.CustomerData{
		Customers: customers, TransactionStore: transactionstore.New(l.pool),
		Arrangements: arrangementstore.NewQueries(l.pool), Now: time.Now,
	}

	found, total, err := data.Transactions(t.Context(), results[0].CustomerNumber, record.Accounts[0].AccountID, 1, 25)

	require.NoError(t, err)
	assert.Equal(t, 2, total, "the transaction with an age of 91 days is outside the default window")

	byDescription := map[string]domain.Transaction{}
	for _, transaction := range found {
		byDescription[transaction.Description] = transaction
	}

	rent := byDescription["Rent"]
	require.NotNil(t, rent.PostingDateTime)
	assert.WithinDuration(t, time.Now().Add(-74*time.Hour), *rent.PostingDateTime, time.Minute)
	assert.Equal(t, rent.PostingDateTime, rent.ExecutionDateTime)

	food := byDescription["Food"]
	assert.Nil(t, food.PostingDateTime)
	require.NotNil(t, food.ExecutionDateTime)
	assert.WithinDuration(t, time.Now().Add(-30*time.Minute), *food.ExecutionDateTime, time.Minute)
}

// Verifies: DH-2666
func TestASecondLoadOfTheSameFileCreatesNothing(t *testing.T) {
	l := newLoadLab(t)
	first := l.loaded(t, seedFile)
	before := l.rows(t)

	second := l.loaded(t, seedFile)

	assert.Equal(t, before, l.rows(t))
	require.Len(t, second, 2)

	for i := range second {
		assert.Equal(t, app.LoadResult{CustomerNumber: first[i].CustomerNumber, Created: false}, second[i])
	}
}

// Verifies: DH-2661
func TestALoadChangesNothingForACustomerThatExists(t *testing.T) {
	l := newLoadLab(t)
	reply := l.operate(t, http.MethodPost, "/v1/customers", customerRequest{
		LastName: "Byron", EmailAddress: "ADA@Mail.Test",
	})
	require.Equal(t, http.StatusCreated, reply.status, string(reply.body))

	var existing operatorCustomer
	reply.decode(t, &existing)

	results := l.loaded(t, seedFile)

	assert.Equal(t, app.LoadResult{CustomerNumber: existing.CustomerNumber, Created: false}, results[0])
	assert.True(t, results[1].Created)

	var after operatorCustomer
	l.customer(t, existing.CustomerNumber).decode(t, &after)
	assert.Empty(t, after.Accounts, "the customer gets no account")
	after.Accounts = nil
	assert.Equal(t, existing, after, "the customer keeps its fields")
}

// Verifies: DH-2659, DH-2665
func TestALoadWithOneInvalidRecordCreatesNothing(t *testing.T) {
	l := newLoadLab(t)

	reply := l.load(t, `{"customers": [
		{"lastName": "Lovelace", "emailAddress": "ada@mail.test", "accounts": [
			{"productId": "everyday", "accountNumber": "11112222"}]},
		{"lastName": "Hopper", "emailAddress": "grace@mail.test", "accounts": [
			{"productId": "everyday", "accountNumber": "12345"}]}]}`)

	assert.Equal(t, "customers[1].accounts[0]: accountNumber must have 6 to 10 digits",
		reply.problemDetail(t, http.StatusBadRequest))
	assert.Equal(t, rowCounts{}, l.rows(t))
}

// Verifies: DH-2654, DH-2665
func TestALoadAccountOfAProductThatTheBrandDoesNotHaveCreatesNothing(t *testing.T) {
	l := newLoadLab(t)

	reply := l.load(t, `{"customers": [{"lastName": "Lovelace", "emailAddress": "ada@mail.test", "accounts": [
		{"productId": "everyday", "accountNumber": "11112222"},
		{"productId": "wombat-gold", "accountNumber": "11113333"}]}]}`)

	assert.Equal(t, "customers[0].accounts[1]: the data holder has no product with this productId",
		reply.problemDetail(t, http.StatusBadRequest))
	assert.Equal(t, rowCounts{}, l.rows(t))
}

// The conflict is in the second customer, so that the load undoes the writes of the first customer.
//
// Verifies: DH-2664, DH-2665
func TestALoadAccountWithTheBSBAndAccountNumberOfADifferentAccountGives409AndCreatesNothing(t *testing.T) {
	l := newLoadLab(t)
	other := l.newCustomer(t)
	inUse := everydayAccount("55556666")
	inUse.BSB = "376100"
	l.createAccount(t, other.CustomerNumber, inUse)

	before := l.rows(t)

	for name, accountNumber := range map[string]string{"of the brand": "55556666", "of the file": "11112222"} {
		t.Run(name, func(t *testing.T) {
			reply := l.load(t, `{"customers": [
				{"lastName": "Lovelace", "emailAddress": "ada@mail.test", "accounts": [
					{"productId": "everyday", "accountNumber": "11112222", "transactions": [
						{"type": "FEE", "status": "POSTED", "description": "Fee", "amount": "-1.00", "age": "P1D"}]}]},
				{"lastName": "Hopper", "emailAddress": "grace@mail.test", "accounts": [
					{"productId": "everyday", "accountNumber": "`+accountNumber+`"}]}]}`)

			assert.Equal(t, "customers[1].accounts[0]: An account has this bsb and accountNumber.",
				reply.problemDetail(t, http.StatusConflict))
			assert.Equal(t, before, l.rows(t))
		})
	}
}

// Verifies: DH-2663
func TestALoadWithTwoCustomersOfTheSameEmailAddressGives400(t *testing.T) {
	l := newLoadLab(t)

	reply := l.load(t, `{"customers": [{"lastName": "Lovelace", "emailAddress": "ada@mail.test"},
		{"lastName": "Byron", "emailAddress": "Ada@Mail.Test"}]}`)

	assert.Equal(t, "customers[1]: emailAddress is also the emailAddress of customers[0]",
		reply.problemDetail(t, http.StatusBadRequest))
	assert.Equal(t, rowCounts{}, l.rows(t))
}

// sameCustomer has the values of a customer that do not depend on the data holder: no customer number, account ID,
// transaction ID or absolute time.
type sameCustomer struct {
	FirstName, LastName, EmailAddress string
	Accounts                          map[string]sameAccount
}

type sameAccount struct {
	DisplayName, ProductCategory, ProductName, CreditLimit, OpenStatus string
	Transactions                                                       []string
}

func (l operatorLab) sameCustomers(t *testing.T, results []app.LoadResult) []sameCustomer {
	t.Helper()

	customers := []sameCustomer{}

	for _, result := range results {
		var c operatorCustomer
		l.customer(t, result.CustomerNumber).decode(t, &c)

		same := sameCustomer{
			FirstName: c.FirstName, LastName: c.LastName, EmailAddress: c.EmailAddress,
			Accounts: map[string]sameAccount{},
		}

		for _, a := range c.Accounts {
			account := sameAccount{
				DisplayName: a.DisplayName, ProductCategory: a.ProductCategory, ProductName: a.ProductName,
				CreditLimit: a.CreditLimit, OpenStatus: a.OpenStatus,
			}
			for _, transaction := range l.transactionsOf(t, c.CustomerNumber, a.AccountID) {
				account.Transactions = append(account.Transactions,
					transaction.Type+" "+transaction.Status+" "+transaction.Description+" "+transaction.Amount)
			}

			same.Accounts[a.BSB+" "+a.AccountNumber] = account
		}

		customers = append(customers, same)
	}

	return customers
}

// Verifies: DH-2668
func TestTwoLoadsOfTheSameFileIntoTwoNewDataHoldersGiveTheSameCustomers(t *testing.T) {
	first, second := newLoadLab(t), newLoadLab(t)

	assert.Equal(
		t,
		first.sameCustomers(t, first.loaded(t, seedFile)),
		second.sameCustomers(t, second.loaded(t, seedFile)),
	)
}
