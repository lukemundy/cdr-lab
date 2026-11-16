package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const transactionID = "0d9a2b4e-1f3c-8a5b-9c6d-7e8f9a0b1c2d"

var (
	transactionsPath = accountsPath + "/" + accountID + "/transactions"
	postedAt         = time.Date(2026, 10, 15, 1, 30, 0, 0, time.UTC)
)

// accountTransactions keeps the last filter.
type accountTransactions struct {
	filter app.TransactionFilter
}

func consentTransaction() app.ConsentTransaction {
	executed := postedAt.Add(-time.Hour)

	return app.ConsentTransaction{ID: transactionID, Transaction: domain.Transaction{
		TransactionID:     "internal-1",
		Type:              domain.TransactionTransferOutgoing,
		Status:            domain.TransactionPosted,
		Description:       "Rent",
		Amount:            -120_000,
		Reference:         "October",
		PostingDateTime:   &postedAt,
		ExecutionDateTime: &executed,
		ExtendedData: &domain.ExtendedData{
			Payee: "Landlord", ExtensionUType: domain.ExtensionNPPPayload,
			NPPPayload: domain.NPPPayload{
				Service: domain.NPPServiceX2P1, ServiceVersion: "03", ExtendedDescription: "Rent for October",
			},
		},
	}}
}

func (a *accountTransactions) Transactions(
	_ context.Context,
	arrangementID, id string,
	filter app.TransactionFilter,
) ([]app.ConsentTransaction, int, error) {
	if arrangementID != "arrangement-1" {
		return nil, 0, errors.New("database failure")
	}

	if id != accountID {
		return nil, 0, app.ErrAccountNotFound
	}

	a.filter = filter

	return []app.ConsentTransaction{consentTransaction()}, 1, nil
}

func (a *accountTransactions) Transaction(
	_ context.Context,
	arrangementID, account, id string,
) (app.ConsentTransaction, error) {
	switch {
	case arrangementID != "arrangement-1":
		return app.ConsentTransaction{}, errors.New("database failure")
	case account != accountID:
		return app.ConsentTransaction{}, app.ErrAccountNotFound
	case id != transactionID:
		return app.ConsentTransaction{}, app.ErrTransactionNotFound
	}

	return consentTransaction(), nil
}

func transactionsRequest(t *testing.T, target, token string) *http.Request {
	t.Helper()

	req := resourceRequest(t, http.MethodGet, target, token)
	req.Header.Set("x-v", "2")

	return req
}

func transactionDetailRequest(t *testing.T, target string) *http.Request {
	t.Helper()

	req := resourceRequest(t, http.MethodGet, target, "transactions")
	req.Header.Set("x-v", "3")

	return req
}

const transactionFields = `
	"accountId": "` + accountID + `",
	"transactionId": "` + transactionID + `",
	"isDetailAvailable": true,
	"type": "TRANSFER_OUTGOING",
	"status": "POSTED",
	"description": "Rent",
	"postingDateTime": "2026-10-15T01:30:00Z",
	"executionDateTime": "2026-10-15T00:30:00Z",
	"amount": "-1200.00",
	"reference": "October"`

