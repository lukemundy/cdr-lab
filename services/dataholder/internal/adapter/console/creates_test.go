package console_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	fieldErrorsType = "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002"
	newAccountID    = "5d1f7a52-6a3b-4c1e-9f0d-2b8e4c7a9e13"
	createAccounts  = "/api/customers/00000001/accounts"
	newAccountBody  = `{"displayName": "Everyday", "productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
		"productName": "Mock Everyday Account", "bsb": "654321", "accountNumber": "87654321", "creditLimit": "500.00"}`
)

// withCreates gives the ports the use case of the operator interface, so that a create has the same rules.
func (r records) withCreates() records {
	r.ports.Customers = r.customers
	r.ports.Creates = app.Customers{
		Store: r.customers, Now: func() time.Time { return now }, Random: func() uint64 { return 47_291 },
		NewID: func() string { return newAccountID },
	}

	return r
}

// Verifies: DH-1653
func TestACreateOfACustomerGivesTheLocationOfTheNewCustomerAndNoBody(t *testing.T) {
	r := newRecords(t).withCreates()

	created := post(t, r.ports, "/api/customers",
		`{"firstName": "Alex", "lastName": "Example", "emailAddress": "alex@mail.test"}`)

	require.Equal(t, http.StatusCreated, created.status, created.body)
	assert.Empty(t, created.body)
	assert.Equal(t, "/api/customers/00047291", created.location)

	resp := get(t, r.ports, "/api/customers/00047291")
	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var detail struct {
		FirstName    string    `json:"firstName"`
		LastName     string    `json:"lastName"`
		EmailAddress string    `json:"emailAddress"`
		LastUpdated  time.Time `json:"lastUpdated"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &detail))
	assert.Equal(t, "Alex", detail.FirstName)
	assert.Equal(t, "Example", detail.LastName)
	assert.Equal(t, "alex@mail.test", detail.EmailAddress)
	assert.True(t, now.Equal(detail.LastUpdated))
}

func TestACustomerWithNoFirstNameIsCreated(t *testing.T) {
	r := newRecords(t).withCreates()

	resp := post(t, r.ports, "/api/customers", `{"lastName": "Example", "emailAddress": "alex@mail.test"}`)

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.Equal(t, http.StatusOK, get(t, r.ports, "/api/customers/00047291").status)
}

func TestACustomerWithTheEmailAddressOfADifferentCustomerGivesAConflictAtTheEmailAddress(t *testing.T) {
	r := newRecords(t).withCreates()
	addCustomer(t, r.customers, "00000001", "Alex", "Example")

	resp := post(t, r.ports, "/api/customers", `{"lastName": "Sample", "emailAddress": "00000001@MAIL.test"}`)

	require.Equal(t, http.StatusConflict, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 409,
		"detail": "One or more values in the request are not valid.",
		"errors": [{"pointer": "#/emailAddress", "detail": "A customer has this emailAddress."}]
	}`, resp.body)
	assert.Contains(t, get(t, r.ports, "/api/customers").body, `"totalRecords":1`)
}

// Verifies: PROB-0002, PROB-0003, DH-1656
func TestACustomerThatTheRulesRefuseGivesAnErrorAtEachFieldAndIsNotCreated(t *testing.T) {
	r := newRecords(t).withCreates()

	resp := post(t, r.ports, "/api/customers",
		`{"firstName": "Alex", "lastName": "", "emailAddress": "zebracorn at mail.test"}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/lastName", "detail": "lastName must have 1 to 100 characters"},
			{"pointer": "#/emailAddress", "detail": "emailAddress must be one email address"}
		]
	}`, resp.body)
	assert.JSONEq(t, `{"data": [], "totalRecords": 0}`, get(t, r.ports, "/api/customers").body)
}

// Verifies: DH-1620
func TestACreateOfACustomerWithABodyThatIsNotValidIsABadRequestWithNoValue(t *testing.T) {
	r := newRecords(t).withCreates()

	for _, body := range []string{
		"", "zebracorn", "null", "[]", `{"lastName": 1, "emailAddress": "alex@mail.test"}`,
		`{"customerNumber": "00000001", "lastName": "zebracorn", "emailAddress": "alex@mail.test"}`,
		`{"lastName": "zebracorn", "emailAddress": "alex@mail.test", "accounts": []}`,
	} {
		resp := post(t, r.ports, "/api/customers", body)

		require.Equal(t, http.StatusBadRequest, resp.status, body)
		assert.JSONEq(t, `{
			"type": "about:blank", "title": "Bad Request", "status": 400,
			"detail": "The request body must be a JSON object with only the fields of the customer."
		}`, resp.body, body)
	}

	assert.JSONEq(t, `{"data": [], "totalRecords": 0}`, get(t, r.ports, "/api/customers").body)
}

