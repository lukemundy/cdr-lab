package operatorapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/operatorapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	token          = "test-operator-token"
	customerNumber = "00000001"
	customerBody   = `{"firstName":"Alex","lastName":"Example","emailAddress":"alex@example.com"}`
	accountBody    = `{"displayName":"Everyday","productCategory":"TRANS_AND_SAVINGS_ACCOUNTS",` +
		`"productName":"Mock Everyday Account","bsb":"123456","accountNumber":"12345678"}`
	accountID   = "5a1bf696-ee03-408b-b315-97955415d1f0"
	accountJSON = `{"accountId":"` + accountID + `",` + `"displayName":"Everyday",` +
		`"productCategory":"TRANS_AND_SAVINGS_ACCOUNTS","productName":"Mock Everyday Account","bsb":"123456",` +
		`"accountNumber":"12345678","creditLimit":"0.00","openStatus":"OPEN"}`
)

// httptest gives example.com as the host of a request.
var access = operatorhttp.Config{Token: token, AllowedHosts: []string{"example.com"}}

type fakeCustomers struct {
	created        *domain.Customer
	accountOwner   string
	account        *domain.Account
	readNumber     string
	record         app.CustomerRecord
	err            error
	customerNumber string
	closed         [2]string
	transaction    *domain.Transaction
	transactionsOf [2]string
	load           *domain.CustomerLoad
	loadResults    []app.LoadResult
}

func (f *fakeCustomers) Load(_ context.Context, load domain.CustomerLoad) ([]app.LoadResult, error) {
	f.load = &load

	return f.loadResults, f.err
}

func (f *fakeCustomers) Create(_ context.Context, customer domain.Customer) (domain.Customer, error) {
	if f.err != nil {
		return domain.Customer{}, f.err
	}

	requested := customer
	f.created = &requested
	customer.CustomerNumber = f.customerNumber

	return customer, nil
}

func (f *fakeCustomers) AddAccount(
	_ context.Context,
	owner string,
	account domain.Account,
) (domain.Account, error) {
	if f.err != nil {
		return domain.Account{}, f.err
	}

	requested := account
	f.accountOwner, f.account = owner, &requested
	account.AccountID = accountID
	account.OpenStatus = domain.OpenStatusOpen

	return account, nil
}

func (f *fakeCustomers) AddTransaction(
	_ context.Context,
	owner, id string,
	transaction domain.Transaction,
) (domain.Transaction, error) {
	if f.err != nil {
		return domain.Transaction{}, f.err
	}

	requested := transaction
	f.transactionsOf, f.transaction = [2]string{owner, id}, &requested
	transaction.TransactionID = "transaction-1"

	return transaction, nil
}

func (f *fakeCustomers) AccountTransactions(_ context.Context, owner, id string) ([]domain.Transaction, error) {
	f.transactionsOf = [2]string{owner, id}

	return nil, f.err
}

func (f *fakeCustomers) CloseAccount(_ context.Context, owner, id string) (domain.Account, error) {
	if f.err != nil {
		return domain.Account{}, f.err
	}

	f.closed = [2]string{owner, id}

	return domain.Account{
		AccountID:       id,
		DisplayName:     "Everyday",
		ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName:     "Mock Everyday Account",
		BSB:             "123456",
		AccountNumber:   "12345678",
		OpenStatus:      domain.OpenStatusClosed,
	}, nil
}

func (f *fakeCustomers) Customer(_ context.Context, number string) (app.CustomerRecord, error) {
	f.readNumber = number

	return f.record, f.err
}

type request struct {
	method, path, body string
	header             http.Header
}

func serve(t *testing.T, customers *fakeCustomers, req request) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	operatorapi.Register(mux, customers, &fakeArrangements{}, access)

	httpReq := httptest.NewRequestWithContext(t.Context(), req.method, req.path, strings.NewReader(req.body))
	httpReq.Header = req.header

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httpReq)

	return rec
}

func authorised() http.Header {
	return http.Header{"Authorization": {"Bearer " + token}}
}

func createCustomer(body string) request {
	return request{method: http.MethodPost, path: "/v1/customers", body: body, header: authorised()}
}

