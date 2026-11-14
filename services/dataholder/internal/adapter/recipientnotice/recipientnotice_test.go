package recipientnotice_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/recipientnotice"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

const (
	brandID       = "d7385786-d8b1-4d4b-bb80-b21616df1a4d"
	arrangementID = "8f4d7a52-3c1e-4b6a-9d2f-1e0c5b7a9d31"
	jwtLifetime   = 5 * time.Minute
)

var sentAt = time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)

type received struct {
	method, path, contentType, authorization string
	form                                     map[string][]string
}

func newSender(t *testing.T, client *http.Client) (recipientnotice.Sender, authserver.SigningKey) {
	t.Helper()

	key, err := authserver.NewSigningKey()
	require.NoError(t, err)

	return recipientnotice.Sender{
		BrandID: brandID,
		Keys:    func(context.Context) ([]authserver.SigningKey, error) { return []authserver.SigningKey{key}, nil },
		Client:  client,
		NewID:   uuid.NewString,
		Now:     func() time.Time { return sentAt },
	}, key
}

func recipient(t *testing.T, status int, got *received) *httptest.Server {
	t.Helper()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, r.ParseForm())
		*got = received{
			method: r.Method, path: r.URL.Path, contentType: r.Header.Get("Content-Type"),
			authorization: r.Header.Get("Authorization"), form: r.PostForm,
		}

		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)

	return server
}

func claimsOf(t *testing.T, token string, key authserver.SigningKey) map[string]any {
	t.Helper()

	parsed, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(t, err)
	require.Len(t, parsed.Headers, 1)
	assert.Equal(t, key.ID, parsed.Headers[0].KeyID)

	var claims map[string]any
	require.NoError(t, parsed.Claims(&key.Key.PublicKey, &claims))

	return claims
}

func TestANoticeIsAPostWithTwoSignedJWTsAndOnlyTheArrangementJWTInTheBody(t *testing.T) {
	var got received

	server := recipient(t, http.StatusNoContent, &got)
	sender, key := newSender(t, server.Client())
	endpoint := server.URL + "/cdr/arrangements/revoke"

	status, err := sender.Send(
		t.Context(),
		app.Notice{RecipientBaseURI: server.URL + "/cdr", ArrangementID: arrangementID},
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, http.MethodPost, got.method)
	assert.Equal(t, "/cdr/arrangements/revoke", got.path)
	assert.Equal(t, "application/x-www-form-urlencoded", got.contentType)
	require.Equal(t, []string{"cdr_arrangement_jwt"}, keys(got.form))

	bearer, found := strings.CutPrefix(got.authorization, "Bearer ")
	require.True(t, found, got.authorization)
	authClaims := claimsOf(t, bearer, key)
	arrangementClaims := claimsOf(t, got.form["cdr_arrangement_jwt"][0], key)

	jwts := map[string]map[string]any{"bearer": authClaims, "cdr_arrangement_jwt": arrangementClaims}
	for name, claims := range jwts {
		assert.Equal(t, brandID, claims["iss"], name)
		assert.Equal(t, brandID, claims["sub"], name)
		assert.Equal(t, endpoint, claims["aud"], name)
		assert.InDelta(t, sentAt.Unix(), claims["iat"], 0, name)
		assert.InDelta(t, sentAt.Add(jwtLifetime).Unix(), claims["exp"], 0, name)
		require.IsType(t, "", claims["jti"], name)
		require.NoError(t, uuid.Validate(claims["jti"].(string)), name)
	}

	assert.NotEqual(t, authClaims["jti"], arrangementClaims["jti"])
	assert.Equal(t, arrangementID, arrangementClaims["cdr_arrangement_id"])
	assert.NotContains(t, authClaims, "cdr_arrangement_id")
}

func keys(form map[string][]string) []string {
	out := make([]string, 0, len(form))
	for k := range form {
		out = append(out, k)
	}

	return out
}

func TestARedirectIsNotFollowed(t *testing.T) {
	server := httptest.NewTLSServer(http.RedirectHandler("https://other.example/revoke", http.StatusFound))
	t.Cleanup(server.Close)
	client := server.Client()
	client.CheckRedirect = recipientnotice.NewClient(nil).CheckRedirect
	sender, _ := newSender(t, client)

	status, err := sender.Send(t.Context(), app.Notice{RecipientBaseURI: server.URL, ArrangementID: arrangementID})

	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, status)
}

func TestASendWithNoResponseGivesItsClass(t *testing.T) {
	release := make(chan struct{})
	slow := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
	t.Cleanup(slow.Close)
	// Close waits for the handler, so the handler must end first.
	t.Cleanup(func() { close(release) })

	slowClient := slow.Client()
	slowClient.Timeout = 50 * time.Millisecond

	untrusted := httptest.NewTLSServer(http.NotFoundHandler())
	t.Cleanup(untrusted.Close)

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	closedAddress := listener.Addr().String()
	require.NoError(t, listener.Close())

	tests := map[string]struct {
		client  *http.Client
		baseURI string
		want    error
	}{
		"no response in the time limit": {client: slowClient, baseURI: slow.URL, want: app.ErrNoticeTimeout},
		"certificate that is not trusted": {
			client: &http.Client{}, baseURI: untrusted.URL, want: app.ErrNoticeTLS,
		},
		"no connection": {client: &http.Client{}, baseURI: "https://" + closedAddress, want: app.ErrNoticeNoConnection},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sender, _ := newSender(t, tc.client)

			_, err := sender.Send(t.Context(), app.Notice{RecipientBaseURI: tc.baseURI, ArrangementID: arrangementID})

			require.ErrorIs(t, err, tc.want)
		})
	}
}
