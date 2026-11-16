package asset_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/asset"
)

func get(t *testing.T, handler http.Handler, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/file", nil)
	req.Header = header

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func TestTheHandlerGivesTheDataWithItsTypeAndNoSniff(t *testing.T) {
	rec := get(t, asset.Handler("text/css; charset=utf-8", []byte("body {}"), ""), http.Header{})

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/css; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Empty(t, rec.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "body {}", rec.Body.String())
}

func TestTheHandlerSetsThePolicy(t *testing.T) {
	rec := get(t, asset.Handler("image/svg+xml", []byte("<svg/>"), "sandbox"), http.Header{})

	assert.Equal(t, "sandbox", rec.Header().Get("Content-Security-Policy"))
}

// Verifies: ASSET-0001
func TestABrowserCanUseItsCopyUntilTheDataChanges(t *testing.T) {
	handler := asset.Handler("text/plain", []byte("one"), "")

	first := get(t, handler, http.Header{})
	etag := first.Header().Get("ETag")

	require.NotEmpty(t, etag)
	assert.Equal(t, "no-cache", first.Header().Get("Cache-Control"))
	assert.Equal(t, http.StatusNotModified, get(t, handler, http.Header{"If-None-Match": {etag}}).Code)

	changed := get(t, asset.Handler("text/plain", []byte("two"), ""), http.Header{"If-None-Match": {etag}})

	assert.Equal(t, http.StatusOK, changed.Code)
}
