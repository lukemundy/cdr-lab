package consolehttp_test

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
)

const (
	host      = "console.test:8084"
	indexHTML = `<!doctype html><title>Console</title><script type="module" src="/assets/index-a1b2c3.js"></script>`
)

var ui = fstest.MapFS{
	"index.html":             {Data: []byte(indexHTML)},
	"favicon.svg":            {Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`)},
	"assets/index-a1b2c3.js": {Data: []byte(`console.log("console")`)},
}

func newHandler(t *testing.T) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	require.NoError(
		t,
		consolehttp.Console.Register(mux, ui, []string{host, "127.0.0.1:8084"}, func(api *httproute.Routes) {
			api.Handle("POST /api/echo", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
		}),
	)

	return mux
}

func serve(t *testing.T, method, target string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	return serveHost(t, host, method, target, header)
}

func serveHost(t *testing.T, requestHost, method, target string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, "http://"+requestHost+target, nil)
	maps.Copy(req.Header, header)

	rec := httptest.NewRecorder()
	newHandler(t).ServeHTTP(rec, req)

	return rec
}

func requireProblem(t *testing.T, rec *httptest.ResponseRecorder, status int) {
	t.Helper()
	require.Equal(t, status, rec.Code, rec.Body.String())
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var p struct {
		Type   string `json:"type"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p))
	assert.Equal(t, "about:blank", p.Type)
	assert.Equal(t, status, p.Status)
	assert.NotEmpty(t, p.Detail)
}

func requireHTMLPage(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
}

func TestTheRootGivesTheHTMLPage(t *testing.T) {
	rec := serve(t, http.MethodGet, "/", nil)

	requireHTMLPage(t, rec)
	assert.Equal(t, indexHTML, rec.Body.String())
}

// Verifies: CNSL-0006, CNSL-0007
func TestEachResponseHasTheSecurityHeadersAndNoCORSHeaders(t *testing.T) {
	tests := map[string]struct {
		method, target string
		header         http.Header
	}{
		"page":            {method: http.MethodGet, target: "/"},
		"hashed file":     {method: http.MethodGet, target: "/assets/index-a1b2c3.js"},
		"route of the UI": {method: http.MethodGet, target: "/data-recipients"},
		"API": {
			method: http.MethodPost,
			target: "/api/echo",
			header: http.Header{"Sec-Fetch-Site": {"same-origin"}},
		},
		"unknown API path": {method: http.MethodGet, target: "/api/unknown"},
		"cross-origin POST": {
			method: http.MethodPost,
			target: "/api/echo",
			header: http.Header{"Sec-Fetch-Site": {"cross-site"}},
		},
		"POST to the UI": {
			method: http.MethodPost,
			target: "/",
			header: http.Header{"Sec-Fetch-Site": {"same-origin"}},
		},
		"preflight": {
			method: http.MethodOptions,
			target: "/api/echo",
			header: http.Header{
				"Origin":                         {"http://other.test"},
				"Access-Control-Request-Method":  {http.MethodPost},
				"Access-Control-Request-Headers": {"Content-Type"},
			},
		},
		"OPTIONS on the UI": {
			method: http.MethodOptions,
			target: "/",
			header: http.Header{"Origin": {"http://other.test"}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, tt.method, tt.target, tt.header)

			assert.Equal(t,
				"default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; "+
					"connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'; "+
					"frame-ancestors 'none'; require-trusted-types-for 'script'",
				rec.Header().Get("Content-Security-Policy"),
			)
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
			assert.Equal(t, "same-origin", rec.Header().Get("Cross-Origin-Opener-Policy"))

			for name := range rec.Header() {
				assert.False(t, strings.HasPrefix(name, "Access-Control-"), "header %s", name)
			}
		})
	}
}

// Verifies: CNSL-0005
func TestAChangeFromADifferentOriginIsRefused(t *testing.T) {
	tests := map[string]http.Header{
		"cross-site":                 {"Sec-Fetch-Site": {"cross-site"}},
		"same-site":                  {"Sec-Fetch-Site": {"same-site"}},
		"origin of a different port": {"Origin": {"http://console.test:8082"}},
		"origin of a different host": {"Origin": {"http://other.test:8084"}},
	}
	for name, header := range tests {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
			t.Run(name+" "+method, func(t *testing.T) {
				requireProblem(t, serve(t, method, "/api/echo", header), http.StatusForbidden)
			})
		}
	}
}

func TestAChangeFromTheSameOriginIsNotRefused(t *testing.T) {
	tests := map[string]http.Header{
		"same-origin":        {"Sec-Fetch-Site": {"same-origin"}},
		"user action":        {"Sec-Fetch-Site": {"none"}},
		"origin of the host": {"Origin": {"http://" + host}},
		"no fetch metadata":  {},
	}
	for name, header := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, http.MethodPost, "/api/echo", header)

			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
	}
}

