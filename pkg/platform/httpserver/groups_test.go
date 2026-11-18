package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/pkg/platform/hostallowlist"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
)

// sharedListener has the probes, a group under /v1/ and a catch-all group, each of the last two with a Host allowlist.
func sharedListener() *http.ServeMux {
	mux := http.NewServeMux()
	httpserver.RegisterProbes(mux, ready)

	routes := httproute.New(mux)
	routes.Use(hostallowlist.Middleware([]string{"internal.test:8080"}))
	routes.Handle("/v1/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	return mux
}

func withCatchAll(mux *http.ServeMux) *http.ServeMux {
	routes := httproute.New(mux)
	routes.Use(hostallowlist.Middleware([]string{"internal.test:8080"}))
	routes.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.WriteHeader(http.StatusOK)
	}))

	return mux
}

func serveOn(t *testing.T, mux *http.ServeMux, method, target string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, target, nil))

	return rec
}

// Verifies: SRV-0009, SRV-0011, SRV-0012
func TestTheProbesDoNotGetThePolicyOfADifferentGroup(t *testing.T) {
	mux := withCatchAll(sharedListener())

	for _, path := range []string{"/livez", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			rec := serveOn(t, mux, http.MethodGet, "http://10.0.0.7:8080"+path)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "ok", rec.Body.String())
			assert.Empty(t, rec.Header().Get("Content-Security-Policy"))
		})
	}
}

// Verifies: SRV-0009
func TestAProbeGives405ForADifferentMethodAlsoWithACatchAllGroup(t *testing.T) {
	mux := withCatchAll(sharedListener())

	for _, path := range []string{"/livez", "/readyz"} {
		for _, method := range []string{http.MethodPost, http.MethodDelete} {
			t.Run(method+" "+path, func(t *testing.T) {
				rec := serveOn(t, mux, method, "http://10.0.0.7:8080"+path)

				assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
				assert.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))
				assert.Empty(t, rec.Header().Get("Content-Security-Policy"))
			})
		}
	}
}

func TestAProbeAcceptsHEAD(t *testing.T) {
	rec := serveOn(t, sharedListener(), http.MethodHead, "/livez")

	assert.Equal(t, http.StatusOK, rec.Code)
}

// Verifies: SRV-0009
func TestThePolicyOfAGroupAppliesToItsOwnPaths(t *testing.T) {
	mux := sharedListener()

	assert.Equal(t, http.StatusForbidden, serveOn(t, mux, http.MethodGet, "http://10.0.0.7:8080/v1/x").Code)
	assert.Equal(t, http.StatusTeapot, serveOn(t, mux, http.MethodGet, "http://internal.test:8080/v1/x").Code)
}

// Verifies: SRV-0010
func TestAPathOfNoGroupIsNotFound(t *testing.T) {
	mux := sharedListener()

	for _, path := range []string{"/", "/v2/x", "/livez/x", "/crl/x"} {
		t.Run(path, func(t *testing.T) {
			assert.Equal(t, http.StatusNotFound, serveOn(t, mux, http.MethodGet, "http://internal.test:8080"+path).Code)
		})
	}
}

func TestGetOnlyGives405ForADifferentMethod(t *testing.T) {
	handler := httpserver.GetOnly(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for method, want := range map[string]int{
		http.MethodGet:    http.StatusNoContent,
		http.MethodHead:   http.StatusNoContent,
		http.MethodPost:   http.StatusMethodNotAllowed,
		http.MethodPut:    http.StatusMethodNotAllowed,
		http.MethodDelete: http.StatusMethodNotAllowed,
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, "/x", nil))

		assert.Equal(t, want, rec.Code, method)
	}
}
