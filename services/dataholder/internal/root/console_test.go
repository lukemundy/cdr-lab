package root

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
)

var ui = fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>Data holder console</title>")}}

// The host of a request from httptest.NewRequestWithContext.
var testConsole = consolehttp.Config{Enabled: true, AllowedHosts: []string{"example.com"}}

// Verifies: DH-1481
func TestWithoutTheSettingTheDataHolderOpensNoConsoleListener(t *testing.T) {
	listeners, err := consoleListeners(consolehttp.Config{}, ui, console.Ports{})

	require.NoError(t, err)
	assert.Empty(t, listeners)
}

func TestWithTheSettingTheConsolePortGivesTheHTMLPage(t *testing.T) {
	listeners, err := consoleListeners(testConsole, ui, console.Ports{})
	require.NoError(t, err)
	require.Len(t, listeners, 1)
	assert.Equal(t, 8084, listeners[0].Port)
	assert.Nil(t, listeners[0].TLSConfig)

	rec := httptest.NewRecorder()
	listeners[0].Handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/customers", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "<!doctype html><title>Data holder console</title>", rec.Body.String())
	assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
}

// Verifies: DH-1482
func TestTheConsolePortRefusesAHostThatIsNotAllowed(t *testing.T) {
	listeners, err := consoleListeners(testConsole, ui, console.Ports{})
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://attacker.example:8084/", nil)
	rec := httptest.NewRecorder()
	listeners[0].Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestTheConsoleDoesNotStartWithoutAUIBuild(t *testing.T) {
	_, err := consoleListeners(testConsole, fstest.MapFS{".gitkeep": {}}, console.Ports{})

	assert.ErrorContains(t, err, "index.html")
}

func TestTheAPIProcessOpensNoConsoleListenerWithoutTheSetting(t *testing.T) {
	lab := newOperatorLab(t)

	assert.NotContains(t, lab.handlers, consolePort)
}
