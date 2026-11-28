package webapp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/services/datarecipient/internal/adapter/webapp"
)

const host = "localhost:8585"

var ui = fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>Sorrel</title>")}}

func serve(t *testing.T, ports webapp.Ports, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	require.NoError(t, webapp.Register(mux, ui, []string{host}, ports))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, "https://"+host+path, nil))

	return rec
}

// Verifies: DR-0208
func TestEachPathOutsideTheAPIGivesTheHTMLPageWithTheHeadersOfAWebUI(t *testing.T) {
	for _, path := range []string{"/", "/sign-up"} {
		rec := serve(t, webapp.Ports{}, http.MethodGet, path)

		require.Equal(t, http.StatusOK, rec.Code, path)
		assert.Contains(t, rec.Body.String(), "<title>Sorrel</title>", path)
		assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"), path)
	}
}

// Verifies: DR-0208
func TestAnUnknownAPIPathIsNotFoundAndNamesTheWebApp(t *testing.T) {
	rec := serve(t, webapp.Ports{}, http.MethodGet, "/api/members")

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "web app API operation")
}

func TestTheWebAppServesTheFilesOfTheBrand(t *testing.T) {
	ports := webapp.Ports{BrandFiles: func(routes *httproute.Routes) {
		routes.Handle("GET /brand/styles.css", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(":root:root {}"))
		}))
	}}

	rec := serve(t, ports, http.MethodGet, "/brand/styles.css")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, ":root:root {}", rec.Body.String())
}

// Verifies: DR-0206
func TestTheWebAppNeedsItsAllowedHosts(t *testing.T) {
	err := webapp.Register(http.NewServeMux(), ui, nil, webapp.Ports{})

	assert.ErrorContains(t, err, "WEB_APP_ALLOWED_HOSTS")
}

func TestAPlainGoBuildHasNoBuildOfTheUI(t *testing.T) {
	err := webapp.Register(http.NewServeMux(), fstest.MapFS{".gitkeep": {}}, []string{host}, webapp.Ports{})

	assert.ErrorContains(t, err, "index.html")
}
