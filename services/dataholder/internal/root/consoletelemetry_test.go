package root

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/overviewstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// The customer is invented. Each search is a part of one field of the customer.
const (
	searchNumber = "47291"
	searchName   = "zebracorn"
	searchEmail  = "quokka@"
)

// The account and the transaction are also invented, with values that a span has no other reason to contain.
var customerValues = []string{
	"00047291", searchNumber, searchName, "quagga", searchEmail, "quokka@mail.test",
	"account-wombat", "numbat saver", "98765432", "transaction-platypus", "echidna rent", "1234.56",
	"arrangement-bilby", "00058314", "58314", "wallaby", "bandicoot", "bilby@mail.test", "kookaburra",
	"account-possum", "possum", "glider", "314159", "27182818", "14142", "1732",
	"transaction-dingo", "possum groceries", "87.65", "galah mart", "cockatoo", "wombat-2", "lyrebird",
	"00058315", "58315",
}

// assertNoValue fails when the text contains the value. A digit next to the value does not count, because a timestamp
// or a duration can contain the digits of a short number.
func assertNoValue(t assert.TestingT, text, value string, msgAndArgs ...any) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}

	pattern := regexp.MustCompile(`(^|[^0-9])` + regexp.QuoteMeta(value) + `([^0-9]|$)`)
	assert.NotRegexp(t, pattern, text, msgAndArgs...)
}

type failures struct{ count int }

func (f *failures) Errorf(string, ...any) { f.count++ }

func TestADigitNextToAValueIsNotTheValue(t *testing.T) {
	t.Parallel()

	for text, want := range map[string]int{
		`{"time":"2026-11-23t06:06:27.819817327z"}`: 0,
		`{"accountNumber":"1732"}`:                  1,
		`{"accountNumber":"1732a"}`:                 1,
		"1732":                                      1,
	} {
		var f failures

		assertNoValue(&f, text, "1732")

		assert.Equal(t, want, f.count, text)
	}
}

// The body that gives 409 has the bsb and accountNumber of account-wombat.
func createAccounts(t *testing.T, handler http.Handler) {
	t.Helper()

	for body, status := range map[string]int{
		`{"displayName": "Possum Everyday", "productCategory": "TRANS_AND_SAVINGS_ACCOUNTS", ` +
			`"productName": "Mock Glider", "bsb": "314159", "accountNumber": "27182818"}`: http.StatusCreated,
		`{"displayName": "", "productCategory": "POSSUM", "productName": "Mock Glider", ` +
			`"bsb": "14142", "accountNumber": "1732a"}`: http.StatusBadRequest,
		`{"displayName": "Possum Copy", "productCategory": "TRAVEL_CARDS", "productName": "Mock Glider", ` +
			`"bsb": "123456", "accountNumber": "98765432"}`: http.StatusConflict,
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(
			t.Context(), http.MethodPost, "/api/customers/00047291/accounts", strings.NewReader(body),
		))
		require.Equal(t, status, rec.Code, rec.Body.String())
	}
}

// randomsFrom gives first, and then the next number at each call.
func randomsFrom(first uint64) func() uint64 {
	next := first

	return func() uint64 {
		next++

		return next - 1
	}
}

// newIDs gives the IDs in their sequence, one for each create, and then the last ID again.
func newIDs(ids ...string) func() string {
	next := 0

	return func() string {
		id := ids[min(next, len(ids)-1)]
		next++

		return id
	}
}