// Verifies: CNSL-0008, CNSL-0009, CNSL-0010
func TestTheCacheRulesDependOnTheFile(t *testing.T) {
	tests := map[string]struct {
		method, target, cacheControl string
	}{
		"page":            {method: http.MethodGet, target: "/", cacheControl: "no-cache"},
		"route of the UI": {method: http.MethodGet, target: "/data-recipients/1", cacheControl: "no-cache"},
		"hashed file": {
			method:       http.MethodGet,
			target:       "/assets/index-a1b2c3.js",
			cacheControl: "public, max-age=31536000, immutable",
		},
		"file with no hash": {method: http.MethodGet, target: "/favicon.svg", cacheControl: "no-cache"},
		"API":               {method: http.MethodPost, target: "/api/echo", cacheControl: "no-store"},
		"unknown API path":  {method: http.MethodGet, target: "/api/unknown", cacheControl: "no-store"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, tt.method, tt.target, nil)

			assert.Equal(t, tt.cacheControl, rec.Header().Get("Cache-Control"))
		})
	}
}

func TestAFileOfTheUIHasItsContentType(t *testing.T) {
	rec := serve(t, http.MethodGet, "/assets/index-a1b2c3.js", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/javascript; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, `console.log("console")`, rec.Body.String())
}

// Verifies: CNSL-0013
func TestAnUnknownPathOutsideTheAPIGivesTheHTMLPage(t *testing.T) {
	for _, target := range []string{"/data-recipients", "/data-recipients/1/history", "/assets/", "/index.html/x"} {
		t.Run(target, func(t *testing.T) {
			rec := serve(t, http.MethodGet, target, nil)

			requireHTMLPage(t, rec)
			assert.Equal(t, indexHTML, rec.Body.String())
		})
	}
}

// Verifies: CNSL-0013
func TestTheCatchAllDoesNotCatchThePathsOfTheOtherGroupsOnTheListener(t *testing.T) {
	mux := http.NewServeMux()
	httpserver.RegisterProbes(mux, func(context.Context) error { return nil })
	mux.Handle("/v1/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	require.NoError(t, consolehttp.Console.Register(mux, ui, []string{host}, func(*httproute.Routes) {}))

	for _, tt := range []struct {
		method, target string
		want           int
	}{
		{http.MethodGet, "/readyz", http.StatusOK},
		{http.MethodPost, "/livez", http.StatusMethodNotAllowed},
		{http.MethodGet, "/v1/unknown", http.StatusTeapot},
		{http.MethodPost, "/v1/", http.StatusTeapot},
	} {
		t.Run(tt.method+" "+tt.target, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), tt.method, "http://10.0.0.7:8080"+tt.target, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.want, rec.Code)
			assert.Empty(t, rec.Header().Get("Content-Security-Policy"))
		})
	}
}

// Verifies: CNSL-0028
func TestAHEADRequestGetsTheHeadersOfTheHTMLPage(t *testing.T) {
	rec := serve(t, http.MethodHead, "/data-recipients", nil)

	requireHTMLPage(t, rec)
	assert.Empty(t, rec.Body.String())
}

// Verifies: CNSL-0012
func TestAnUnknownAPIPathIsNotFound(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			requireProblem(t, serve(t, method, "/api/unknown", nil), http.StatusNotFound)
		})
	}

	t.Run("the API root", func(t *testing.T) {
		requireProblem(t, serve(t, http.MethodGet, "/api/", nil), http.StatusNotFound)
	})
}

// Verifies: CNSL-0027
func TestAChangeToAPathOfTheUIIsNotAllowed(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodOptions} {
		t.Run(method, func(t *testing.T) {
			rec := serve(t, method, "/data-recipients", nil)

			requireProblem(t, rec, http.StatusMethodNotAllowed)
			assert.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))
		})
	}
}

func TestAUIWithNoHTMLPageIsAnError(t *testing.T) {
	err := consolehttp.Console.Register(
		http.NewServeMux(), fstest.MapFS{".gitkeep": {}}, []string{host}, func(*httproute.Routes) {},
	)

	assert.ErrorContains(t, err, "index.html")
}

func TestAConsoleWithNoAllowedHostIsAnError(t *testing.T) {
	err := consolehttp.Console.Register(http.NewServeMux(), ui, nil, func(*httproute.Routes) {})

	assert.ErrorContains(t, err, consolehttp.Console.AllowedHostsEnv)
}

