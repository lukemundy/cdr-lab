package httpserver_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
)

func probe(t *testing.T, ready func(context.Context) error, path string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	httpserver.RegisterProbes(mux, ready)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))

	return rec
}

func notReady(context.Context) error { return errors.New("database unavailable at 10.0.0.1") }

func ready(context.Context) error { return nil }

// Verifies: SRV-0001
func TestLivenessDoesNotDependOnReadiness(t *testing.T) {
	rec := probe(t, notReady, "/livez")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

// Verifies: SRV-0002
func TestReadinessIsOKWhenTheCheckPasses(t *testing.T) {
	rec := probe(t, ready, "/readyz")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

// Verifies: SRV-0002
func TestReadinessIsUnavailableWhenTheCheckFails(t *testing.T) {
	rec := probe(t, notReady, "/readyz")

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "not ready", rec.Body.String(), "the body must not show the error")
}
