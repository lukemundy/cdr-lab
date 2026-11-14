package consumerpages_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/consumerpages"
)

func serveStatic(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	consumerpages.RegisterStatic(mux, fstest.MapFS{
		"consumer.css":            {Data: []byte("body {}")},
		"fonts/inter-latin.woff2": {Data: []byte("wOF2")},
		"fonts/inter-LICENSE.txt": {Data: []byte("SIL Open Font License")},
		"notes.md":                {Data: []byte("# Notes")},
	})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))

	return rec
}

func TestTheStylesheetAndTheFontsOfThePagesAreOnTheStaticPath(t *testing.T) {
	tests := map[string]struct{ contentType, body string }{
		"/static/consumer.css":            {"text/css; charset=utf-8", "body {}"},
		"/static/fonts/inter-latin.woff2": {"font/woff2", "wOF2"},
		"/static/fonts/inter-LICENSE.txt": {"text/plain; charset=utf-8", "SIL Open Font License"},
	}
	for path, tt := range tests {
		t.Run(path, func(t *testing.T) {
			rec := serveStatic(t, path)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.contentType, rec.Header().Get("Content-Type"))
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
			assert.NotEmpty(t, rec.Header().Get("ETag"))
			assert.Equal(t, tt.body, rec.Body.String())
		})
	}
}

func TestTheStaticPathHasNoListAndNoFileOfAnotherType(t *testing.T) {
	for _, path := range []string{"/static/", "/static/fonts/", "/static/notes.md", "/static/missing.css"} {
		t.Run(path, func(t *testing.T) {
			assert.Equal(t, http.StatusNotFound, serveStatic(t, path).Code)
		})
	}
}