// Verifies: DH-1080, DH-1081
func TestGetTransactionsGivesTheTransactionsWithTheIDsOfTheSoftwareProduct(t *testing.T) {
	rec := serve(transactionsRequest(t, transactionsPath, "transactions"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "2", rec.Header().Get("x-v"))

	self := resourceBaseURL + transactionsPath + "?page=1&page-size=25"
	assert.JSONEq(t, `{
		"data": {"transactions": [{`+transactionFields+`}]},
		"links": {"self": "`+self+`", "first": "`+self+`", "last": "`+self+`"},
		"meta": {"totalRecords": 1, "totalPages": 1, "isQueryParamUnsupported": false}
	}`, rec.Body.String())
}

func TestGetTransactionsGivesEachFilterToTheApp(t *testing.T) {
	transactions := &accountTransactions{}
	s := consentSecure()
	s.Transactions = transactions
	query := "?oldest-time=2026-07-01T00:00:00%2B08:00&newest-time=2026-10-15T00:00:00Z" +
		"&min-amount=-10.005&max-amount=10.005&text=Rent"

	rec := httptest.NewRecorder()
	secureMux(s).ServeHTTP(rec, transactionsRequest(t, transactionsPath+query, "transactions"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	oldest := time.Date(2026, 6, 30, 16, 0, 0, 0, time.UTC)
	newest := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	assert.True(t, oldest.Equal(*transactions.filter.Oldest))
	assert.Equal(t, newest, *transactions.filter.Newest)
	assert.Equal(t, domain.Amount(-1000), *transactions.filter.MinAmount, "the smallest amount that is not less")
	assert.Equal(t, domain.Amount(1000), *transactions.filter.MaxAmount, "the largest amount that is not more")
	assert.Equal(t, "Rent", transactions.filter.Text)

	var body struct {
		Links map[string]string `json:"links"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Contains(t, body.Links["self"], "min-amount=-10.005")
}

func TestGetTransactionsRefusesAFilterThatIsNotValid(t *testing.T) {
	tests := map[string]struct {
		query string
		code  cds.Error
		field string
	}{
		"oldest-time with no offset": {"oldest-time=2026-10-15T00:00:00", cds.InvalidDateTime, "oldest-time"},
		"newest-time as a date":      {"newest-time=2026-10-15", cds.InvalidDateTime, "newest-time"},
		"min-amount with 1 decimal":  {"min-amount=10.5", cds.FieldInvalid, "min-amount"},
		"max-amount not a number":    {"max-amount=ten", cds.FieldInvalid, "max-amount"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(transactionsRequest(t, transactionsPath+"?"+tt.query, "transactions"))

			requireCDSError(t, rec, http.StatusBadRequest, tt.code, tt.field)
		})
	}
}

// Verifies: DH-1091
func TestGetTransactionDetailGivesTheTransactionWithItsExtendedData(t *testing.T) {
	rec := serve(transactionDetailRequest(t, transactionsPath+"/"+transactionID))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "3", rec.Header().Get("x-v"))
	assert.JSONEq(t, `{
		"data": {`+transactionFields+`,
			"extendedData": {
				"payee": "Landlord",
				"extensionUType": "nppPayload",
				"nppPayload": {"service": "X2P1", "serviceVersion": "03", "extendedDescription": "Rent for October"}
			}
		},
		"links": {"self": "`+resourceBaseURL+transactionsPath+"/"+transactionID+`"},
		"meta": {}
	}`, rec.Body.String())
}

func TestTheTransactionAPIsGiveTheNotFoundErrorsOfTheCDS(t *testing.T) {
	otherAccount := accountsPath + "/other-account/transactions"

	requireCDSError(t, serve(transactionsRequest(t, otherAccount, "transactions")),
		http.StatusNotFound, cds.InvalidBankingAccount, "other-account")
	requireCDSError(t, serve(transactionDetailRequest(t, otherAccount+"/"+transactionID)),
		http.StatusNotFound, cds.InvalidBankingAccount, "other-account")
	requireCDSError(t, serve(transactionDetailRequest(t, transactionsPath+"/internal-1")),
		http.StatusNotFound, cds.ResourceInvalid, "internal-1")
}

func TestTheTransactionAPIsNeedTheirScopeAndVersion(t *testing.T) {
	requireCDSError(t, serve(transactionsRequest(t, transactionsPath, "bank-only")), http.StatusForbidden,
		cds.InvalidConsent, "The consent does not have the scope bank:transactions:read.")

	detail := transactionDetailRequest(t, transactionsPath+"/"+transactionID)
	detail.Header.Set("x-v", "2")
	requireCDSError(
		t, serve(detail), http.StatusNotAcceptable, cds.UnsupportedVersion, "The supported versions are 3 to 3.",
	)
}

func TestAFailureToReadTheTransactionsGivesUnexpectedError(t *testing.T) {
	s := consentSecure()
	s.ConsentTokens = consentTokens{
		"other": {Scopes: []string{"bank:transactions:read"}, ArrangementID: "arrangement-9"},
	}

	for _, req := range []*http.Request{
		transactionsRequest(t, transactionsPath, "other"),
		func() *http.Request {
			r := transactionsRequest(t, transactionsPath+"/"+transactionID, "other")
			r.Header.Set("x-v", "3")

			return r
		}(),
	} {
		rec := httptest.NewRecorder()
		secureMux(s).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestTheLogOfGetTransactionDetailHasTheRouteAndNoTransactionData(t *testing.T) {
	logs := captureLogs(t)

	serve(transactionDetailRequest(t, transactionsPath+"/"+transactionID))

	assert.Contains(t, logs.String(), `"http.route":"GET `+accountsPath+`/{accountId}/transactions/{transactionId}"`)

	for _, value := range []string{transactionID, accountID, "Rent", "1200", "Landlord"} {
		assert.NotContains(t, logs.String(), value)
	}
}
