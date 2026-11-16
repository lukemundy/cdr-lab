package operatorapi_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	transactionsPath = "/v1/customers/" + customerNumber + "/accounts/" + accountID + "/transactions"
	transactionBody  = `{"type":"PAYMENT","status":"POSTED","description":"Rent","amount":"-1200.00",` +
		`"reference":"October","postingDateTime":"2026-10-15T09:30:00+08:00",` +
		`"extendedData":{"payee":"Landlord","extensionUType":"nppPayload",` +
		`"nppPayload":{"service":"X2P1","serviceVersion":"03","extendedDescription":"Rent for October"}}}`
)

func createTransaction(body string) request {
	return request{method: http.MethodPost, path: transactionsPath, body: body, header: authorised()}
}

// Verifies: DH-0724
func TestATransactionIsCreatedForTheAccountOfThePathAndReturned(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, createTransaction(transactionBody))

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, [2]string{customerNumber, accountID}, customers.transactionsOf)

	postedAt := time.Date(2026, 10, 15, 1, 30, 0, 0, time.UTC)
	assert.True(t, postedAt.Equal(*customers.transaction.PostingDateTime))

	customers.transaction.PostingDateTime = nil
	assert.Equal(t, &domain.Transaction{
		Type:        domain.TransactionPayment,
		Status:      domain.TransactionPosted,
		Description: "Rent",
		Amount:      -120_000,
		Reference:   "October",
		ExtendedData: &domain.ExtendedData{
			Payee: "Landlord", ExtensionUType: domain.ExtensionNPPPayload,
			NPPPayload: domain.NPPPayload{
				Service: domain.NPPServiceX2P1, ServiceVersion: "03", ExtendedDescription: "Rent for October",
			},
		},
	}, customers.transaction)
	assert.Contains(t, rec.Body.String(), `"transactionId":"transaction-1"`)
	assert.Contains(t, rec.Body.String(), `"amount":"-1200.00"`)
}

// Verifies: DH-0731, DH-0749
func TestATransactionWithAnAmountOrATimeThatHasNoValidFormatIsRefusedWithoutItsValue(t *testing.T) {
	tests := map[string]struct {
		body string
		want []string
	}{
		"no amount": {
			body: strings.Replace(transactionBody, `"amount":"-1200.00",`, "", 1),
			want: []string{"amount must be present"},
		},
		"amount and time": {
			body: strings.NewReplacer(`"-1200.00"`, `"-1200"`, `"2026-10-15T09:30:00+08:00"`, `"2026-10-15T09:30:00"`).
				Replace(transactionBody),
			want: []string{
				"amount must be a string with 1 to 16 digits",
				"postingDateTime must be an RFC 3339 date-time with an offset",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			customers := &fakeCustomers{}

			detail := requireProblem(t, serve(t, customers, createTransaction(tt.body)), http.StatusBadRequest)

			for _, want := range tt.want {
				assert.Contains(t, detail, want)
			}

			assert.NotContains(t, detail, "-1200")
			assert.NotContains(t, detail, "2026-10-15")
			assert.Nil(t, customers.transaction)
		})
	}
}

func TestATransactionWithAFieldThatTheDataHolderDoesNotHoldIsRefused(t *testing.T) {
	for _, body := range []string{
		strings.Replace(transactionBody, `"reference"`, `"transactionId":"t1","reference"`, 1),
		strings.Replace(transactionBody, `"serviceVersion"`, `"endToEndId":"e1","serviceVersion"`, 1),
	} {
		customers := &fakeCustomers{}

		requireProblem(t, serve(t, customers, createTransaction(body)), http.StatusBadRequest)
		assert.Nil(t, customers.transaction)
	}
}

func TestTheErrorsOfATransaction(t *testing.T) {
	tests := map[string]struct {
		err    error
		status int
	}{
		"rule": {
			err: domain.ValidationError{
				Errors: []domain.FieldError{{Field: "description", Detail: "description must have"}},
			},
			status: 400,
		},
		"closed account":              {err: domain.ErrAccountClosed, status: http.StatusConflict},
		"account of another customer": {err: app.ErrAccountNotFound, status: http.StatusNotFound},
		"no customer": {
			err: errors.Join(errors.New("read the customer"), app.ErrCustomerNotFound), status: http.StatusNotFound,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			requireProblem(t, serve(t, &fakeCustomers{err: tt.err}, createTransaction(transactionBody)), tt.status)
		})
	}
}

// Verifies: DH-0725
func TestTheTransactionsOfAnAccountAreRead(t *testing.T) {
	customers := &fakeCustomers{}

	rec := serve(t, customers, request{method: http.MethodGet, path: transactionsPath, header: authorised()})

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, `[]`, rec.Body.String())
	assert.Equal(t, [2]string{customerNumber, accountID}, customers.transactionsOf)

	missing := serve(t, &fakeCustomers{err: app.ErrAccountNotFound},
		request{method: http.MethodGet, path: transactionsPath, header: authorised()})
	requireProblem(t, missing, http.StatusNotFound)
}