type accountsOfDetail struct {
	Accounts []struct {
		AccountID       string `json:"accountId"`
		DisplayName     string `json:"displayName"`
		ProductCategory string `json:"productCategory"`
		ProductName     string `json:"productName"`
		BSB             string `json:"bsb"`
		AccountNumber   string `json:"accountNumber"`
		CreditLimit     string `json:"creditLimit"`
		OpenStatus      string `json:"openStatus"`
		CreationDate    string `json:"creationDate"`
	} `json:"accounts"`
}

func accountsOf(t *testing.T, ports console.Ports, customerNumber string) accountsOfDetail {
	t.Helper()

	resp := get(t, ports, "/api/customers/"+customerNumber)
	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var detail accountsOfDetail
	require.NoError(t, json.Unmarshal([]byte(resp.body), &detail))

	return detail
}

// Verifies: DH-1653
func TestACreateOfAnAccountGivesTheLocationOfTheNewAccountAndNoBody(t *testing.T) {
	r := newRecords(t).withCreates()
	addCustomer(t, r.customers, "00000001", "Alex", "Example")

	resp := post(t, r.ports, createAccounts, newAccountBody)

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.Empty(t, resp.body)
	assert.Equal(t, "/api/customers/00000001/accounts/"+newAccountID, resp.location)

	detail := accountsOf(t, r.ports, "00000001")
	require.Len(t, detail.Accounts, 1)
	account := detail.Accounts[0]
	assert.Equal(t, newAccountID, account.AccountID)
	assert.Equal(t, "Everyday", account.DisplayName)
	assert.Equal(t, "TRANS_AND_SAVINGS_ACCOUNTS", account.ProductCategory)
	assert.Equal(t, "Mock Everyday Account", account.ProductName)
	assert.Equal(t, "654321", account.BSB)
	assert.Equal(t, "87654321", account.AccountNumber)
	assert.Equal(t, "500.00", account.CreditLimit)
	assert.Equal(t, "OPEN", account.OpenStatus)
	assert.Equal(t, "2026-10-20", account.CreationDate)
}

func TestACustomerWithOnlyAClosedAccountGetsANewAccount(t *testing.T) {
	r := newRecords(t).withCreates()
	addCustomer(t, r.customers, "00000001", "Alex", "Example")

	old := domain.Account{AccountID: "account-1", DisplayName: "Old", AccountNumber: "11111111"}
	r.addAccount(t, "00000001", old)
	old.OpenStatus = domain.OpenStatusClosed
	_, err := r.customers.CloseAccount(t.Context(), old, now, nil)
	require.NoError(t, err)

	resp := post(t, r.ports, createAccounts, newAccountBody)

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.Len(t, accountsOf(t, r.ports, "00000001").Accounts, 2)
}

func TestAnAccountThatTheRulesRefuseGivesAnErrorAtEachFieldAndIsNotCreated(t *testing.T) {
	r := newRecords(t).withCreates()
	addCustomer(t, r.customers, "00000001", "Alex", "Example")

	resp := post(t, r.ports, createAccounts, `{"displayName": "", "productCategory": "SAVINGS",
		"productName": "Mock", "bsb": "123-456", "accountNumber": "12345"}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/displayName", "detail": "displayName must have 1 to 100 characters"},
			{"pointer": "#/productCategory", "detail": "productCategory must be a value of BankingProductCategoryV2"},
			{"pointer": "#/bsb", "detail": "bsb must have 6 digits"},
			{"pointer": "#/accountNumber", "detail": "accountNumber must have 6 to 10 digits"}
		]
	}`, resp.body)
	assert.Empty(t, accountsOf(t, r.ports, "00000001").Accounts)
}

