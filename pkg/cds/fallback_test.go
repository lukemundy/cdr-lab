package cds_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

func fallbackMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", noContent)
	mux.HandleFunc("GET /items/{id}", noContent)
	mux.HandleFunc("PUT /items/{id}", noContent)
	mux.HandleFunc("POST /orders", noContent)
	cds.RegisterFallback(mux, "The URL is not an API endpoint.")

	return mux
}

func sendFallback(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	fallbackMux().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, path, nil))

	return rec
}

// Verifies: CDSL-0080, CDSL-0081
func TestAMethodThatThePathDoesNotSupportGivesMethodNotAllowed(t *testing.T) {
	tests := map[string]struct {
		method, path, allow string
	}{
		"POST on a GET path":           {http.MethodPost, "/items", "GET, HEAD"},
		"DELETE on a GET and PUT path": {http.MethodDelete, "/items/a", "GET, HEAD, PUT"},
		"GET on a POST path":           {http.MethodGet, "/orders", "POST"},
		"OPTIONS on a GET path":        {http.MethodOptions, "/items", "GET, HEAD"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rec := sendFallback(t, tc.method, tc.path)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
			assert.Equal(t, tc.allow, rec.Header().Get("Allow"))
			assert.JSONEq(t, `{"errors": [{
				"code": "urn:au-cds:error:cds-all:GeneralError/Expected",
				"title": "Expected Error Encountered",
				"detail": "The endpoint does not support the method `+tc.method+`."
			}]}`, rec.Body.String())
		})
	}
}

// Verifies: CDSL-0082
func TestAPathWithNoRouteGivesResourceNotFound(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		for _, path := range []string{"/", "/unknown", "/items/a/b", "/items/"} {
			t.Run(method+" "+path, func(t *testing.T) {
				rec := sendFallback(t, method, path)

				assert.Equal(t, http.StatusNotFound, rec.Code)
				assert.Empty(t, rec.Header().Get("Allow"))
				assert.JSONEq(t, `{"errors": [{
					"code": "urn:au-cds:error:cds-all:Resource/NotFound",
					"title": "Resource Not Found",
					"detail": "The URL is not an API endpoint."
				}]}`, rec.Body.String())
			})
		}
	}
}

func TestARouteStillServesItsMethod(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		assert.Equal(t, http.StatusNoContent, sendFallback(t, method, "/items").Code)
	}
}
