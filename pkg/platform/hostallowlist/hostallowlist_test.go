package hostallowlist_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/hostallowlist"
)

func serve(t *testing.T, allowed []string, requestHost, method string) *httptest.ResponseRecorder {
	t.Helper()

	handler := hostallowlist.Middleware(allowed)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequestWithContext(t.Context(), method, "http://"+requestHost+"/any/path", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func TestARequestWithAHostThatIsNotAllowedIsRefused(t *testing.T) {
	allowed := []string{"server.test:8082", "127.0.0.1:8082", "proxy.test"}
	hosts := map[string]string{
		"DNS rebinding name":             "attacker.example:8082",
		"DNS rebinding name, no port":    "attacker.example",
		"different port":                 "server.test:8083",
		"no port":                        "server.test",
		"allowed IP, no port":            "127.0.0.1",
		"allowed name with no port, 443": "proxy.test:443",
	}

	for name, requestHost := range hosts {
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
			t.Run(name+" "+method, func(t *testing.T) {
				rec := serve(t, allowed, requestHost, method)

				require.Equal(t, http.StatusForbidden, rec.Code)
				assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

				var p struct {
					Status int    `json:"status"`
					Detail string `json:"detail"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p))
				assert.Equal(t, http.StatusForbidden, p.Status)
				assert.Equal(t, "The server refuses a request for a host name that it does not serve.", p.Detail)
			})
		}
	}
}

func TestARequestWithAnAllowedHostIsNotRefused(t *testing.T) {
	allowed := []string{"server.test:8082", "127.0.0.1:8082", "proxy.test"}

	for _, requestHost := range []string{
		"server.test:8082", "SERVER.Test:8082", "127.0.0.1:8082", "proxy.test", "Proxy.TEST",
	} {
		t.Run(requestHost, func(t *testing.T) {
			assert.Equal(t, http.StatusNoContent, serve(t, allowed, requestHost, http.MethodPost).Code)
		})
	}
}

func TestAnEmptyAllowlistRefusesEachRequest(t *testing.T) {
	assert.Equal(t, http.StatusForbidden, serve(t, nil, "server.test:8082", http.MethodGet).Code)
}

func TestParseGivesTheHostsInLowerCase(t *testing.T) {
	got, err := hostallowlist.Parse(" 127.0.0.1:8084, Localhost:8084 ,[::1]:8084, Proxy.Test, [::1]")

	require.NoError(t, err)
	assert.Equal(t, []string{"127.0.0.1:8084", "localhost:8084", "[::1]:8084", "proxy.test", "[::1]"}, got)
}

func TestParseRefusesAValueThatIsNotAListOfHosts(t *testing.T) {
	for _, value := range []string{
		"", " ", "127.0.0.1:8084,", "localhost:", ":8084", "localhost:99999",
		"http://localhost:8084", "http://localhost", "::1", "[]",
	} {
		t.Run(value, func(t *testing.T) {
			_, err := hostallowlist.Parse(value)

			assert.Error(t, err)
		})
	}
}