func createAccount(body string) request {
	return request{
		method: http.MethodPost,
		path:   "/v1/customers/" + customerNumber + "/accounts",
		body:   body,
		header: authorised(),
	}
}

func closeAccount() request {
	return request{
		method: http.MethodPost,
		path:   "/v1/customers/" + customerNumber + "/accounts/" + accountID + "/close",
		header: authorised(),
	}
}

func readCustomer() request {
	return request{method: http.MethodGet, path: "/v1/customers/" + customerNumber, header: authorised()}
}

func requireProblem(t *testing.T, rec *httptest.ResponseRecorder, status int) string {
	t.Helper()
	require.Equal(t, status, rec.Code, rec.Body.String())
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var problem struct {
		Detail string `json:"detail"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem))
	require.NotEmpty(t, problem.Detail)

	return problem.Detail
}

// Verifies: DH-0720
func TestACustomerIsCreatedAndReturnedWithItsCustomerNumber(t *testing.T) {
	customers := &fakeCustomers{customerNumber: customerNumber}

	rec := serve(t, customers, createCustomer(customerBody))

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{"customerNumber":"00000001","firstName":"Alex","lastName":"Example","emailAddress":"alex@example.com"}`,
		rec.Body.String(),
	)
	assert.Equal(
		t,
		&domain.Customer{FirstName: "Alex", LastName: "Example", EmailAddress: "alex@example.com"},
		customers.created,
	)
}

// Verifies: DH-0722
func TestAnAccountIsCreatedForTheCustomerOfThePathAndReturned(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, createAccount(accountBody))

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.JSONEq(t, accountJSON, rec.Body.String())
	assert.Equal(t, customerNumber, customers.accountOwner)
	assert.Equal(
		t,
		&domain.Account{
			DisplayName:     "Everyday",
			ProductCategory: domain.CategoryTransAndSavingsAccounts,
			ProductName:     "Mock Everyday Account",
			BSB:             "123456",
			AccountNumber:   "12345678",
		},
		customers.account,
	)
}

func TestAnAccountIsCreatedWithItsCreditLimit(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, createAccount(strings.Replace(accountBody, "}", `,"creditLimit":"1500.00"}`, 1)))

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"creditLimit":"1500.00"`)
	assert.Equal(t, domain.Amount(150_000), customers.account.CreditLimit)
}

// Verifies: DH-0731, DH-0741
func TestACreditLimitThatIsNotAnAmountWithTwoDecimalDigitsIsRefusedWithoutItsValue(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, createAccount(strings.Replace(accountBody, "}", `,"creditLimit":"1500.5"}`, 1)))

	detail := requireProblem(t, rec, http.StatusBadRequest)
	assert.Contains(t, detail, "creditLimit")
	assert.Contains(t, detail, "a string with 1 to 16 digits, a full stop and 2 digits")
	assert.NotContains(t, detail, "1500.5")
	assert.Nil(t, customers.account)
}

// Verifies: DH-0723
func TestAnAccountIsClosedAndReturned(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, closeAccount())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, strings.Replace(accountJSON, `"OPEN"`, `"CLOSED"`, 1), rec.Body.String())
	assert.Equal(t, [2]string{customerNumber, accountID}, customers.closed)
}

// Verifies: DH-0721
func TestACustomerIsReadWithItsAccounts(t *testing.T) {
	customers := &fakeCustomers{record: app.CustomerRecord{
		Customer: domain.Customer{
			CustomerNumber: customerNumber,
			LastName:       "Example",
			EmailAddress:   "alex@example.com",
		},
		Accounts: []domain.Account{{
			AccountID:       accountID,
			DisplayName:     "Everyday",
			ProductCategory: domain.CategoryTransAndSavingsAccounts,
			ProductName:     "Mock Everyday Account",
			BSB:             "123456",
			AccountNumber:   "12345678",
			OpenStatus:      domain.OpenStatusOpen,
		}},
	}}

	rec := serve(t, customers, readCustomer())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, customerNumber, customers.readNumber)
	assert.JSONEq(
		t,
		`{"customerNumber":"00000001","lastName":"Example","emailAddress":"alex@example.com","accounts":[`+
			accountJSON+`]}`,
		rec.Body.String(),
	)
}

func TestACustomerWithNoAccountsHasAnEmptyList(t *testing.T) {
	customers := &fakeCustomers{record: app.CustomerRecord{Customer: domain.Customer{CustomerNumber: customerNumber}}}

	rec := serve(t, customers, readCustomer())

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"accounts":[]`)
}