// Verifies: CNSL-0004
func TestARequestWithAHostThatIsNotAllowedIsRefused(t *testing.T) {
	hosts := map[string]string{
		"DNS rebinding name":  "attacker.example:8084",
		"different port":      "console.test:8085",
		"no port":             "console.test",
		"allowed IP, no port": "127.0.0.1",
	}
	targets := map[string]string{"page": "/", "file": "/assets/index-a1b2c3.js", "API": "/api/echo"}

	for hostName, requestHost := range hosts {
		for targetName, target := range targets {
			for _, method := range []string{http.MethodGet, http.MethodPost} {
				t.Run(hostName+" "+targetName+" "+method, func(t *testing.T) {
					// Same-origin fetch metadata: DNS rebinding makes the page and the console one origin.
					rec := serveHost(t, requestHost, method, target, http.Header{
						"Sec-Fetch-Site": {"same-origin"},
						"Origin":         {"http://" + requestHost},
					})

					requireProblem(t, rec, http.StatusForbidden)
					assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
				})
			}
		}
	}
}

func TestARequestWithAnAllowedHostIsNotRefused(t *testing.T) {
	for _, requestHost := range []string{host, "CONSOLE.Test:8084", "127.0.0.1:8084"} {
		t.Run(requestHost, func(t *testing.T) {
			requireHTMLPage(t, serveHost(t, requestHost, http.MethodGet, "/", nil))

			rec := serveHost(
				t,
				requestHost,
				http.MethodPost,
				"/api/echo",
				http.Header{"Sec-Fetch-Site": {"same-origin"}},
			)
			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
	}
}

// Verifies: CNSL-0002
func TestTheConsoleIsOffUnlessTheSettingTurnsItOn(t *testing.T) {
	tests := []struct {
		value   string
		want    bool
		wantErr bool
	}{
		{value: "", want: false},
		{value: "false", want: false},
		{value: "true", want: true},
		{value: "yes", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Setenv(consolehttp.Console.EnabledEnv, tt.value)
			t.Setenv(consolehttp.Console.AllowedHostsEnv, host)

			got, err := consolehttp.Console.LoadConfig()

			if tt.wantErr {
				assert.ErrorContains(t, err, consolehttp.Console.EnabledEnv)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Enabled)
		})
	}
}

func TestAnEnabledConsoleGetsTheAllowedHostsOfTheSetting(t *testing.T) {
	t.Setenv(consolehttp.Console.EnabledEnv, "true")
	t.Setenv(consolehttp.Console.AllowedHostsEnv, " 127.0.0.1:8084, Localhost:8084 ,[::1]:8084")

	got, err := consolehttp.Console.LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, consolehttp.Config{
		Enabled:      true,
		AllowedHosts: []string{"127.0.0.1:8084", "localhost:8084", "[::1]:8084"},
	}, got)
}

// Verifies: CNSL-0003
func TestAnEnabledConsoleNeedsValidAllowedHosts(t *testing.T) {
	for _, value := range []string{"", " ", "127.0.0.1:8084,", "localhost:", "http://localhost:8084"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(consolehttp.Console.EnabledEnv, "true")
			t.Setenv(consolehttp.Console.AllowedHostsEnv, value)

			_, err := consolehttp.Console.LoadConfig()

			assert.ErrorContains(t, err, consolehttp.Console.AllowedHostsEnv)
		})
	}
}

func TestADisabledConsoleDoesNotNeedAllowedHosts(t *testing.T) {
	t.Setenv(consolehttp.Console.EnabledEnv, "false")
	t.Setenv(consolehttp.Console.AllowedHostsEnv, "")

	got, err := consolehttp.Console.LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, consolehttp.Config{}, got)
}

var site = consolehttp.Settings{
	Name: "bank site", EnabledEnv: "SITE_ENABLED", AllowedHostsEnv: "SITE_ALLOWED_HOSTS", ExampleHost: "127.0.0.1:8185",
}

func TestTheSettingsOfAUIGiveTheNamesOfItsSettings(t *testing.T) {
	t.Setenv(site.EnabledEnv, "true")
	t.Setenv(site.AllowedHostsEnv, "127.0.0.1:8185")

	got, err := site.LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, consolehttp.Config{Enabled: true, AllowedHosts: []string{"127.0.0.1:8185"}}, got)

	t.Setenv(site.AllowedHostsEnv, "")

	_, err = site.LoadConfig()

	require.ErrorContains(t, err, site.AllowedHostsEnv)
	assert.ErrorContains(t, err, site.ExampleHost)
}

func TestAUIWithNoAllowedHostNamesTheSettingOfTheUI(t *testing.T) {
	err := site.Register(http.NewServeMux(), ui, nil, func(*httproute.Routes) {})

	assert.ErrorContains(t, err, site.AllowedHostsEnv)
}

func TestTheRefusalOfAChangeFromADifferentOriginNamesTheUI(t *testing.T) {
	mux := http.NewServeMux()
	require.NoError(t, site.Register(mux, ui, []string{host}, func(*httproute.Routes) {}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "http://"+host+"/api/echo", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "The bank site refuses a change from a different origin.")
}