// The create that succeeds gives the transaction ID transaction-dingo. account-wombat is closed.
func createTransactions(t *testing.T, handler http.Handler) {
	t.Helper()

	transaction := `{"type": "PAYMENT", "status": "POSTED", "description": "Possum Groceries", "amount": "-87.65",
		"reference": "Cockatoo", "postingDateTime": "2026-10-20T09:30:00+08:00", "merchantName": "Galah Mart",
		"extendedData": {"payee": "Galah Mart", "extensionUType": "nppPayload",
			"nppPayload": {"extendedDescription": "Lyrebird", "service": "X2P1", "serviceVersion": "03"}}}`
	refused := strings.Replace(transaction, `"service": "X2P1"`, `"service": "Lyrebird"`, 1)

	for _, c := range []struct {
		path, body string
		status     int
	}{
		{"/api/customers/00047291/accounts/account-wombat-2/transactions", transaction, http.StatusCreated},
		{"/api/customers/00047291/accounts/account-wombat-2/transactions", refused, http.StatusBadRequest},
		{"/api/customers/00047291/accounts/account-wombat/transactions", transaction, http.StatusConflict},
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(
			t.Context(), http.MethodPost, c.path, strings.NewReader(c.body),
		))
		require.Equal(t, c.status, rec.Code, rec.Body.String())
	}
}

// The create that succeeds gives the customer number 00058314.
func createCustomers(t *testing.T, handler http.Handler) {
	t.Helper()

	for body, status := range map[string]int{
		`{"firstName": "Wallaby", "lastName": "Bandicoot", "emailAddress": "bilby@mail.test"}`: http.StatusCreated,
		`{"firstName": "Wallaby", "lastName": "", "emailAddress": "kookaburra at mail.test"}`:  http.StatusBadRequest,
		`{"firstName": "Wallaby", "lastName": "Galah", "emailAddress": "QUOKKA@Mail.test"}`:    http.StatusConflict,
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(
			t.Context(), http.MethodPost, "/api/customers", strings.NewReader(body),
		))
		require.Equal(t, status, rec.Code, rec.Body.String())
	}
}

type noGrants struct{}

func (noGrants) RevokeArrangement(context.Context, string) error { return nil }

// recordTelemetry has no URL redactor, so the test is stricter than the telemetry of the data holder.
func recordTelemetry(t *testing.T) (*tracetest.SpanRecorder, *sdkmetric.ManualReader, *bytes.Buffer) {
	t.Helper()

	spans := tracetest.NewSpanRecorder()
	metrics := sdkmetric.NewManualReader()
	previousTracer, previousMeter := otel.GetTracerProvider(), otel.GetMeterProvider()

	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)))
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(metrics)))

	var logs bytes.Buffer

	previousLogger := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))

	t.Cleanup(func() {
		otel.SetTracerProvider(previousTracer)
		otel.SetMeterProvider(previousMeter)
		slog.SetDefault(previousLogger)
	})

	return spans, metrics, &logs
}

// spanText has no url.path, because recordTelemetry has no URL redactor to set it to the route.
func spanText(span sdktrace.ReadOnlySpan) string {
	parts := []string{span.Name(), span.Status().Description}

	for _, a := range span.Attributes() {
		if a.Key != semconv.URLPathKey {
			parts = append(parts, a.Value.String())
		}
	}

	for _, e := range span.Events() {
		parts = append(parts, e.Name)
		for _, a := range e.Attributes {
			parts = append(parts, a.Value.String())
		}
	}

	return strings.ToLower(strings.Join(parts, "\n"))
}

func metricText(t *testing.T, reader *sdkmetric.ManualReader) string {
	t.Helper()

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))

	var parts []string

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			parts = append(parts, m.Name)

			switch data := m.Data.(type) {
			case metricdata.Histogram[float64]:
				for _, dp := range data.DataPoints {
					for _, kv := range dp.Attributes.ToSlice() {
						parts = append(parts, kv.Value.String())
					}
				}
			case metricdata.Sum[int64]:
				for _, dp := range data.DataPoints {
					for _, kv := range dp.Attributes.ToSlice() {
						parts = append(parts, kv.Value.String())
					}
				}
			default:
			}
		}
	}

	return strings.ToLower(strings.Join(parts, "\n"))
}

