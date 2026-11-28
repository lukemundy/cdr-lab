package root

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
)

var ui = fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>Web app</title>")}}

func testListeners(t *testing.T) (internal, web httpserver.Listener) {
	t.Helper()

	setServe(t)

	cfg, err := loadConfig()
	require.NoError(t, err)

	pair, err := tls.LoadX509KeyPair(cfg.webAppCert.cert, cfg.webAppCert.key)
	require.NoError(t, err)
	cert, err := cds.NewServerCertificate(t.Context(), "web app", pair)
	require.NoError(t, err)

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	got, err := listeners(pool, cfg, cert, ui)
	require.NoError(t, err)
	require.Len(t, got, 2)

	return got[0], got[1]
}

func send(t *testing.T, listener httpserver.Listener, host, path string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	listener.Handler.ServeHTTP(
		rec,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+host+path, nil),
	)

	return rec
}

// Verifies: DR-0014
func TestTheInternalListenerIsPlainHTTPAndServesTheHealthProbes(t *testing.T) {
	internal, _ := testListeners(t)

	assert.Equal(t, 8080, internal.Port)
	assert.Nil(t, internal.TLSConfig)

	// The kubelet sends the IP address of the pod as the Host of a probe.
	for _, path := range []string{"/livez", "/readyz"} {
		assert.Equal(t, http.StatusOK, send(t, internal, "10.1.2.3:8080", path).Code, path)
	}

	assert.Equal(
		t,
		http.StatusNotFound,
		send(t, internal, "10.1.2.3:8080", "/").Code,
		"the internal listener has no web app",
	)
}

// Verifies: DR-0203, DR-0204, DR-0205
func TestTheWebAppListenerHasTLSAndServesOnlyTheWebApp(t *testing.T) {
	_, web := testListeners(t)

	assert.Equal(t, 8085, web.Port)
	require.NotNil(t, web.TLSConfig)
	assert.Equal(t, tls.NoClientCert, web.TLSConfig.ClientAuth)
	assert.Equal(t, uint16(tls.VersionTLS12), web.TLSConfig.MinVersion)

	page := send(t, web, webAppHost, "/")
	require.Equal(t, http.StatusOK, page.Code)
	assert.Contains(t, page.Body.String(), "<title>Web app</title>")

	// The catch-all of the web app gives its HTML page, not a probe.
	assert.Contains(t, send(t, web, webAppHost, "/readyz").Body.String(), "<title>Web app</title>")
}

// Verifies: DR-0206
func TestTheWebAppRefusesAHostThatIsNotAllowed(t *testing.T) {
	_, web := testListeners(t)

	assert.Equal(t, http.StatusForbidden, send(t, web, "attacker.example:8585", "/").Code)
}

// Verifies: DR-0219, DR-0222
func TestTheWebAppServesTheStylesheetAndTheMarkOfTheBrand(t *testing.T) {
	_, web := testListeners(t)

	styles := send(t, web, webAppHost, "/brand/styles.css")
	require.Equal(t, http.StatusOK, styles.Code)
	assert.Equal(t, "text/css; charset=utf-8", styles.Header().Get("Content-Type"))
	assert.Contains(t, styles.Body.String(), "--ds-color-brand")

	mark := send(t, web, webAppHost, "/brand/mark.svg")
	require.Equal(t, http.StatusOK, mark.Code)
	assert.Equal(t, "image/svg+xml", mark.Header().Get("Content-Type"))
}
