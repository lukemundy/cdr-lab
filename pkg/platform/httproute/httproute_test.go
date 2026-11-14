package httproute_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
)

func record(name string, calls *[]string) httproute.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*calls = append(*calls, name)

			next.ServeHTTP(w, r)
		})
	}
}

func handler(calls *[]string) http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) { *calls = append(*calls, "handler") })
}

func get(t *testing.T, mux *http.ServeMux, path string) {
	t.Helper()

	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
}

func TestARouteRunsTheUseMiddlewareThenTheWithMiddlewareThenTheHandler(t *testing.T) {
	var calls []string

	mux := http.NewServeMux()
	routes := httproute.New(mux)
	routes.Use(record("use-1", &calls), record("use-2", &calls))
	routes.With(record("with-1", &calls), record("with-2", &calls)).Handle("GET /a", handler(&calls))

	get(t, mux, "/a")

	assert.Equal(t, "use-1,use-2,with-1,with-2,handler", strings.Join(calls, ","))
}

func TestARouteWithNoMiddlewareRunsTheHandler(t *testing.T) {
	var calls []string

	mux := http.NewServeMux()
	httproute.New(mux).Handle("GET /a", handler(&calls))

	get(t, mux, "/a")

	assert.Equal(t, []string{"handler"}, calls)
}

func TestWithDoesNotChangeTheOtherRoutes(t *testing.T) {
	var calls []string

	mux := http.NewServeMux()
	routes := httproute.New(mux)
	routes.Use(record("use", &calls))
	routes.With(record("a", &calls)).Handle("GET /a", handler(&calls))
	routes.With(record("b", &calls)).Handle("GET /b", handler(&calls))
	routes.Handle("GET /c", handler(&calls))

	get(t, mux, "/a")
	get(t, mux, "/b")
	get(t, mux, "/c")

	assert.Equal(t, "use,a,handler,use,b,handler,use,handler", strings.Join(calls, ","))
}

func TestUseAfterHandlePanics(t *testing.T) {
	routes := httproute.New(http.NewServeMux())
	routes.Handle("GET /a", http.NotFoundHandler())

	assert.Panics(t, func() { routes.Use() })
}

func TestUseAfterWithPanics(t *testing.T) {
	routes := httproute.New(http.NewServeMux())
	routes.With()

	assert.Panics(t, func() { routes.Use() })
}
