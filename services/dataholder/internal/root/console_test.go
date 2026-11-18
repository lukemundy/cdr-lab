package root

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
)

var ui = fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>Data holder console</title>")}}

// The host of a request from httptest.NewRequestWithContext.
var testConsole = consolehttp.Config{Enabled: true, AllowedHosts: []string{"example.com"}}

// testConsoleRoutes gives the console routes of the internal listener, with no other group.
func testConsoleRoutes(ui fs.FS, ports console.Ports) (http.Handler, error) {
	mux := http.NewServeMux()
	if err := registerConsole(mux, testConsole, ui, ports); err != nil {
		return nil, err
	}

	return telemetry.NewHTTPHandler(mux), nil
}

// Verifies: DH-1481
func TestWithoutTheSettingTheInternalListenerHasNoConsole(t *testing.T) {
	mux := http.NewServeMux()
	require.NoError(t, registerConsole(mux, consolehttp.Config{}, ui, console.Ports{}))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestWithTheSettingTheConsoleGivesTheHTMLPage(t *testing.T) {
	consoleRoutes, err := testConsoleRoutes(ui, console.Ports{})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	consoleRoutes.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/customers", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "<!doctype html><title>Data holder console</title>", rec.Body.String())
	assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
}

// Verifies: DH-1482
func TestTheConsoleRefusesAHostThatIsNotAllowed(t *testing.T) {
	consoleRoutes, err := testConsoleRoutes(ui, console.Ports{})
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://attacker.example:8080/", nil)
	rec := httptest.NewRecorder()
	consoleRoutes.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestTheConsoleDoesNotStartWithoutAUIBuild(t *testing.T) {
	_, err := testConsoleRoutes(fstest.MapFS{".gitkeep": {}}, console.Ports{})

	assert.ErrorContains(t, err, "index.html")
}

func TestTheAPIProcessServesNoConsoleWithoutTheSetting(t *testing.T) {
	lab := newOperatorLab(t)

	rec := httptest.NewRecorder()
	lab.handlers[internalPort].ServeHTTP(
		rec,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+operatorHost+"/customers", nil),
	)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
