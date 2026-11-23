package operator_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/operator"
)

const token = "operator-token"

type seen struct {
	method, path, authorization, contentType, accept string
	body                                             string
}

func server(t *testing.T, status int, response string) (*httptest.Server, *seen) {
	t.Helper()

	got := &seen{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		*got = seen{
			method: r.Method, path: r.URL.Path, authorization: r.Header.Get("Authorization"),
			contentType: r.Header.Get("Content-Type"), accept: r.Header.Get("Accept"), body: string(body),
		}

		if status != http.StatusOK {
			w.Header().Set("Content-Type", "application/problem+json")
		}

		w.WriteHeader(status)
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(srv.Close)

	return srv, got
}

func TestDoSendsTheRequestWithTheTokenAndGivesTheBody(t *testing.T) {
	srv, got := server(t, http.StatusOK, "the body")

	body, err := operator.New("the Register", srv.URL+"/", token, srv.Client()).Do(t.Context(), operator.Request{
		Method: http.MethodPost, Path: "/v1/things", ContentType: "text/plain", Accept: "text/plain",
		Body: []byte("a thing"),
	})

	require.NoError(t, err)
	assert.Equal(t, "the body", string(body))
	assert.Equal(t, seen{
		method: http.MethodPost, path: "/v1/things", authorization: "Bearer " + token,
		contentType: "text/plain", accept: "text/plain", body: "a thing",
	}, *got)
}

func TestPostJSONSendsJSONAndDecodesTheResponse(t *testing.T) {
	srv, got := server(t, http.StatusOK, `{"count": 2, "other": "x"}`)

	var into struct {
		Count int `json:"count"`
	}

	err := operator.New("the Register", srv.URL, token, srv.Client()).
		PostJSON(t.Context(), "/v1/loads", []byte(`{"a": 1}`), &into)

	require.NoError(t, err)
	assert.Equal(t, 2, into.Count)
	assert.Equal(t, "application/json", got.contentType)
	assert.Equal(t, "application/json", got.accept)
	assert.JSONEq(t, `{"a": 1}`, got.body)
}

func TestPostJSONRefusesAResponseThatIsNotJSON(t *testing.T) {
	srv, _ := server(t, http.StatusOK, "<html>not ready</html>")

	var into struct{}

	err := operator.New("the Register", srv.URL, token, srv.Client()).PostJSON(t.Context(), "/v1/loads", nil, &into)

	require.Error(t, err)
}

// Verifies: CLI-0038, CLI-0184
func TestARefusalGivesTheNameTheStatusAndTheDetailOfTheProblem(t *testing.T) {
	srv, _ := server(t, http.StatusBadRequest,
		`{"type":"about:blank","title":"Bad Request","status":400,"detail":"O must be Wattle Bank."}`)

	_, err := operator.New("the data holder", srv.URL, token, srv.Client()).
		Do(t.Context(), operator.Request{Method: http.MethodGet, Path: "/"})

	require.EqualError(t, err, "the data holder gave 400 Bad Request: O must be Wattle Bank.")
}

func TestARefusalWithNoProblemGivesTheStatus(t *testing.T) {
	srv, _ := server(t, http.StatusServiceUnavailable, "<html>not ready</html>")

	_, err := operator.New("the Register", srv.URL, token, srv.Client()).
		Do(t.Context(), operator.Request{Method: http.MethodGet, Path: "/"})

	require.EqualError(t, err, "the Register gave 503 Service Unavailable")
}

// Verifies: CLI-0060
func TestAnErrorDoesNotGiveTheToken(t *testing.T) {
	srv, _ := server(t, http.StatusUnauthorized, `{"status":401,"detail":"The request has no valid operator token."}`)
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()

	for _, base := range []string{srv.URL, closed.URL} {
		_, err := operator.New("the Register", base, token, srv.Client()).
			Do(t.Context(), operator.Request{Method: http.MethodGet, Path: "/"})

		require.Error(t, err)
		assert.NotContains(t, err.Error(), token)
	}
}
