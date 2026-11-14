package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

type acceptedDisclosures struct{}

func (acceptedDisclosures) RecordDisclosure(context.Context, string, string) error { return nil }

func TestAHeldResponseHasExactlyTheHeadersOfTheHandler(t *testing.T) {
	handler := recordDisclosure(acceptedDisclosures{}, func(authserver.ConsentToken) string { return scopeProfile })(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Del("X-Outer")
			w.Header().Set("X-Inner", "set")
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/userinfo", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxKeyConsentToken{}, authserver.ConsentToken{
		ArrangementID: "arrangement-1",
	}))
	rec := httptest.NewRecorder()
	rec.Header().Set("X-Outer", "from an outer middleware")

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, http.Header{"X-Inner": {"set"}}, rec.Header())
}