// Verifies: DH-1488, DH-1489
func TestTheConsolePutsNoCustomerAccountOrTransactionDataInTheSpansMetricsOrLogs(t *testing.T) {
	spans, metrics, logs := recordTelemetry(t)

	// The pool gets the tracer provider when it opens.
	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	store := customerstore.New(pool, nil, nil)
	require.NoError(t, store.AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: "00047291", FirstName: "Zebracorn", LastName: "Quagga", EmailAddress: "quokka@mail.test",
		LastUpdated: time.Now(),
	}))
	require.NoError(t, store.AddAccount(t.Context(), "00047291", domain.Account{
		AccountID:       "account-wombat",
		DisplayName:     "Numbat Saver",
		ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName:     "Mock Saver",
		BSB:             "123456",
		AccountNumber:   "98765432",
		OpenStatus:      domain.OpenStatusOpen,
		CreatedAt:       time.Now(),
	}))

	require.NoError(t, store.AddAccount(t.Context(), "00047291", domain.Account{
		AccountID:       "account-wombat-2",
		DisplayName:     "Numbat Saver 2",
		ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName:     "Mock Saver",
		BSB:             "123456",
		AccountNumber:   "98765433",
		OpenStatus:      domain.OpenStatusOpen,
		CreatedAt:       time.Now(),
	}))

	transactions := transactionstore.New(pool)
	postedAt := time.Now()
	require.NoError(t, transactions.AddTransaction(t.Context(), "account-wombat", domain.Transaction{
		TransactionID: "transaction-platypus", Type: domain.TransactionPayment, Status: domain.TransactionPosted,
		Description: "Echidna Rent", Amount: -123_456, PostingDateTime: &postedAt,
	}))

	events := eventstore.New(pool, append(registrationstore.Projections(), arrangementstore.Projections()...)...)
	require.NoError(t, app.ClientRegistrationType.Execute(t.Context(), events, "software-1", domain.RegisterClient{
		Registration: domain.Registration{ClientID: "client-1", SoftwareID: "software-1", ClientName: "Recipient Web"},
	}))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "arrangement-bilby", domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-bilby", CustomerNumber: "00047291", ClientID: "client-1",
			SoftwareID: "software-1", ClientName: "Recipient Web", Scopes: []string{"openid"},
			AccountIDs: []string{"account-wombat"}, ConfirmedAt: time.Now(), SharingDurationSeconds: 3600,
		},
	}))

	consoleRoutes, err := testConsoleRoutes(ui, console.Ports{
		BrandName: "Coastal Bank",
		Customers: store,
		Details: app.CustomerDetails{
			Customers: store, Transactions: transactions, Arrangements: arrangementstore.NewQueries(pool),
			Registrations: registrationstore.NewQueries(pool),
		},
		Registrations:      registrationstore.NewQueries(pool),
		ClientArrangements: arrangementstore.NewQueries(pool),
		Accounts:           app.Customers{Store: store, Now: time.Now},
		Creates: app.Customers{
			Store: store, Transactions: transactions, Now: time.Now, Random: randomsFrom(58_314),
			NewID: newIDs("account-possum", "transaction-dingo"),
		},
		Arrangements: app.Arrangements{
			Store:     events,
			Grants:    noGrants{},
			Customers: store,
			Reader:    arrangementstore.NewQueries(pool),
			Now:       time.Now,
		},
		Overview: overviewstore.New(pool),
		Now:      time.Now,
	})
	require.NoError(t, err)

	// The spans of the set-up also contain the table names, so only the spans of the requests are counted.
	setUpSpans := len(spans.Ended())

	targets := map[string]int{"/api/brand": http.StatusOK, "/api/customers": http.StatusOK}
	for _, search := range []string{searchNumber, searchName, strings.ToUpper(searchName), searchEmail} {
		targets["/api/customers?q="+search] = http.StatusOK
		targets["/api/customers?q="+search+"&sort=customerNumber&order=desc&page=1"] = http.StatusOK
		targets["/api/customers?q="+search+"&page=0"] = http.StatusBadRequest
		targets["/customers?q="+search] = http.StatusOK
	}

	targets["/api/customers/00047291"] = http.StatusOK
	targets["/api/customers/00047292"] = http.StatusNotFound
	targets["/api/customers/00047291/accounts/account-wombat/transactions?page=1"] = http.StatusOK
	targets["/api/customers/00047291/accounts/account-wombat/transactions?page=0"] = http.StatusBadRequest
	targets["/api/customers/00047291/accounts/account-wombats/transactions"] = http.StatusNotFound
	targets["/customers/00047291"] = http.StatusOK
	targets["/customers/00047291/accounts/account-wombat/transactions"] = http.StatusOK
	targets["/api/registrations"] = http.StatusOK
	targets["/api/registrations?sort=currentArrangementCount&order=desc"] = http.StatusOK
	targets["/api/registrations/client-1"] = http.StatusOK
	// The arrangements of the client give the customer number and the name of each customer.
	targets["/api/registrations/client-1/arrangements?page=1"] = http.StatusOK
	targets["/api/registrations/client-wombat"] = http.StatusNotFound
	// The overview gives the customer number of each change.
	targets["/api/overview"] = http.StatusOK

	for target, status := range targets {
		rec := httptest.NewRecorder()
		consoleRoutes.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
		require.Equal(t, status, rec.Code, target)
	}

	// A close that is not the last, then the same close again, which the data holder refuses.
	closePath := "/api/customers/00047291/accounts/account-wombat/close"

	for _, status := range []int{http.StatusNoContent, http.StatusConflict} {
		rec := httptest.NewRecorder()
		consoleRoutes.ServeHTTP(
			rec,
			httptest.NewRequestWithContext(t.Context(), http.MethodPost, closePath, strings.NewReader(`{"ends": []}`)),
		)
		require.Equal(t, status, rec.Code, closePath)
	}

	// A withdrawal, then the same withdrawal again, which the data holder refuses, and a withdrawal for no customer.
	withdrawals := []struct {
		path   string
		status int
	}{
		{"/api/customers/00047291/arrangements/arrangement-bilby/withdraw", http.StatusNoContent},
		{"/api/customers/00047291/arrangements/arrangement-bilby/withdraw", http.StatusConflict},
		{"/api/customers/00047292/arrangements/arrangement-bilby/withdraw", http.StatusNotFound},
	}

	for _, w := range withdrawals {
		rec := httptest.NewRecorder()
		consoleRoutes.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, w.path, nil))
		require.Equal(t, w.status, rec.Code, w.path)
	}

	createCustomers(t, consoleRoutes)
	createAccounts(t, consoleRoutes)
	createTransactions(t, consoleRoutes)

	routes, customerQueries := map[string]int{}, 0

	for _, span := range spans.Ended()[setUpSpans:] {
		text := spanText(span)
		for _, value := range customerValues {
			assertNoValue(t, text, value, "span %q", span.Name())
		}

		for _, a := range span.Attributes() {
			if a.Key == semconv.HTTPRouteKey {
				routes[a.Value.AsString()]++
			}
		}

		if strings.Contains(text, "from customers") {
			customerQueries++
		}
	}

	assert.Equal(t, map[string]int{
		"/api/brand": 1, "/api/overview": 1, "/api/customers": 16, "/api/customers/{customerNumber}": 2,
		"/api/customers/{customerNumber}/accounts/{accountId}/transactions": 6, "/": 6,
		"/api/registrations": 2, "/api/registrations/{clientId}": 2,
		"/api/registrations/{clientId}/arrangements":                            1,
		"/api/customers/{customerNumber}/accounts/{accountId}/close":            2,
		"/api/customers/{customerNumber}/arrangements/{arrangementId}/withdraw": 3,
		"/api/customers/{customerNumber}/accounts":                              3,
	}, routes,
		"each server span has the route pattern")
	assert.Positive(t, customerQueries, "the test sees the database spans of the search")

	text := metricText(t, metrics)
	assert.NotContains(t, text, "/api/customers", "the requests make no metrics, because their spans give the detail")

	for _, value := range customerValues {
		assertNoValue(t, text, value)
		assertNoValue(t, strings.ToLower(logs.String()), value)
	}
}
