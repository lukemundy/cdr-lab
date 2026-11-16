package console_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	everydayTransactions = "/api/customers/00000001/accounts/account-1/transactions"
	noTransactions       = `{"data": [], "totalRecords": 0}`
)

// withTransactionCreates gives the ports the use case of the operator interface, so that a create has the same rules.
func (r records) withTransactionCreates() records {
	r.ports.Creates = app.Customers{
		Store: r.customers, Transactions: r.transactions, NewID: func() string { return "transaction/new" },
	}

	return r
}

func anOpenAccount(t *testing.T) records {
	t.Helper()

	r := newRecords(t).withTransactionCreates()
	addCustomer(t, r.customers, "00000001", "Alex", "Example")
	r.addAccount(t, "00000001", domain.Account{AccountID: "account-1", DisplayName: "Everyday", AccountNumber: "1"})

	return r
}

// Verifies: DH-1654
func TestACreateOfATransactionGivesTheLocationOfTheNewTransactionAndNoBody(t *testing.T) {
	r := anOpenAccount(t)

	created := post(t, r.ports, everydayTransactions, `{
		"type": "TRANSFER_INCOMING", "status": "POSTED", "description": "Salary", "amount": "2500.00",
		"reference": "Pay 42", "postingDateTime": "2026-10-20T09:30:00+08:00",
		"executionDateTime": "2026-10-20T09:00:00+08:00", "valueDateTime": "2026-10-21T00:00:00Z",
		"merchantName": "Example Employer", "merchantCategoryCode": "5411", "billerCode": "12345",
		"billerName": "Example Biller", "crn": "987654", "apcaNumber": "654321",
		"extendedData": {"payer": "Example Employer", "extensionUType": "nppPayload",
			"nppPayload": {"extendedDescription": "October pay", "service": "X2P1", "serviceVersion": "03"}}
	}`)

	require.Equal(t, http.StatusCreated, created.status, created.body)
	assert.Empty(t, created.body)
	assert.Equal(t, everydayTransactions+"/transaction%2Fnew", created.location)

	resp := get(t, r.ports, everydayTransactions)
	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{"data": [{
		"transactionId": "transaction/new", "type": "TRANSFER_INCOMING", "status": "POSTED",
		"description": "Salary", "amount": "2500.00", "reference": "Pay 42",
		"postingDateTime": "2026-10-20T01:30:00Z", "executionDateTime": "2026-10-20T01:00:00Z",
		"valueDateTime": "2026-10-21T00:00:00Z", "merchantName": "Example Employer",
		"merchantCategoryCode": "5411", "billerCode": "12345", "billerName": "Example Biller", "crn": "987654",
		"apcaNumber": "654321",
		"extendedData": {"payer": "Example Employer", "extensionUType": "nppPayload",
			"nppPayload": {"extendedDescription": "October pay", "service": "X2P1", "serviceVersion": "03"}}
	}], "totalRecords": 1}`, resp.body)
}

func TestATransactionWithOnlyTheRequiredFieldsIsCreated(t *testing.T) {
	r := anOpenAccount(t)

	resp := post(t, r.ports, everydayTransactions, `{
		"type": "FEE", "status": "PENDING", "description": "Fee", "amount": "-2.50",
		"executionDateTime": "2026-10-20T00:00:00Z"
	}`)

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.JSONEq(t, `{"data": [{
		"transactionId": "transaction/new", "type": "FEE", "status": "PENDING", "description": "Fee",
		"amount": "-2.50", "reference": "", "executionDateTime": "2026-10-20T00:00:00Z"
	}], "totalRecords": 1}`, get(t, r.ports, everydayTransactions).body)
}

func TestARefusedValueInTheExtendedDataGivesAnErrorAtItsFieldAndCreatesNothing(t *testing.T) {
	r := anOpenAccount(t)

	resp := post(t, r.ports, everydayTransactions, `{
		"type": "PAYMENT", "status": "POSTED", "description": "Refund", "amount": "12.50",
		"postingDateTime": "2026-10-20T00:00:00Z",
		"extendedData": {"payee": "Example Shop", "extensionUType": "nppPayload",
			"nppPayload": {"extendedDescription": "Refund", "service": "zebracorn", "serviceVersion": "03"}}
	}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/extendedData/nppPayload/service",
				"detail": "extendedData.nppPayload.service must be a value of the CDS enumeration"},
			{"pointer": "#/extendedData/payer",
				"detail": "extendedData.payer must be present for a positive amount"}
		]
	}`, resp.body)
	assert.JSONEq(t, noTransactions, get(t, r.ports, everydayTransactions).body)
}

// Verifies: DH-1657
func TestAValueWithNoValidFormatGivesAnErrorAtItsFieldAndCreatesNothing(t *testing.T) {
	r := anOpenAccount(t)

	for field, body := range map[string]string{
		"amount": `{"type": "FEE", "status": "PENDING", "description": "Fee", "amount": "2.5",
			"executionDateTime": "2026-10-20T00:00:00Z"}`,
		"executionDateTime": `{"type": "FEE", "status": "PENDING", "description": "Fee", "amount": "2.50",
			"executionDateTime": "2026-10-20 08:00"}`,
	} {
		resp := post(t, r.ports, everydayTransactions, body)

		require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
		assert.Contains(t, resp.body, `"pointer":"#/`+field+`"`, field)
	}

	assert.JSONEq(t, noTransactions, get(t, r.ports, everydayTransactions).body)
}

func TestATransactionForAClosedAccountIsAConflictAndCreatesNothing(t *testing.T) {
	r := anOpenAccount(t).withCloses()
	require.Equal(t, http.StatusNoContent, post(t, r.ports, closeEveryday, endsNothing).status)

	resp := post(t, r.ports, everydayTransactions, `{
		"type": "FEE", "status": "PENDING", "description": "Fee", "amount": "-2.50",
		"executionDateTime": "2026-10-20T00:00:00Z"
	}`)

	require.Equal(t, http.StatusConflict, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "about:blank", "title": "Conflict", "status": 409,
		"detail": "The account is closed. The data holder changed nothing."
	}`, resp.body)
	assert.JSONEq(t, noTransactions, get(t, r.ports, everydayTransactions).body)
}