func TestABodyThatIsNotTheExpectedJSONObjectIsRefused(t *testing.T) {
	bodies := map[string]string{
		"empty":             "",
		"not JSON":          "lastName=Example",
		"a list":            "[" + customerBody + "]",
		"an unknown field":  `{"lastName":"Example","emailAddress":"alex@example.com","dateOfBirth":"2000-01-01"}`,
		"a customer number": `{"lastName":"Example","emailAddress":"alex@example.com","customerNumber":"1"}`,
		"a wrong type":      `{"lastName":1,"emailAddress":"alex@example.com"}`,
		"two objects":       customerBody + customerBody,
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			customers := &fakeCustomers{}

			rec := serve(t, customers, createCustomer(body))

			requireProblem(t, rec, http.StatusBadRequest)
			assert.Nil(t, customers.created)
		})
	}
}

// Verifies: DH-2615
func TestABodyLargerThan64KiBGives413(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, createCustomer(`{"lastName":"`+strings.Repeat("a", 64<<10)+`"}`))

	assert.Equal(t, "The request body is larger than 64 KiB.", requireProblem(t, rec, http.StatusRequestEntityTooLarge))
	assert.Nil(t, customers.created)
}

// Verifies: DH-0730
func TestAnAccountBodyWithAnUnknownFieldIsRefused(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, createAccount(`{"displayName":"Everyday","balance":"1.00"}`))

	requireProblem(t, rec, http.StatusBadRequest)
	assert.Nil(t, customers.account)
}

func TestAnAccountBodyWithAnAccountIdentifierIsRefused(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, createAccount(accountJSON))

	requireProblem(t, rec, http.StatusBadRequest)
	assert.Nil(t, customers.account)
}

func TestErrors(t *testing.T) {
	invalid := domain.ValidationError{Errors: []domain.FieldError{
		{Field: "bsb", Detail: "bsb must have 6 digits"},
		{Field: "accountNumber", Detail: "accountNumber must have 6 to 10 digits"},
	}}

	tests := map[string]struct {
		request request
		err     error
		status  int
		detail  string
	}{
		"customer that is not valid": {
			request: createCustomer(customerBody),
			err:     invalid,
			status:  http.StatusBadRequest,
			detail:  "bsb must have 6 digits; accountNumber must have 6 to 10 digits",
		},
		"customer with an email address in use": {
			request: createCustomer(customerBody),
			err:     errors.Join(errors.New("add the customer"), app.ErrEmailAddressUsed),
			status:  http.StatusConflict,
			detail:  "A customer has this emailAddress.",
		},
		"account that is not valid": {
			request: createAccount(accountBody),
			err:     invalid,
			status:  http.StatusBadRequest,
			detail:  "bsb must have 6 digits; accountNumber must have 6 to 10 digits",
		},
		"account for no customer": {
			request: createAccount(accountBody),
			err:     errors.Join(errors.New("add the account"), app.ErrCustomerNotFound),
			status:  http.StatusNotFound,
			detail:  "The data holder has no customer with this customer number.",
		},
		"account that exists": {
			request: createAccount(accountBody),
			err:     errors.Join(errors.New("add the account"), app.ErrAccountExists),
			status:  http.StatusConflict,
			detail:  "An account has this bsb and accountNumber.",
		},
		"close of an account that the customer does not have": {
			request: closeAccount(),
			err:     app.ErrAccountNotFound,
			status:  http.StatusNotFound,
			detail:  "The customer has no account with this accountId.",
		},
		"close of a closed account": {
			request: closeAccount(),
			err:     domain.ErrAccountClosed,
			status:  http.StatusConflict,
			detail:  "The account is closed.",
		},
		"close for no customer": {
			request: closeAccount(),
			err:     errors.Join(errors.New("read the customer"), app.ErrCustomerNotFound),
			status:  http.StatusNotFound,
			detail:  "The data holder has no customer with this customer number.",
		},
		"read of no customer": {
			request: readCustomer(),
			err:     app.ErrCustomerNotFound,
			status:  http.StatusNotFound,
			detail:  "The data holder has no customer with this customer number.",
		},
		"unexpected create error": {
			request: createCustomer(customerBody),
			err:     errors.New("connection lost: alex@example.com"),
			status:  http.StatusInternalServerError,
			detail:  "The data holder could not complete the request.",
		},
		"unexpected read error": {
			request: readCustomer(),
			err:     errors.New("connection lost"),
			status:  http.StatusInternalServerError,
			detail:  "The data holder could not complete the request.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, &fakeCustomers{err: tt.err}, tt.request)

			assert.Equal(t, tt.detail, requireProblem(t, rec, tt.status))
		})
	}
}

