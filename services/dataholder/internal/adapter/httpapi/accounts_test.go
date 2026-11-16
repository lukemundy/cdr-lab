package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
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

const (
	accountsPath = "/cds-au/v1/banking/accounts"
	accountID    = "8f2b6a39-4f0c-4f7e-9d5d-2b1c3e4f5a60"
)

// arrangementAccounts keeps the last query.
type arrangementAccounts struct {
	query app.AccountQuery
}

func consentAccounts() []app.ConsentAccount {
	return []app.ConsentAccount{
		{ID: accountID, Account: domain.Account{
			AccountID:       "internal-1",
			DisplayName:     "Everyday",
			ProductCategory: domain.CategoryTransAndSavingsAccounts,
			ProductName:     "Mock Everyday Account",
			BSB:             "123456",
			AccountNumber:   "12345678",
			CreditLimit:     150_000,
			OpenStatus:      domain.OpenStatusOpen,
			// After 10:00 in Perth on 16 October, and on 15 October in UTC.
			CreatedAt: time.Date(2026, 10, 16, 1, 0, 0, 0, time.FixedZone("AWST", 8*60*60)),
		}},
		{ID: "d0c1b2a3-0000-4000-8000-000000000002", Account: domain.Account{
			AccountID:       "internal-2",
			DisplayName:     "Card",
			ProductCategory: domain.CategoryCredAndChrgCards,
			ProductName:     "Mock Card",
			BSB:             "123456",
			AccountNumber:   "1234567890",
			OpenStatus:      domain.OpenStatusClosed,
			CreatedAt:       time.Date(2026, 10, 15, 13, 30, 0, 0, time.UTC),
		}},
	}
}

func (a *arrangementAccounts) Accounts(
	_ context.Context,
	arrangementID string,
	query app.AccountQuery,
) ([]app.ConsentAccount, int, error) {
	if arrangementID != "arrangement-1" {
		return nil, 0, errors.New("database failure")
	}

	a.query = query
	all := consentAccounts()
	start := min((query.Page-1)*query.PageSize, len(all))

	return all[start:min(start+query.PageSize, len(all))], len(all), nil
}

func (a *arrangementAccounts) Account(_ context.Context, arrangementID, id string) (app.ConsentAccount, error) {
	if arrangementID != "arrangement-1" {
		return app.ConsentAccount{}, errors.New("database failure")
	}

	for _, account := range consentAccounts() {
		if account.ID == id {
			return account, nil
		}
	}

	return app.ConsentAccount{}, app.ErrAccountNotFound
}

func accountsRequest(t *testing.T, target, token string) *http.Request {
	t.Helper()

	req := resourceRequest(t, http.MethodGet, target, token)
	req.Header.Set("x-v", "3")

	return req
}

func detailRequest(t *testing.T, id, token string) *http.Request {
	t.Helper()

	req := resourceRequest(t, http.MethodGet, accountsPath+"/"+id, token)
	req.Header.Set("x-v", "5")

	return req
}

