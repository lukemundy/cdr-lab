package console_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
)

func serveBrandFile(t *testing.T, host string) *httptest.ResponseRecorder {
	t.Helper()

	ports := console.Ports{BrandFiles: func(routes *httproute.Routes) {
		routes.Handle("GET /brand/console.css", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = io.WriteString(w, ":root:root {}\n")
		}))
	}}

	mux := http.NewServeMux()
	require.NoError(t, console.Register(mux, ui, []string{"example.com"}, ports))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+host+"/brand/console.css", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

// The CSP of the console permits images and styles from the same origin only (DESIGN.md: Listener and security).
func TestTheConsoleServesTheBrandFilesWithTheResponseHeadersOfTheConsole(t *testing.T) {
	rec := serveBrandFile(t, "example.com")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, ":root:root {}\n", rec.Body.String())
	assert.True(t, strings.HasPrefix(rec.Header().Get("Content-Security-Policy"), "default-src 'self';"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
}

func TestTheConsoleRefusesABrandFileForAHostThatIsNotAllowed(t *testing.T) {
	rec := serveBrandFile(t, "rebound.test")

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.NotContains(t, rec.Body.String(), ":root")
}
