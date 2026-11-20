package operatorhttp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
)

const (
	token       = "test-operator-token"
	allowedHost = "operator.test:8080"
)

var cfg = operatorhttp.Config{Token: token, AllowedHosts: []string{allowedHost}}

type problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func requireProblem(t *testing.T, rec *httptest.ResponseRecorder, status int) problem {
	t.Helper()
	require.Equal(t, status, rec.Code, rec.Body.String())
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var p problem
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p))

	return p
}

func serve(t *testing.T, header http.Header) (rec *httptest.ResponseRecorder, called bool) {
	t.Helper()

	return serveHost(t, allowedHost, header)
}

func serveHost(t *testing.T, requestHost string, header http.Header) (rec *httptest.ResponseRecorder, called bool) {
	t.Helper()

	handler := operatorhttp.Guard(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "http://"+requestHost+"/", nil)
	req.Header = header

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec, called
}

// Verifies: OPI-0001
func TestARequestWithTheOperatorTokenReachesTheHandler(t *testing.T) {
	for _, scheme := range []string{"Bearer", "bearer"} {
		t.Run(scheme, func(t *testing.T) {
			rec, called := serve(t, http.Header{"Authorization": {scheme + " " + token}})

			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.True(t, called)
		})
	}
}

// Verifies: OPI-0001, OPI-0002, OPI-0003, OPI-0010
func TestARequestWithoutTheOperatorTokenIsUnauthorised(t *testing.T) {
	const (
		noErrorChallenge, invalidChallenge = `Bearer`, `Bearer error="invalid_token"`
		missing                            = "The request must have the operator token as a bearer token."
		invalid                            = "The bearer token is not the operator token."
	)

	tests := map[string]struct {
		header    http.Header
		challenge string
		detail    string
	}{
		"no authorization": {header: http.Header{}, challenge: noErrorChallenge, detail: missing},
		"not a bearer token": {
			header:    http.Header{"Authorization": {"Basic " + token}},
			challenge: noErrorChallenge,
			detail:    missing,
		},
		"no scheme": {header: http.Header{"Authorization": {token}}, challenge: noErrorChallenge, detail: missing},
		"wrong token": {
			header:    http.Header{"Authorization": {"Bearer not-the-token"}},
			challenge: invalidChallenge,
			detail:    invalid,
		},
		"token as a prefix": {
			header:    http.Header{"Authorization": {"Bearer " + token + "x"}},
			challenge: invalidChallenge,
			detail:    invalid,
		},
		"empty token": {
			header:    http.Header{"Authorization": {"Bearer "}},
			challenge: invalidChallenge,
			detail:    invalid,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec, called := serve(t, tt.header)

			p := requireProblem(t, rec, http.StatusUnauthorized)
			assert.Equal(t, tt.detail, p.Detail)
			assert.Equal(t, tt.challenge, rec.Header().Get("WWW-Authenticate"))
			assert.False(t, called)
		})
	}
}

// Verifies: OPI-0005
func TestARequestForAHostThatIsNotAllowedIsRefusedBeforeTheTokenCheck(t *testing.T) {
	for name, header := range map[string]http.Header{
		"operator token": {"Authorization": {"Bearer " + token}},
		"no token":       {},
	} {
		t.Run(name, func(t *testing.T) {
			rec, called := serveHost(t, "attacker.example:8082", header)

			requireProblem(t, rec, http.StatusForbidden)
			assert.Empty(t, rec.Header().Get("WWW-Authenticate"))
			assert.False(t, called)
		})
	}
}

func setEnv(t *testing.T, tokenValue, hostsValue string) {
	t.Helper()
	t.Setenv(operatorhttp.TokenEnv, tokenValue)
	t.Setenv(operatorhttp.AllowedHostsEnv, hostsValue)
}

// Verifies: OPI-0004, OPI-0006
func TestLoadConfigGetsTheTokenAndTheAllowedHosts(t *testing.T) {
	setEnv(t, token, "127.0.0.1:8080, Register:8080")

	got, err := operatorhttp.LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, operatorhttp.Config{Token: token, AllowedHosts: []string{"127.0.0.1:8080", "register:8080"}}, got)
}

// Verifies: OPI-0004, OPI-0006
func TestLoadConfigNeedsEachSetting(t *testing.T) {
	tests := map[string]struct {
		token, hosts, wantEnv string
	}{
		"no token":            {hosts: allowedHost, wantEnv: operatorhttp.TokenEnv},
		"no allowed hosts":    {token: token, wantEnv: operatorhttp.AllowedHostsEnv},
		"a port with no host": {token: token, hosts: ":8082", wantEnv: operatorhttp.AllowedHostsEnv},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			setEnv(t, tt.token, tt.hosts)

			_, err := operatorhttp.LoadConfig()

			assert.ErrorContains(t, err, tt.wantEnv)
		})
	}
}