func TestGetAccountsGivesTheAccountsOfTheArrangementWithTheIDsOfTheSoftwareProduct(t *testing.T) {
	accounts := &arrangementAccounts{}
	s := consentSecure()
	s.Accounts = accounts

	rec := httptest.NewRecorder()
	secureMux(s).ServeHTTP(rec, accountsRequest(t, accountsPath, "bank-only"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "3", rec.Header().Get("x-v"))
	assert.Equal(t, app.AccountQuery{Page: 1, PageSize: 25}, accounts.query)
	assert.JSONEq(t, `{
		"data": {"accounts": [
			{
				"accountId": "`+accountID+`",
				"creationDate": "2026-10-15",
				"displayName": "Everyday",
				"maskedNumber": "xxx-xxx xxxx5678",
				"openStatus": "OPEN",
				"isOwned": true,
				"accountOwnership": "ONE_PARTY",
				"productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
				"productName": "Mock Everyday Account"
			},
			{
				"accountId": "d0c1b2a3-0000-4000-8000-000000000002",
				"creationDate": "2026-10-15",
				"displayName": "Card",
				"maskedNumber": "xxx-xxx xxxxxx7890",
				"openStatus": "CLOSED",
				"isOwned": true,
				"accountOwnership": "ONE_PARTY",
				"productCategory": "CRED_AND_CHRG_CARDS",
				"productName": "Mock Card"
			}
		]},
		"links": {
			"self": "https://mtls.bank.example/cds-au/v1/banking/accounts?page=1&page-size=25",
			"first": "https://mtls.bank.example/cds-au/v1/banking/accounts?page=1&page-size=25",
			"last": "https://mtls.bank.example/cds-au/v1/banking/accounts?page=1&page-size=25"
		},
		"meta": {"totalRecords": 2, "totalPages": 1}
	}`, rec.Body.String())
}

func TestGetAccountsGivesEachFilterToTheQuery(t *testing.T) {
	owned, notOwned := true, false

	tests := map[string]struct {
		query string
		want  app.AccountQuery
	}{
		"product category": {
			query: "product-category=CRED_AND_CHRG_CARDS",
			want:  app.AccountQuery{Category: domain.CategoryCredAndChrgCards},
		},
		"open":      {query: "open-status=OPEN", want: app.AccountQuery{OpenStatus: domain.OpenStatusOpen}},
		"closed":    {query: "open-status=CLOSED", want: app.AccountQuery{OpenStatus: domain.OpenStatusClosed}},
		"all":       {query: "open-status=ALL", want: app.AccountQuery{}},
		"owned":     {query: "is-owned=true", want: app.AccountQuery{IsOwned: &owned}},
		"not owned": {query: "is-owned=false", want: app.AccountQuery{IsOwned: &notOwned}},
		"no value":  {query: "open-status=&is-owned=&product-category=", want: app.AccountQuery{}},
		"each value": {
			query: "product-category=CRED_AND_CHRG_CARDS&open-status=OPEN&is-owned=true",
			want: app.AccountQuery{
				Category: domain.CategoryCredAndChrgCards, OpenStatus: domain.OpenStatusOpen, IsOwned: &owned,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			accounts := &arrangementAccounts{}
			s := consentSecure()
			s.Accounts = accounts

			rec := httptest.NewRecorder()
			secureMux(s).ServeHTTP(rec, accountsRequest(t, accountsPath+"?"+tt.query+"&page-size=1", "bank-only"))

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			tt.want.Page, tt.want.PageSize = 1, 1
			assert.Equal(t, tt.want, accounts.query)
		})
	}
}

func TestGetAccountsRefusesAFilterThatIsNotValid(t *testing.T) {
	for query, field := range map[string]string{
		"product-category=SAVINGS": "product-category",
		"open-status=open":         "open-status",
		"open-status=DELETED":      "open-status",
		"is-owned=yes":             "is-owned",
		"is-owned=1":               "is-owned",
		"is-owned=TRUE":            "is-owned",
	} {
		rec := serve(accountsRequest(t, accountsPath+"?"+query, "bank-only"))

		requireCDSError(t, rec, http.StatusBadRequest, cds.FieldInvalid, field)
	}
}

func TestGetAccountsIsPaginated(t *testing.T) {
	rec := serve(accountsRequest(t, accountsPath+"?page=2&page-size=1&open-status=ALL", "bank-only"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct {
		Data struct {
			Accounts []map[string]any `json:"accounts"`
		} `json:"data"`
		Links map[string]string `json:"links"`
		Meta  map[string]int    `json:"meta"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Accounts, 1)
	assert.Equal(t, "d0c1b2a3-0000-4000-8000-000000000002", body.Data.Accounts[0]["accountId"])
	assert.Equal(t, map[string]int{"totalRecords": 2, "totalPages": 2}, body.Meta)
	assert.Equal(t, resourceBaseURL+accountsPath+"?open-status=ALL&page=1&page-size=1", body.Links["prev"])

	requireCDSError(
		t, serve(accountsRequest(t, accountsPath+"?page=3&page-size=1", "bank-only")),
		http.StatusUnprocessableEntity, cds.InvalidPage, "2",
	)
	requireCDSError(
		t, serve(accountsRequest(t, accountsPath+"?page-size=1001", "bank-only")),
		http.StatusBadRequest, cds.InvalidPageSize, "page-size",
	)
}

func TestGetAccountDetailGivesTheAccountWithItsNumbers(t *testing.T) {
	rec := serve(detailRequest(t, accountID, "account-detail"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "5", rec.Header().Get("x-v"))
	assert.JSONEq(t, `{
		"data": {
			"accountId": "`+accountID+`",
			"creationDate": "2026-10-15",
			"displayName": "Everyday",
			"maskedNumber": "xxx-xxx xxxx5678",
			"openStatus": "OPEN",
			"isOwned": true,
			"accountOwnership": "ONE_PARTY",
			"productCategory": "TRANS_AND_SAVINGS_ACCOUNTS",
			"productName": "Mock Everyday Account",
			"bsb": "123456",
			"accountNumber": "12345678"
		},
		"links": {"self": "https://mtls.bank.example/cds-au/v1/banking/accounts/`+accountID+`"},
		"meta": {}
	}`, rec.Body.String())
}

// Verifies: DH-0445
func TestGetAccountDetailOfAnAccountThatTheTokenDoesNotGiveIsAnInvalidBankingAccount(t *testing.T) {
	rec := serve(detailRequest(t, "internal-1", "account-detail"))

	requireCDSError(t, rec, http.StatusNotFound, cds.InvalidBankingAccount, "internal-1")
}

func TestTheAccountAPIsNeedTheirScopeAndVersion(t *testing.T) {
	requireCDSError(
		t, serve(accountsRequest(t, accountsPath, "account-detail")), http.StatusForbidden, cds.InvalidConsent,
		"The consent does not have the scope bank:accounts.basic:read.",
	)
	requireCDSError(
		t, serve(detailRequest(t, accountID, "bank-only")), http.StatusForbidden, cds.InvalidConsent,
		"The consent does not have the scope bank:accounts.detail:read.",
	)

	list := accountsRequest(t, accountsPath, "bank-only")
	list.Header.Set("x-v", "2")
	requireCDSError(
		t, serve(list), http.StatusNotAcceptable, cds.UnsupportedVersion, "The supported versions are 3 to 3.",
	)

	detail := detailRequest(t, accountID, "account-detail")
	detail.Header.Set("x-v", "4")
	requireCDSError(
		t, serve(detail), http.StatusNotAcceptable, cds.UnsupportedVersion, "The supported versions are 5 to 5.",
	)
}

func TestAFailureToReadTheAccountsGivesUnexpectedError(t *testing.T) {
	s := consentSecure()
	s.ConsentTokens = consentTokens{"other": {Scopes: []string{
		"bank:accounts.basic:read", "bank:accounts.detail:read",
	}, ArrangementID: "arrangement-9"}}

	for _, req := range []*http.Request{
		accountsRequest(t, accountsPath, "other"), detailRequest(t, accountID, "other"),
	} {
		rec := httptest.NewRecorder()
		secureMux(s).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), cds.Unexpected.Code)
	}
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var logs bytes.Buffer

	previous := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return &logs
}

func TestTheLogOfGetAccountDetailHasTheRouteAndNoAccountID(t *testing.T) {
	logs := captureLogs(t)

	serve(detailRequest(t, accountID, "account-detail"))
	serve(detailRequest(t, "internal-1", "account-detail"))

	assert.Contains(t, logs.String(), `"http.route":"GET `+accountsPath+`/{accountId}"`)
	assert.NotContains(t, logs.String(), accountID)
	assert.NotContains(t, logs.String(), "internal-1")
	assert.NotContains(t, logs.String(), "12345678")
}
