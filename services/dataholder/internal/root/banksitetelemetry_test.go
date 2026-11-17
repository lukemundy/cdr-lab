package root

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/disclosurestore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/eligibilityjobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/grantstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/productstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// The values are invented, with values that a span has no other reason to contain. The data holder creates the
// customer number 00058314 from the random value 58314, tries 00027182 for the sign-up with the email address in use,
// and makes the account numbers 71942068 and 63518204 of the opens from the random values 71942068 and 63518204. 2846,
// 2068 and 8204 are the ends of the masked numbers of the accounts. 48213.77 is the loan amount of the open of the
// loan, and the drawdown is -48213.77. The comparison has no difference for upper and lower case.
var siteValues = []string{
	"00058314", "58314", "bandicoot", "wallaby", "bilby@mail.test", "kookaburra", "correct-horse-battery",
	"00027182", "27182", "numbat.unknown@mail.test", "numbat.unknown",
	"account-numbat", "potoroo saver", "xxx-xxx xxxx2846", "echidna rent", "736291", "59172846",
	"arrangement-wombat", "376100", "71942068", "xxx-xxx xxxx2068", "account-quokka", "quokka-everyday",
	"quokka everyday", "dingo-loan", "platypus planner", "wombat works", "wombat works pty ltd",
	"account name, type and balance", "arrangement-possum", "platypus budget", "bettong brands",
	"63518204", "xxx-xxx xxxx8204", "account-dingo", "transaction-dingo", "48213.77", "4821377", "dingo loan",
}

// siteProducts gives a deposit product and a lending product for the opens of the site.
func siteProducts(t *testing.T, pool *pgxpool.Pool) *productstore.Store {
	t.Helper()

	products := productstore.New(pool)
	require.NoError(t, products.ImportProducts(t.Context(), []domain.Product{
		{
			ProductID: "quokka-everyday", LastUpdated: time.Now(),
			ProductCategory: domain.CategoryTransAndSavingsAccounts, Name: "Quokka Everyday",
			Description: "An account.", Brand: "Coastal Bank",
		},
		{
			ProductID: "dingo-loan", LastUpdated: time.Now(), ProductCategory: domain.CategoryPersLoans,
			Name: "Dingo Loan", Description: "A loan.", Brand: "Coastal Bank",
		},
	}))

	return products
}

// siteRequest gives the session cookie of the response, if any.
func siteRequest(
	t *testing.T, handler http.Handler, method, target, body string, status int, cookies ...*http.Cookie,
) []*http.Cookie {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, "https://example.com"+target, strings.NewReader(body))
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, status, rec.Code, "%s %s: %s", method, target, rec.Body.String())

	return rec.Result().Cookies()
}

