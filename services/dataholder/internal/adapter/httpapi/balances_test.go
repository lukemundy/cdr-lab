package httpapi_test

import (
	"cmp"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	balancePath  = accountsPath + "/" + accountID + "/balance"
	balancesPath = accountsPath + "/balances"
)

// arrangementBalances keeps the last query, and does not give an accountId that starts with "unknown".
type arrangementBalances struct {
	query app.AccountQuery
}

func consentBalance() app.ConsentBalance {
	return app.ConsentBalance{
		ID:      accountID,
		Balance: domain.Balance{Current: -120_050, Available: 29_950, CreditLimit: 150_000},
	}
}

func (a *arrangementBalances) Balance(_ context.Context, arrangementID, id string) (app.ConsentBalance, error) {
	if arrangementID != "arrangement-1" {
		return app.ConsentBalance{}, errors.New("database failure")
	}

	if id != accountID {
		return app.ConsentBalance{}, app.ErrAccountNotFound
	}

	return consentBalance(), nil
}

func (a *arrangementBalances) Balances(
	_ context.Context,
	arrangementID string,
	query app.AccountQuery,
) ([]app.ConsentBalance, int, error) {
	if arrangementID != "arrangement-1" {
		return nil, 0, errors.New("database failure")
	}

	a.query = query

	var unknown []string

	for _, id := range query.IDs {
		if strings.HasPrefix(id, "unknown") {
			unknown = append(unknown, id)
		}
	}

	if unknown != nil {
		return nil, 0, app.UnknownAccountsError{IDs: unknown}
	}

	return []app.ConsentBalance{consentBalance()}, 1, nil
}

func balancesSecure() (httpapi.Secure, *arrangementBalances) {
	balances := &arrangementBalances{}
	s := consentSecure()
	s.Balances = balances

	return s, balances
}

func serveBalances(t *testing.T, req *http.Request) (*httptest.ResponseRecorder, *arrangementBalances) {
	t.Helper()

	s, balances := balancesSecure()
	rec := httptest.NewRecorder()
	secureMux(s).ServeHTTP(rec, req)

	return rec, balances
}

func balanceRequest(t *testing.T, target, token string) *http.Request {
	t.Helper()

	return resourceRequest(t, http.MethodGet, target, token)
}

func bulkBalancesRequest(t *testing.T, target, token string) *http.Request {
	t.Helper()

	req := resourceRequest(t, http.MethodGet, target, token)
	req.Header.Set("x-v", "2")

	return req
}

func specificBalancesRequest(t *testing.T, target, body string) *http.Request {
	t.Helper()

	req := resourceRequest(t, http.MethodPost, target, "bank-only")
	req.Body = io.NopCloser(strings.NewReader(body))
	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Type", "application/json")

	return req
}

const balanceFields = `
	"accountId": "` + accountID + `",
	"currentBalance": "-1200.50",
	"availableBalance": "299.50",
	"creditLimit": "1500.00"`

func TestGetAccountBalanceGivesTheBalanceOfTheAccount(t *testing.T) {
	rec, _ := serveBalances(t, balanceRequest(t, balancePath, "bank-only"))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "1", rec.Header().Get("x-v"))
	assert.JSONEq(t, `{
		"data": {`+balanceFields+`},
		"links": {"self": "`+resourceBaseURL+balancePath+`"},
		"meta": {}
	}`, rec.Body.String())
}

func TestGetBulkBalancesGivesTheBalancesWithTheFiltersOfGetAccounts(t *testing.T) {
	rec, balances := serveBalances(
		t, bulkBalancesRequest(t, balancesPath+"?open-status=CLOSED&page-size=10", "bank-only"),
	)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "2", rec.Header().Get("x-v"))
	assert.Equal(t, app.AccountQuery{OpenStatus: domain.OpenStatusClosed, Page: 1, PageSize: 10}, balances.query)

	self := resourceBaseURL + balancesPath + "?open-status=CLOSED&page=1&page-size=10"
	assert.JSONEq(t, `{
		"data": {"balances": [{`+balanceFields+`}]},
		"links": {"self": "`+self+`", "first": "`+self+`", "last": "`+self+`"},
		"meta": {"totalRecords": 1, "totalPages": 1}
	}`, rec.Body.String())

	rec, _ = serveBalances(t, bulkBalancesRequest(t, balancesPath+"?is-owned=maybe", "bank-only"))
	requireCDSError(t, rec, http.StatusBadRequest, cds.FieldInvalid, "is-owned")
}