func TestEachRouteAndAnUnknownPathNeedTheOperatorToken(t *testing.T) {
	requests := map[string]request{
		"create customer": createCustomer(customerBody),
		"create account":  createAccount(accountBody),
		"close account":   closeAccount(),
		"add transaction": createTransaction(transactionBody),
		"read transactions": {
			method: http.MethodGet, path: transactionsPath, header: authorised(),
		},
		"read customer": readCustomer(),
		"unknown path":  {method: http.MethodGet, path: "/v1/accounts", header: authorised()},
	}
	for name, req := range requests {
		t.Run(name+" no token", func(t *testing.T) {
			customers := &fakeCustomers{}
			req.header = http.Header{}

			rec := serve(t, customers, req)

			requireProblem(t, rec, http.StatusUnauthorized)
			assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
			assert.Nil(t, customers.created)
			assert.Nil(t, customers.account)
			assert.Empty(t, customers.closed)
			assert.Empty(t, customers.transactionsOf)
			assert.Empty(t, customers.readNumber)
		})
		t.Run(name+" wrong token", func(t *testing.T) {
			req.header = http.Header{"Authorization": {"Bearer not-the-token"}}

			rec := serve(t, &fakeCustomers{}, req)

			requireProblem(t, rec, http.StatusUnauthorized)
			assert.Equal(t, `Bearer error="invalid_token"`, rec.Header().Get("WWW-Authenticate"))
		})
	}
}

func TestAnUnknownPathOrMethodIsNotFound(t *testing.T) {
	requests := map[string]request{
		"unknown path":      {method: http.MethodGet, path: "/v1/accounts", header: authorised()},
		"list of customers": {method: http.MethodGet, path: "/v1/customers", header: authorised()},
		"change a customer": {method: http.MethodPut, path: "/v1/customers/" + customerNumber, header: authorised()},
		"remove a customer": {
			method: http.MethodDelete,
			path:   "/v1/customers/" + customerNumber,
			header: authorised(),
		},
		"list of accounts": {
			method: http.MethodGet,
			path:   "/v1/customers/" + customerNumber + "/accounts",
			header: authorised(),
		},
	}
	for name, req := range requests {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, &fakeCustomers{}, req)

			requireProblem(t, rec, http.StatusNotFound)
		})
	}
}

// Verifies: DH-2628
func TestTheOperatorInterfaceServesOnlyThePathsUnderV1(t *testing.T) {
	unknown := serve(t, &fakeCustomers{}, request{method: http.MethodGet, path: "/v1/unknown", header: http.Header{}})
	requireProblem(t, unknown, http.StatusUnauthorized)

	for _, req := range []request{
		{method: http.MethodGet, path: "/", header: authorised()},
		{method: http.MethodPost, path: "/customers", body: customerBody, header: authorised()},
		{method: http.MethodGet, path: "/v2/unknown", header: http.Header{}},
	} {
		rec := serve(t, &fakeCustomers{}, req)

		assert.Equal(t, http.StatusNotFound, rec.Code, req.path)
		assert.Empty(t, rec.Header().Get("WWW-Authenticate"), req.path)
	}
}
