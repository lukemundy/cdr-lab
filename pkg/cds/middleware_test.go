package cds_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

func run(t *testing.T, h http.Handler) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	return rec
}

func noContent(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }

// Verifies: CDSL-0037, CDSL-0038
func TestAllowAnyOriginSetsOnlyTheCORSHeaders(t *testing.T) {
	rec := run(t, cds.AllowAnyOrigin(http.HandlerFunc(noContent), "x-v", "x-fapi-interaction-id"))

	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "x-v, x-fapi-interaction-id", rec.Header().Get("Access-Control-Expose-Headers"))
	assert.Empty(t, rec.Header().Get("x-fapi-interaction-id"))
}

func TestInteractionIDSetsOnlyTheInteractionID(t *testing.T) {
	rec := run(t, cds.InteractionID(http.HandlerFunc(noContent)))

	assert.NotEmpty(t, rec.Header().Get("x-fapi-interaction-id"))
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}
