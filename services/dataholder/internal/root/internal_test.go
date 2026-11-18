package root

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// The kubelet sends the IP address of the pod as the Host of a probe.
const podHost = "10.1.2.3:8080"

// internalListener gives the handler of the internal listener of a data holder with the console on.
func internalListener(t *testing.T, m mode) http.Handler {
	t.Helper()

	setOnly(t, m)
	t.Setenv("CONSOLE_ENABLED", "true")

	cfg, err := loadConfig(m)
	require.NoError(t, err)

	cfg.api.consoleUI = ui

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))
	// The readiness probe of a process with the APIs needs the statuses.
	require.NoError(t, statusstore.New(pool, nil).Save(t.Context(), map[string]domain.SoftwareProductStatus{}, `"v1"`))

	proc, err := newProcess(t.Context(), pool, cfg, m)
	require.NoError(t, err)

	for _, listener := range proc.listeners {
		if listener.Port == internalPort {
			require.Nil(t, listener.TLSConfig, "the internal listener is plain HTTP")

			return listener.Handler
		}
	}

	require.FailNow(t, "the process has no internal listener")

	return nil
}

func sendInternal(
	t *testing.T,
	handler http.Handler,
	method, host, path string,
	header http.Header,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, "http://"+host+path, nil)
	maps.Copy(req.Header, header)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func authorisedOperator() http.Header {
	return http.Header{"Authorization": {"Bearer test-operator-token"}}
}

func assertNotTheConsole(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	assert.Empty(t, rec.Header().Get("Content-Security-Policy"), "a header of the console")
	assert.NotContains(t, rec.Body.String(), "<!doctype html>", "the HTML page of the console")
}

// Verifies: DH-1310, DH-2622
func TestTheListenersHaveTheLabPorts(t *testing.T) {
	assert.Equal(t, []int{8080, 8081, 8083, 8085}, []int{internalPort, apiPort, securePort, bankSitePort})
}

// Verifies: DH-0017, DH-2623, SRV-0011, SRV-0012
func TestTheProbesNeedNoAllowedHostAndNoToken(t *testing.T) {
	handler := internalListener(t, serveMode)

	for _, path := range []string{"/livez", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			rec := sendInternal(t, handler, http.MethodGet, podHost, path, nil)

			assert.Equal(t, http.StatusOK, rec.Code)
			assertNotTheConsole(t, rec)
		})
	}
}

// Verifies: DH-0728, DH-2624, OPI-0005, OPI-0007
func TestTheOperatorInterfaceRefusesAHostThatIsNotAllowedAndHasNoHeadersOfTheConsole(t *testing.T) {
	handler := internalListener(t, serveMode)

	refused := sendInternal(t, handler, http.MethodGet, podHost, "/v1/customers/"+unknownCustomerNumber,
		authorisedOperator())
	assert.Equal(t, http.StatusForbidden, refused.Code)
	assert.Equal(t, "application/problem+json", refused.Header().Get("Content-Type"))
	assertNotTheConsole(t, refused)

	rec := sendInternal(t, handler, http.MethodGet, operatorHost, "/v1/customers/"+unknownCustomerNumber,
		authorisedOperator())
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	assertNotTheConsole(t, rec)

	for name := range rec.Header() {
		assert.False(t, strings.HasPrefix(name, "Access-Control-"), "header %s", name)
	}
}

// Verifies: DH-2628, OPI-0001
func TestEachPathUnderV1NeedsTheOperatorToken(t *testing.T) {
	handler := internalListener(t, serveMode)

	for _, path := range []string{"/v1/unknown", "/v1/customers/" + unknownCustomerNumber} {
		rec := sendInternal(t, handler, http.MethodGet, operatorHost, path, nil)

		assert.Equal(t, http.StatusUnauthorized, rec.Code, path)
		assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"), path)
		assertNotTheConsole(t, rec)
	}
}

// Verifies: DH-2627, DH-1482, DH-1483, DH-1485, CNSL-0004, CNSL-0007
func TestTheConsoleUsesTheAllowlistOfTheInternalListener(t *testing.T) {
	handler := internalListener(t, serveMode)

	for _, path := range []string{"/", "/customers", "/api/customers", "/brand/console.css"} {
		t.Run(path, func(t *testing.T) {
			refused := sendInternal(t, handler, http.MethodGet, podHost, path, nil)
			assert.Equal(t, http.StatusForbidden, refused.Code)
			assert.NotEmpty(t, refused.Header().Get("Content-Security-Policy"))

			rec := sendInternal(t, handler, http.MethodGet, operatorHost, path, nil)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
		})
	}
}

// Verifies: CNSL-0005
func TestTheConsoleRefusesAChangeFromADifferentOriginOnTheInternalListener(t *testing.T) {
	handler := internalListener(t, serveMode)

	rec := sendInternal(t, handler, http.MethodPost, operatorHost, "/api/customers",
		http.Header{"Sec-Fetch-Site": {"cross-site"}})

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// Verifies: CNSL-0013, SRV-0009
func TestTheConsoleCatchAllDoesNotCatchThePathsOfTheOtherGroups(t *testing.T) {
	handler := internalListener(t, serveMode)

	for _, tt := range []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/livez", http.StatusOK},
		{http.MethodGet, "/readyz", http.StatusOK},
		{http.MethodPost, "/livez", http.StatusMethodNotAllowed},
		{http.MethodDelete, "/readyz", http.StatusMethodNotAllowed},
		{http.MethodGet, "/v1/unknown", http.StatusUnauthorized},
		{http.MethodPost, "/v1/", http.StatusUnauthorized},
		{http.MethodPut, "/v1/customers", http.StatusUnauthorized},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := sendInternal(t, handler, tt.method, operatorHost, tt.path, nil)

			assert.Equal(t, tt.want, rec.Code)
			assertNotTheConsole(t, rec)
		})
	}
}

// Verifies: DH-2625, DH-1487
func TestAWorkerServesOnlyTheProbesOnTheInternalListener(t *testing.T) {
	handler := internalListener(t, workerMode)

	assert.Equal(t, http.StatusOK, sendInternal(t, handler, http.MethodGet, podHost, "/livez", nil).Code)
	assert.Equal(t, http.StatusOK, sendInternal(t, handler, http.MethodGet, podHost, "/readyz", nil).Code)

	for _, path := range []string{"/", "/v1/unknown", "/v1/customers/" + unknownCustomerNumber, "/api/customers"} {
		rec := sendInternal(t, handler, http.MethodGet, operatorHost, path, authorisedOperator())

		assert.Equal(t, http.StatusNotFound, rec.Code, path)
	}
}