func TestGetBalancesForSpecificAccountsGivesTheBalancesOfTheAccountsInTheBody(t *testing.T) {
	rec, balances := serveBalances(t, specificBalancesRequest(
		t, balancesPath+"?page-size=10", `{"data": {"accountIds": ["`+accountID+`"]}, "meta": {}}`,
	))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "1", rec.Header().Get("x-v"))
	assert.Equal(t, app.AccountQuery{IDs: []string{accountID}, Page: 1, PageSize: 10}, balances.query)

	self := resourceBaseURL + balancesPath + "?page=1&page-size=10"
	assert.JSONEq(t, `{
		"data": {"balances": [{`+balanceFields+`}]},
		"links": {"self": "`+self+`", "first": "`+self+`", "last": "`+self+`"},
		"meta": {"totalRecords": 1, "totalPages": 1}
	}`, rec.Body.String())
}

func TestGetBalancesForSpecificAccountsAcceptsABodyThatTheSchemaAccepts(t *testing.T) {
	tests := map[string]struct {
		body        string
		contentType string
		want        []string
	}{
		"no accountIds": {
			body: `{"data": {"accountIds": []}}`, contentType: "application/json", want: []string{},
		},
		"unknown fields": {
			body:        `{"data": {"accountIds": ["a"], "other": 1}, "other": true}`,
			contentType: "application/json", want: []string{"a"},
		},
		"a charset": {
			body:        `{"data": {"accountIds": ["a"]}}`,
			contentType: "application/json; charset=utf-8", want: []string{"a"},
		},
		"upper case media type": {
			body: `{"data": {"accountIds": ["a"]}}`, contentType: "Application/JSON", want: []string{"a"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req := specificBalancesRequest(t, balancesPath, tt.body)
			req.Header.Set("Content-Type", tt.contentType)

			rec, balances := serveBalances(t, req)

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, tt.want, balances.query.IDs)
		})
	}
}

func TestGetBalancesForSpecificAccountsGivesAnErrorForEachAccountThatTheConsentDoesNotGive(t *testing.T) {
	rec, _ := serveBalances(t, specificBalancesRequest(
		t, balancesPath, `{"data": {"accountIds": ["unknown-1", "`+accountID+`", "unknown-2"]}}`,
	))

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	invalid := `"code": "` + cds.InvalidBankingAccount.Code + `", "title": "` + cds.InvalidBankingAccount.Title + `"`
	assert.JSONEq(t, `{"errors": [
		{`+invalid+`, "detail": "unknown-1"},
		{`+invalid+`, "detail": "unknown-2"}
	]}`, rec.Body.String())
}

func TestGetBalancesForSpecificAccountsRefusesABodyThatIsNotValid(t *testing.T) {
	tests := map[string]struct {
		body   string
		status int
		e      cds.Error
		detail string
	}{
		"no data":            {body: `{}`, e: cds.FieldMissing, detail: "data"},
		"null data":          {body: `{"data": null}`, e: cds.FieldMissing, detail: "data"},
		"no accountIds":      {body: `{"data": {}}`, e: cds.FieldMissing, detail: "data.accountIds"},
		"null accountIds":    {body: `{"data": {"accountIds": null}}`, e: cds.FieldMissing, detail: "data.accountIds"},
		"accountIds string":  {body: `{"data": {"accountIds": "a"}}`, e: cds.FieldInvalid, detail: "data.accountIds"},
		"accountId a number": {body: `{"data": {"accountIds": [1]}}`, e: cds.FieldInvalid, detail: "data.accountIds.0"},
		"data an array":      {body: `{"data": []}`, e: cds.FieldInvalid, detail: "data"},
		"body an array":      {body: `[]`, e: cds.FieldInvalid, detail: "data"},
		"not JSON":           {body: `hello`, e: cds.FieldInvalid, detail: "data"},
		"cut short":          {body: `{"data": `, e: cds.FieldInvalid, detail: "data"},
		"no body":            {body: ``, e: cds.FieldInvalid, detail: "data"},
		"two JSON values":    {body: `{"data": {"accountIds": []}} {}`, e: cds.FieldInvalid, detail: "data"},
		"too large": {
			body:   `{"data": {"accountIds": ["` + strings.Repeat("a", 64<<10) + `"]}}`,
			status: http.StatusRequestEntityTooLarge, e: cds.Expected,
			detail: "The request body must be 64 KiB or less.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec, _ := serveBalances(t, specificBalancesRequest(t, balancesPath, tt.body))

			requireCDSError(t, rec, cmp.Or(tt.status, http.StatusBadRequest), tt.e, tt.detail)
		})
	}
}

