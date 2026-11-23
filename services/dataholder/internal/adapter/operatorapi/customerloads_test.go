package operatorapi_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func loadCustomers(body string) request {
	return request{method: http.MethodPost, path: "/v1/customer-loads", body: body, header: authorised()}
}

// Verifies: DH-2649, DH-2662
func TestALoadGivesTheResultOfEachCustomerInTheSequenceOfTheFile(t *testing.T) {
	customers := &fakeCustomers{loadResults: []app.LoadResult{
		{CustomerNumber: "00000001", Created: true}, {CustomerNumber: "00000002", Created: false},
	}}

	rec := serve(t, customers, loadCustomers(`{"customers": [
		{"lastName": "Lovelace", "emailAddress": "ada@mail.test", "accounts": [
			{"productId": "saver", "accountNumber": "11112222", "transactions": [
				{"type": "FEE", "status": "POSTED", "description": "Fee", "amount": "-1.00", "age": "P1D"}]}]},
		{"lastName": "Hopper", "emailAddress": "grace@mail.test"}]}`))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, `{"customers": [{"customerNumber": "00000001", "created": true},
		{"customerNumber": "00000002", "created": false}]}`, rec.Body.String())
	require.NotNil(t, customers.load)
	require.Len(t, *customers.load.Customers, 2)
	assert.Equal(t, "P1D", (*customers.load.Customers)[0].Accounts[0].Transactions[0].Age)
}

// Verifies: DH-0730, DH-2657
func TestALoadTransactionWithATimeOfACreateIsRefused(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, loadCustomers(`{"customers": [{"lastName": "Lovelace", "emailAddress": "ada@mail.test",
		"accounts": [{"productId": "saver", "accountNumber": "11112222", "transactions": [
			{"type": "FEE", "status": "POSTED", "description": "Fee", "amount": "-1.00",
				"postingDateTime": "2026-10-15T09:30:00+08:00"}]}]}]}`))

	assert.Contains(t, requireProblem(t, rec, http.StatusBadRequest), "postingDateTime")
	assert.Nil(t, customers.load)
}

// Verifies: DH-2654, DH-2659
func TestAnInvalidLoadGives400WithTheErrorsAtTheirPositions(t *testing.T) {
	customers := &fakeCustomers{err: domain.ValidationError{Errors: domain.AtPosition("customers[1].accounts[0]",
		domain.ValidationError{Errors: []domain.FieldError{
			{Field: "productId", Detail: "the data holder has no product with this productId"},
		}})}}

	rec := serve(t, customers, loadCustomers(`{"customers": []}`))

	assert.Equal(t, "customers[1].accounts[0]: the data holder has no product with this productId",
		requireProblem(t, rec, http.StatusBadRequest))
}

// Verifies: DH-2664
func TestALoadAccountWithABSBAndAccountNumberInUseGives409AtItsPosition(t *testing.T) {
	customers := &fakeCustomers{
		err: app.LoadRecordError{Position: "customers[2].accounts[1]", Err: app.ErrAccountExists},
	}

	rec := serve(t, customers, loadCustomers(`{"customers": []}`))

	assert.Equal(t, "customers[2].accounts[1]: An account has this bsb and accountNumber.",
		requireProblem(t, rec, http.StatusConflict))
}

// Verifies: DH-2615
func TestALoadHasABodyLimitOf1MiB(t *testing.T) {
	customers := &fakeCustomers{loadResults: []app.LoadResult{}}
	padding := `{"customers": [{"lastName": "%s", "emailAddress": "ada@mail.test"}]}`

	rec := serve(t, customers, loadCustomers(strings.Replace(padding, "%s", strings.Repeat("a", 100<<10), 1)))

	require.Equal(t, http.StatusOK, rec.Code, "a load larger than 64 KiB is accepted")

	customers.load = nil
	rec = serve(t, customers, loadCustomers(strings.Replace(padding, "%s", strings.Repeat("a", 1<<20), 1)))

	assert.Equal(t, "The request body is larger than 1 MiB.", requireProblem(t, rec, http.StatusRequestEntityTooLarge))
	assert.Nil(t, customers.load)
}
