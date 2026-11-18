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
)

// The kubelet sends the IP address of the pod as the Host of a probe.
const podHost = "10.1.2.3:8080"

// internalListener gives the handler of the internal listener of a Register with the console on.
func internalListener(t *testing.T, m mode) http.Handler {
	t.Helper()

	setLabSettings(t)
	t.Setenv("CONSOLE_ENABLED", "true")

	cfg, err := loadConfig(m)
	require.NoError(t, err)

	cfg.consoleUI = ui

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	proc, err := newProcess(t.Context(), pool, cfg, m)
	require.NoError(t, err)

	handler, tlsConfig := listener(t, proc, internalPort)
	require.Nil(t, tlsConfig, "the internal listener is plain HTTP")

	return handler
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

func assertNotTheConsole(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	assert.Empty(t, rec.Header().Get("Content-Security-Policy"), "a header of the console")
	assert.NotContains(t, rec.Body.String(), "<!doctype html>", "the HTML page of the console")
}

// Verifies: REG-1500
func TestTheListenersHaveTheLabPorts(t *testing.T) {
	assert.Equal(t, []int{8080, 8081, 8083}, []int{internalPort, apiPort, securePort})
}

// Verifies: REG-0008, REG-1499, SRV-0011, SRV-0012, REG-0716, REG-0717, REG-1505
func TestTheProbesAndTheCRLNeedNoAllowedHostAndNoToken(t *testing.T) {
	handler := internalListener(t, serveMode)

	for _, path := range []string{"/livez", "/readyz", "/crl/cdr-intermediate.crl"} {
		t.Run(path, func(t *testing.T) {
			rec := sendInternal(t, handler, http.MethodGet, podHost, path, nil)

			assert.Equal(t, http.StatusOK, rec.Code)
			assertNotTheConsole(t, rec)
		})
	}
}

// Verifies: REG-1501, REG-0850, OPI-0005, OPI-0007
func TestTheOperatorInterfaceRefusesAHostThatIsNotAllowedAndHasNoHeadersOfTheConsole(t *testing.T) {
	handler := internalListener(t, serveMode)

	refused := sendInternal(t, handler, http.MethodGet, podHost, "/v1/cdr-ca-bundle", authorisedOperator())
	assert.Equal(t, http.StatusForbidden, refused.Code)
	assert.Equal(t, "application/problem+json", refused.Header().Get("Content-Type"))
	assertNotTheConsole(t, refused)

	rec := sendInternal(t, handler, http.MethodGet, operatorHost, "/v1/cdr-ca-bundle", authorisedOperator())
	assert.Equal(t, http.StatusOK, rec.Code)
	assertNotTheConsole(t, rec)

	for name := range rec.Header() {
		assert.False(t, strings.HasPrefix(name, "Access-Control-"), "header %s", name)
	}
}

// Verifies: REG-1508, OPI-0001
func TestEachPathUnderV1NeedsTheOperatorToken(t *testing.T) {
	handler := internalListener(t, serveMode)

	for _, path := range []string{"/v1/unknown", "/v1/cdr-ca-bundle"} {
		rec := sendInternal(t, handler, http.MethodGet, operatorHost, path, nil)

		assert.Equal(t, http.StatusUnauthorized, rec.Code, path)
		assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"), path)
		assertNotTheConsole(t, rec)
	}
}

// Verifies: REG-1507, REG-1063, CNSL-0004, CNSL-0007
func TestTheConsoleUsesTheAllowlistOfTheInternalListener(t *testing.T) {
	handler := internalListener(t, serveMode)

	for _, path := range []string{"/", "/data-recipients", "/api/data-recipients"} {
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

	rec := sendInternal(t, handler, http.MethodPost, operatorHost, "/api/data-recipients",
		http.Header{"Sec-Fetch-Site": {"cross-site"}})

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// Verifies: CNSL-0013, REG-0720, REG-0721, SRV-0009
func TestTheConsoleCatchAllDoesNotCatchThePathsOfTheOtherGroups(t *testing.T) {
	handler := internalListener(t, serveMode)

	for _, tt := range []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/livez", http.StatusOK},
		{http.MethodGet, "/readyz", http.StatusOK},
		{http.MethodPost, "/livez", http.StatusMethodNotAllowed},
		{http.MethodGet, "/v1/unknown", http.StatusUnauthorized},
		{http.MethodPost, "/v1/", http.StatusUnauthorized},
		{http.MethodGet, "/crl/unknown.crl", http.StatusNotFound},
		{http.MethodGet, "/crl/", http.StatusNotFound},
		{http.MethodPost, "/crl/cdr-intermediate.crl", http.StatusMethodNotAllowed},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := sendInternal(t, handler, tt.method, operatorHost, tt.path, nil)

			assert.Equal(t, tt.want, rec.Code)
			assertNotTheConsole(t, rec)
		})
	}
}

// Verifies: REG-1502
func TestAWorkerServesOnlyTheProbesOnTheInternalListener(t *testing.T) {
	handler := internalListener(t, workerMode)

	assert.Equal(t, http.StatusOK, sendInternal(t, handler, http.MethodGet, podHost, "/livez", nil).Code)

	for _, path := range []string{"/", "/v1/cdr-ca-bundle", "/crl/cdr-intermediate.crl", "/api/data-recipients"} {
		rec := sendInternal(t, handler, http.MethodGet, operatorHost, path, authorisedOperator())

		assert.Equal(t, http.StatusNotFound, rec.Code, path)
	}
}