// Verifies: DH-1622
func TestATransactionForAnUnknownCustomerOrAccountIsNotFoundWithNoIdentifierInTheDetail(t *testing.T) {
	r := anOpenAccount(t)

	for _, target := range []string{
		"/api/customers/00047291/accounts/account-1/transactions",
		"/api/customers/00000001/accounts/account-zebracorn/transactions",
	} {
		resp := post(t, r.ports, target, `{
			"type": "FEE", "status": "PENDING", "description": "Fee", "amount": "-2.50",
			"executionDateTime": "2026-10-20T00:00:00Z"
		}`)

		require.Equal(t, http.StatusNotFound, resp.status, target)
		assert.NotContains(t, resp.body, "00047291", target)
		assert.NotContains(t, resp.body, "zebracorn", target)
	}
}

func TestACreateOfATransactionWithABodyThatIsNotValidIsABadRequestWithNoValue(t *testing.T) {
	r := anOpenAccount(t)

	for _, body := range []string{
		"", "zebracorn", "null", "[]",
		`{"type": "FEE", "status": "PENDING", "description": "Fee", "amount": -2.5}`,
		`{"transactionId": "zebracorn", "type": "FEE", "status": "PENDING", "description": "Fee", "amount": "-2.50"}`,
		`{"type": "FEE", "status": "PENDING", "description": "Fee", "amount": "-2.50",
			"extendedData": {"extensionUType": "nppPayload", "zebracorn": true}}`,
	} {
		resp := post(t, r.ports, everydayTransactions, body)

		require.Equal(t, http.StatusBadRequest, resp.status, body)
		assert.JSONEq(t, `{
			"type": "about:blank", "title": "Bad Request", "status": 400,
			"detail": "The request body must be a JSON object with only the fields of the transaction."
		}`, resp.body, body)
	}

	assert.JSONEq(t, noTransactions, get(t, r.ports, everydayTransactions).body)
}