func TestTheBankSitePutsNoCustomerDataSessionIDOrPasswordInTheSpansMetricsOrLogs(t *testing.T) {
	spans, metrics, logs := recordTelemetry(t)

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	inserter, err := jobs.NewInsertOnlyClient(pool)
	require.NoError(t, err)

	events := newEventStore(pool, inserter)
	store := customerstore.New(pool, eligibilityjobs.Hook(inserter), transactionstore.Insert)
	transactions := transactionstore.New(pool)
	reader := arrangementstore.NewQueries(pool)
	arrangements := app.Arrangements{
		Store: events, Grants: grantstore.New(pool), Customers: store, Reader: reader, Now: time.Now,
	}
	randoms := []uint64{58_314, 27_182, 71_942_068, 63_518_204}
	ids := []string{"account-quokka", "account-dingo", "transaction-dingo"}
	customers := app.Customers{
		Store: store, Arrangements: arrangements, Now: time.Now, Products: siteProducts(t, pool), BSB: "376100",
		NewID: func() string {
			next := ids[0]
			ids = ids[1:]

			return next
		},
		Random: func() uint64 {
			next := randoms[0]
			randoms = randoms[1:]

			return next
		},
	}
	customerData := app.CustomerData{
		Customers: store, TransactionStore: transactions, Arrangements: reader,
		Disclosures: disclosurestore.New(pool), Now: time.Now,
	}
	listeners, err := bankSiteListeners(testBankSite, tls.Certificate{}, ui, banksite.Ports{
		BrandName:  "Coastal Bank",
		BrandID:    "d7385786-d8b1-4d4b-bb80-b21616df1a4d",
		Customers:  customers,
		Opener:     customers,
		Accounts:   customerData,
		Consents:   customerData,
		Withdrawer: arrangements,
		Now:        time.Now,
	})
	require.NoError(t, err)

	handler := listeners[0].Handler
	setUpSpans := len(spans.Ended())

	signUp := siteRequest(t, handler, http.MethodPost, "/api/customers",
		`{"firstName": "Bandicoot", "lastName": "Wallaby", "emailAddress": "bilby@mail.test"}`, http.StatusCreated)
	siteRequest(t, handler, http.MethodPost, "/api/customers",
		`{"lastName": "Wallaby", "emailAddress": "kookaburra"}`, http.StatusBadRequest)
	siteRequest(t, handler, http.MethodPost, "/api/customers",
		`{"lastName": "Wallaby", "emailAddress": "BILBY@mail.test"}`, http.StatusConflict)
	siteRequest(t, handler, http.MethodPost, "/api/session",
		`{"identifier": "00058314", "password": "correct-horse-battery"}`, http.StatusBadRequest)
	siteRequest(t, handler, http.MethodPost, "/api/session",
		`{"identifier": "Numbat.Unknown@mail.test"}`, http.StatusBadRequest)
	numberSignIn := siteRequest(t, handler, http.MethodPost, "/api/session",
		`{"identifier": "00058314"}`, http.StatusNoContent, signUp...)
	signIn := siteRequest(t, handler, http.MethodPost, "/api/session",
		`{"identifier": "Bilby@Mail.test"}`, http.StatusNoContent, numberSignIn...)
	siteRequest(t, handler, http.MethodGet, "/api/session", "", http.StatusOK, signIn...)

	// The operator opens the account after the sign-up, so that the spans of the requests also have account data.
	// The consent makes the close of the account an end of eligibility, with a notice.
	addSiteAccount(t, store, transactions)
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "arrangement-wombat", domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-wombat", CustomerNumber: "00058314", ClientID: "client-1",
			SoftwareID: "software-1", ClientName: "Platypus Planner", OrgName: "Wombat Works",
			LegalEntityName: "Wombat Works Pty Ltd", Scopes: []string{"bank:accounts.basic:read"},
			AccountIDs: []string{"account-numbat"}, ConfirmedAt: time.Now(), SharingDurationSeconds: 3600,
		},
	}))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "arrangement-possum", domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-possum", CustomerNumber: "00058314", ClientID: "client-1",
			SoftwareID: "software-1", ClientName: "Platypus Budget", OrgName: "Bettong Brands",
			LegalEntityName: "Bettong Brands Pty Ltd", Scopes: []string{"bank:accounts.basic:read"},
			AccountIDs: []string{}, ConfirmedAt: time.Now(), SharingDurationSeconds: 3600,
		},
	}))

	// A call of an API of the consent records a disclosure, which the consents then show.
	consentData := app.ConsentData{
		Arrangements: events, Customers: store, TransactionStore: transactions, Now: time.Now,
	}
	consentAPI := http.NewServeMux()
	httpapi.RegisterSecure(consentAPI, httpapi.Secure{
		ResourceBaseURL:  "https://mtls.bank.test",
		RegistrationPath: "/register",
		UserInfoPath:     "/userinfo",
		ConsentTokens: arrangementToken{
			arrangementID: "arrangement-wombat",
		},
		ClientStatuses: activeClients{},
		Arrangements:   consentData,
		Customers:      consentData,
		Accounts:       consentData,
		Balances:       consentData,
		Transactions:   consentData,
		Disclosures:    app.Disclosures{Store: disclosurestore.New(pool), Now: time.Now},
	})

	getAccounts := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, "https://mtls.bank.test/cds-au/v1/banking/accounts", nil,
	)
	getAccounts.Header.Set("Authorization", "Bearer "+consentAccess)
	getAccounts.Header.Set("x-v", "3")
	getAccounts.Header.Set("x-fapi-auth-date", "Mon, 02 Nov 2026 08:00:00 GMT")

	disclosed := httptest.NewRecorder()
	telemetry.NewHTTPHandler(consentAPI).ServeHTTP(disclosed, getAccounts)
	require.Equal(t, http.StatusOK, disclosed.Code, disclosed.Body.String())
	siteRequest(t, handler, http.MethodGet, "/api/consents", "", http.StatusOK, signIn...)

	for _, target := range []string{
		"/api/accounts", "/api/accounts/account-numbat/balance", "/api/accounts/account-numbat/transactions",
	} {
		siteRequest(t, handler, http.MethodGet, target, "", http.StatusOK, signIn...)
	}

	siteRequest(t, handler, http.MethodGet, "/api/accounts/account-numbat/transactions?page=0", "",
		http.StatusBadRequest, signIn...)
	siteRequest(t, handler, http.MethodGet, "/api/accounts/account-numbats/balance", "",
		http.StatusNotFound, signIn...)
	siteRequest(t, handler, http.MethodPost, "/api/consents/arrangement-possum/withdraw", "",
		http.StatusNoContent, signIn...)
	siteRequest(t, handler, http.MethodPost, "/api/consents/arrangement-possum/withdraw", "",
		http.StatusConflict, signIn...)
	siteRequest(t, handler, http.MethodPost, "/api/consents/arrangement-possums/withdraw", "",
		http.StatusNotFound, signIn...)
	siteRequest(t, handler, http.MethodPost, "/api/accounts/account-numbat/close", `{"ends": []}`,
		http.StatusConflict, signIn...)
	siteRequest(t, handler, http.MethodPost, "/api/accounts/account-numbat/close", `{"ends": ["arrangement-wombat"]}`,
		http.StatusNoContent, signIn...)
	siteRequest(t, handler, http.MethodGet, "/api/accounts", "", http.StatusOK, signIn...)
	siteRequest(t, handler, http.MethodPost, "/api/accounts", `{"productId": "quokka-everyday"}`,
		http.StatusCreated, signIn...)
	siteRequest(t, handler, http.MethodPost, "/api/accounts", `{"productId": "dingo-loan", "loanAmount": "0.00"}`,
		http.StatusBadRequest, signIn...)
	siteRequest(t, handler, http.MethodPost, "/api/accounts", `{"productId": "dingo-loan", "loanAmount": "48213.77"}`,
		http.StatusCreated, signIn...)
	siteRequest(t, handler, http.MethodDelete, "/api/session", "", http.StatusNoContent, signIn...)
	siteRequest(t, handler, http.MethodGet, "/api/session", "", http.StatusForbidden, signIn...)

	values := append([]string{}, siteValues...)
	for _, cookie := range slices.Concat(signUp, numberSignIn, signIn) {
		values = append(values, strings.ToLower(cookie.Value))
	}

	require.Len(t, values, len(siteValues)+3, "the sign-up and each sign-in set the session cookie")

	routes := map[string]int{}

	for _, span := range spans.Ended()[setUpSpans:] {
		text := spanText(span)
		for _, value := range values {
			assert.NotContains(t, text, value, "span %q", span.Name())
		}

		for _, a := range span.Attributes() {
			if a.Key == semconv.HTTPRouteKey {
				routes[a.Value.AsString()]++
			}
		}
	}

	assert.Equal(t, map[string]int{
		"/api/customers": 3, "/api/session": 7, "/api/accounts": 5, "/api/accounts/{accountId}/balance": 2,
		"/api/accounts/{accountId}/transactions": 2, "/api/accounts/{accountId}/close": 2, "/api/consents": 1,
		"/cds-au/v1/banking/accounts": 1, "/api/consents/{arrangementId}/withdraw": 3,
	}, routes,
		"each server span has the route pattern")

	text := metricText(t, metrics)
	assert.Contains(t, text, "/api/session", "the test sees the metrics of the requests")

	for _, value := range values {
		assert.NotContains(t, text, value)
		assert.NotContains(t, strings.ToLower(logs.String()), value)
	}
}

func addSiteAccount(t *testing.T, store *customerstore.Store, transactions *transactionstore.Store) {
	t.Helper()

	require.NoError(t, store.AddAccount(t.Context(), "00058314", domain.Account{
		AccountID:       "account-numbat",
		DisplayName:     "Potoroo Saver",
		ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName:     "Mock Saver",
		BSB:             "736291",
		AccountNumber:   "59172846",
		OpenStatus:      domain.OpenStatusOpen,
		CreatedAt:       time.Now(),
	}))

	postedAt := time.Now().Add(-time.Hour)
	require.NoError(t, transactions.AddTransaction(t.Context(), "account-numbat", domain.Transaction{
		TransactionID: "transaction-platypus", Type: domain.TransactionPayment, Status: domain.TransactionPosted,
		Description: "Echidna Rent", Amount: -123_456, PostingDateTime: &postedAt,
	}))
}
