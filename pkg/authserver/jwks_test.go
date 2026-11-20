package authserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

// Verifies: AUTH-0242
func TestTheJWKSClientDoesNotFollowARedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/other", http.StatusFound)
	}))
	t.Cleanup(server.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/jwks", nil)
	require.NoError(t, err)

	resp, err := authserver.NewJWKSClient(nil).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
}

// Verifies: AUTH-0243
func TestTheJWKSClientHasATimeoutOf5Seconds(t *testing.T) {
	assert.Equal(t, 5*time.Second, authserver.NewJWKSClient(nil).Timeout)
}