// Verifies: DH-1657
func TestACreditLimitWithNoValidFormatGivesAnErrorAtTheCreditLimitAndAtEachOtherRefusedValue(t *testing.T) {
	r := newRecords(t).withCreates()
	addCustomer(t, r.customers, "00000001", "Alex", "Example")

	resp := post(t, r.ports, createAccounts, `{"displayName": "", "productCategory": "TRAVEL_CARDS",
		"productName": "Mock", "bsb": "654321", "accountNumber": "87654321", "creditLimit": "500"}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/displayName", "detail": "displayName must have 1 to 100 characters"},
			{
				"pointer": "#/creditLimit",
				"detail": "creditLimit must be a string with 1 to 16 digits, a full stop and 2 digits"
			}
		]
	}`, resp.body)
	assert.Empty(t, accountsOf(t, r.ports, "00000001").Accounts)
}

// Verifies: DH-1660
func TestAnAccountWithTheBSBAndAccountNumberOfADifferentAccountGivesAConflictAtTheAccountNumber(t *testing.T) {
	r := newRecords(t).withCreates()
	addCustomer(t, r.customers, "00000001", "Alex", "Example")
	addCustomer(t, r.customers, "00000002", "Sam", "Sample")
	r.addAccount(
		t,
		"00000002",
		domain.Account{AccountID: "account-2", DisplayName: "Theirs", AccountNumber: "87654321"},
	)

	resp := post(t, r.ports, createAccounts, `{"displayName": "Everyday", "productCategory": "TRAVEL_CARDS",
		"productName": "Mock", "bsb": "123456", "accountNumber": "87654321"}`)

	require.Equal(t, http.StatusConflict, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 409,
		"detail": "One or more values in the request are not valid.",
		"errors": [{"pointer": "#/accountNumber", "detail": "An account has this bsb and accountNumber."}]
	}`, resp.body)
	assert.Empty(t, accountsOf(t, r.ports, "00000001").Accounts)
}

// Verifies: DH-1622
func TestACreateOfAnAccountForAnUnknownCustomerIsNotFoundWithNoCustomerNumber(t *testing.T) {
	r := newRecords(t).withCreates()

	resp := post(t, r.ports, createAccounts, newAccountBody)

	require.Equal(t, http.StatusNotFound, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "about:blank", "title": "Not Found", "status": 404,
		"detail": "The data holder has no customer with this customer number."
	}`, resp.body)
}

func TestACreateOfAnAccountWithABodyThatIsNotValidIsABadRequestWithNoValue(t *testing.T) {
	r := newRecords(t).withCreates()
	addCustomer(t, r.customers, "00000001", "Alex", "Example")

	for _, body := range []string{
		"", "zebracorn", "null", "[]", `{"displayName": 1}`, `{"creditLimit": 500}`,
		`{"accountId": "zebracorn", "displayName": "Everyday"}`, `{"openStatus": "CLOSED"}`,
	} {
		resp := post(t, r.ports, createAccounts, body)

		require.Equal(t, http.StatusBadRequest, resp.status, body)
		assert.JSONEq(t, `{
			"type": "about:blank", "title": "Bad Request", "status": 400,
			"detail": "The request body must be a JSON object with only the fields of the account."
		}`, resp.body, body)
	}

	assert.Empty(t, accountsOf(t, r.ports, "00000001").Accounts)
}

// The UI gives the operator the values of the document, so they must be the values of the CDS.
func TestTheProductCategoriesOfTheDocumentAreTheValuesOfBankingProductCategoryV2(t *testing.T) {
	cds, err := os.ReadFile("../../../../../conformance/dataholder/spec/cds_banking.json")
	require.NoError(t, err)

	var banking struct {
		Components struct {
			Schemas struct {
				BankingProductCategoryV2 struct {
					Enum []string `json:"enum"`
				} `json:"BankingProductCategoryV2"`
			} `json:"schemas"`
		} `json:"components"`
	}
	require.NoError(t, json.Unmarshal(cds, &banking))

	document, err := openapi3.NewLoader().LoadFromFile("openapi.yaml")
	require.NoError(t, err)

	var categories []string
	for _, value := range document.Components.Schemas["ProductCategory"].Value.Enum {
		categories = append(categories, value.(string))
	}

	want := banking.Components.Schemas.BankingProductCategoryV2.Enum
	require.NotEmpty(t, want)
	assert.Equal(t, want, categories)

	for _, category := range want {
		_, ok := domain.ParseProductCategory(category)
		assert.True(t, ok, category)
	}
}
