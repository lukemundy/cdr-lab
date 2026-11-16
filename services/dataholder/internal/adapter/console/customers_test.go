package console_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/openapi"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

var (
	ui  = fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>Data holder console</title>")}}
	now = time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
)

func loadSpec(t *testing.T) *openapi.Spec {
	t.Helper()

	data, err := os.ReadFile("openapi.yaml")
	require.NoError(t, err)
	spec, err := openapi.Load(data, "")
	require.NoError(t, err)

	return spec
}

type response struct {
	status   int
	body     string
	location string
}

func get(t *testing.T, ports console.Ports, target string) response {
	t.Helper()

	return do(t, ports, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
}

func post(t *testing.T, ports console.Ports, target, body string) response {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	return do(t, ports, req)
}

// do validates the response against the OpenAPI document.
func do(t *testing.T, ports console.Ports, req *http.Request) response {
	t.Helper()

	if ports.Now == nil {
		ports.Now = func() time.Time { return now }
	}

	mux := http.NewServeMux()
	require.NoError(t, console.Register(mux, ui, []string{"example.com"}, ports))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	resp.Request = req
	require.NoError(t, loadSpec(t).ValidateResponse(resp))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return response{status: resp.StatusCode, body: string(body), location: resp.Header.Get("Location")}
}

func newStore(t *testing.T) (*customerstore.Store, *eventstore.Store) {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return customerstore.New(pool, nil, nil), eventstore.New(pool, arrangementstore.Projections()...)
}

func addCustomer(t *testing.T, store *customerstore.Store, number, first, last string) {
	t.Helper()

	require.NoError(t, store.AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: number, FirstName: first, LastName: last, EmailAddress: number + "@mail.test",
		LastUpdated: time.Date(2026, 10, 15, 13, 30, 0, 123456000, time.UTC),
	}))
}

// Verifies: DH-1615
func TestTheDocumentClosesEachObjectSchema(t *testing.T) {
	assert.NoError(t, loadSpec(t).CheckClosed())
}

func TestTheBrandGivesTheBrandNameOfTheConfiguration(t *testing.T) {
	resp := get(t, console.Ports{BrandName: "Mock Bank"}, "/api/brand")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{"brandName": "Mock Bank"}`, resp.body)
}

func TestTheCustomersGiveOnePageWithTheCountsAndTheTotal(t *testing.T) {
	store, events := newStore(t)
	addCustomer(t, store, "00000001", "Alex", "Example")
	addCustomer(t, store, "00000002", "", "Sample")
	require.NoError(t, store.AddAccount(t.Context(), "00000001", domain.Account{
		AccountID: "account-1", DisplayName: "Everyday", ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName: "Mock Everyday Account", BSB: "123456", AccountNumber: "12345678",
		OpenStatus: domain.OpenStatusOpen, CreatedAt: now,
	}))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "arrangement-1", domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-1", CustomerNumber: "00000001", ClientID: "client-1",
			SoftwareID: "software-1", ClientName: "Recipient Web", Scopes: []string{"openid"},
			AccountIDs: []string{}, ConfirmedAt: now.Add(-time.Minute), SharingDurationSeconds: 3600,
		},
	}))

	resp := get(t, console.Ports{Customers: store}, "/api/customers?pageSize=1&page=1")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{"data": [{
		"customerNumber": "00000001", "firstName": "Alex", "lastName": "Example",
		"emailAddress": "00000001@mail.test", "lastUpdated": "2026-10-15T13:30:00.123456Z",
		"openAccountCount": 1, "currentArrangementCount": 1
	}], "totalRecords": 2}`, resp.body)

	resp = get(t, console.Ports{Customers: store}, "/api/customers?page=2&pageSize=1")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{"data": [{
		"customerNumber": "00000002", "lastName": "Sample", "emailAddress": "00000002@mail.test",
		"lastUpdated": "2026-10-15T13:30:00.123456Z", "openAccountCount": 0, "currentArrangementCount": 0
	}], "totalRecords": 2}`, resp.body)
}

// Verifies: DH-1628, DH-1629
func TestTheCustomersAreSearchedAndSortedByTheQuery(t *testing.T) {
	store, events := newStore(t)
	addCustomer(t, store, "00000001", "Alex", "Example")
	addCustomer(t, store, "00000002", "Robin", "Example")
	addCustomer(t, store, "00000003", "", "Other")
	require.NoError(t, store.AddAccount(t.Context(), "00000003", domain.Account{
		AccountID: "account-1", DisplayName: "Everyday", ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName: "Mock Everyday Account", BSB: "123456", AccountNumber: "12345678",
		OpenStatus: domain.OpenStatusOpen, CreatedAt: now,
	}))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "arrangement-1", domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-1", CustomerNumber: "00000002", ClientID: "client-1",
			SoftwareID: "software-1", ClientName: "Recipient Web", Scopes: []string{"openid"},
			AccountIDs: []string{}, ConfirmedAt: now.Add(-time.Minute), SharingDurationSeconds: 3600,
		},
	}))

	numbers := func(target string) []string {
		resp := get(t, console.Ports{Customers: store}, target)
		require.Equal(t, http.StatusOK, resp.status, resp.body)

		var list struct {
			Data []struct {
				CustomerNumber string `json:"customerNumber"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal([]byte(resp.body), &list))

		var got []string
		for _, c := range list.Data {
			got = append(got, c.CustomerNumber)
		}

		return got
	}

	assert.Equal(t, []string{"00000001", "00000002", "00000003"}, numbers("/api/customers"))
	assert.Equal(t, []string{"00000001", "00000002"}, numbers("/api/customers?q=EXAMPLE"))
	assert.Equal(t, []string{"00000002"}, numbers("/api/customers?q=robin"))
	assert.Equal(t, []string{"00000003", "00000002", "00000001"},
		numbers("/api/customers?sort=customerNumber&order=desc"))
	assert.Equal(t, []string{"00000003", "00000001", "00000002"},
		numbers("/api/customers?sort=lastName&order=desc"))
	assert.Equal(t, []string{"00000001", "00000002", "00000003"}, numbers("/api/customers?sort=lastUpdated"))
	assert.Equal(t, []string{"00000003", "00000001", "00000002"},
		numbers("/api/customers?sort=openAccountCount&order=desc"))
	assert.Equal(t, []string{"00000002", "00000001", "00000003"},
		numbers("/api/customers?sort=currentArrangementCount&order=desc"))
	// An empty value is no value, and a parameter that the API does not know is ignored.
	assert.Equal(t, []string{"00000001", "00000002", "00000003"},
		numbers("/api/customers?q=&sort=&order=&page=&pageSize=&status=CLOSED"))
}

func TestAQueryParameterThatIsNotValidGivesBadRequestWithNoValue(t *testing.T) {
	store, _ := newStore(t)

	for _, target := range []string{
		"/api/customers?sort=zebracorn",
		"/api/customers?order=zebracorn",
		"/api/customers?page=zebracorn",
		"/api/customers?page=0",
		"/api/customers?pageSize=101",
	} {
		resp := get(t, console.Ports{Customers: store}, target)

		require.Equal(t, http.StatusBadRequest, resp.status, target)
		assert.NotContains(t, resp.body, "zebracorn", target)
	}
}