func TestGetBalancesForSpecificAccountsRefusesAContentTypeThatIsNotJSON(t *testing.T) {
	for _, contentType := range []string{"", "text/plain", "application/jsonl", "application/x-www-form-urlencoded"} {
		req := specificBalancesRequest(t, balancesPath, `{"data": {"accountIds": []}}`)
		req.Header.Set("Content-Type", contentType)

		rec, _ := serveBalances(t, req)

		requireCDSError(
			t, rec, http.StatusUnsupportedMediaType, cds.Expected, "The Content-Type must be application/json.",
		)
	}
}

func TestTheBalanceAPIsGiveInvalidBankingAccountForAnAccountOfThePath(t *testing.T) {
	rec, _ := serveBalances(t, balanceRequest(t, accountsPath+"/internal-1/balance", "bank-only"))

	requireCDSError(t, rec, http.StatusNotFound, cds.InvalidBankingAccount, "internal-1")
}

func TestTheBalanceAPIsNeedTheBasicAccountScopeAndTheirVersion(t *testing.T) {
	post := specificBalancesRequest(t, balancesPath, `{"data": {"accountIds": []}}`)
	post.Header.Set("Authorization", "Bearer transactions")

	for _, req := range []*http.Request{
		balanceRequest(t, balancePath, "transactions"),
		bulkBalancesRequest(t, balancesPath, "transactions"),
		post,
	} {
		rec, _ := serveBalances(t, req)

		requireCDSError(t, rec, http.StatusForbidden, cds.InvalidConsent,
			"The consent does not have the scope bank:accounts.basic:read.")
	}

	rec, _ := serveBalances(t, bulkBalancesRequest(t, balancesPath, "bank-only"))
	require.Equal(t, http.StatusOK, rec.Code)

	old := bulkBalancesRequest(t, balancesPath, "bank-only")
	old.Header.Set("x-v", "1")
	rec, _ = serveBalances(t, old)
	requireCDSError(t, rec, http.StatusNotAcceptable, cds.UnsupportedVersion, "The supported versions are 2 to 2.")
}

func TestAFailureToReadTheBalancesGivesUnexpectedError(t *testing.T) {
	s, _ := balancesSecure()
	s.ConsentTokens = consentTokens{
		"other": {Scopes: []string{"bank:accounts.basic:read"}, ArrangementID: "arrangement-9"},
	}

	post := specificBalancesRequest(t, balancesPath, `{"data": {"accountIds": []}}`)
	post.Header.Set("Authorization", "Bearer other")

	for _, req := range []*http.Request{
		balanceRequest(t, balancePath, "other"),
		bulkBalancesRequest(t, balancesPath, "other"),
		post,
	} {
		rec := httptest.NewRecorder()
		secureMux(s).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestTheLogOfTheBalanceAPIsHasTheRouteAndNoBalance(t *testing.T) {
	logs := captureLogs(t)

	serveBalances(t, balanceRequest(t, balancePath, "bank-only"))
	serveBalances(t, specificBalancesRequest(t, balancesPath, `{"data": {"accountIds": ["`+accountID+`"]}}`))

	assert.Contains(t, logs.String(), `"http.route":"GET `+accountsPath+`/{accountId}/balance"`)
	assert.Contains(t, logs.String(), `"http.route":"POST `+balancesPath+`"`)

	for _, value := range []string{accountID, "1200.50", "299.50", "1500.00"} {
		assert.NotContains(t, logs.String(), value)
	}
}
