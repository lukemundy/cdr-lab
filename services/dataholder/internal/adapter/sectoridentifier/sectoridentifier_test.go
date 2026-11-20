package sectoridentifier_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/sectoridentifier"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

func serve(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/moved" {
			http.Redirect(w, r, "/sector.json", http.StatusFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	return srv
}

// Verifies: DH-2633
func TestTheRedirectURIsOfADocumentAreTheValuesOfItsArray(t *testing.T) {
	srv := serve(t, http.StatusOK, `["https://recipient.example/cb", "https://recipient.example/other"]`)

	uris, err := sectoridentifier.NewDocuments(srv.Client().Transport).RedirectURIs(t.Context(), srv.URL+"/sector.json")

	require.NoError(t, err)
	assert.Equal(t, []string{"https://recipient.example/cb", "https://recipient.example/other"}, uris)
}

// Verifies: DH-2633, DH-2635
func TestADocumentThatTheDataHolderCannotGetOrReadIsNotAvailable(t *testing.T) {
	tests := map[string]struct {
		status int
		body   string
		path   string
	}{
		"a status that is not 200": {http.StatusNotFound, `["https://recipient.example/cb"]`, "/sector.json"},
		"a redirect":               {http.StatusOK, `["https://recipient.example/cb"]`, "/moved"},
		"a body that is not JSON":  {http.StatusOK, `<html></html>`, "/sector.json"},
		"a JSON object": {
			http.StatusOK,
			`{"redirect_uris": ["https://recipient.example/cb"]}`,
			"/sector.json",
		},
		"an array that is not string": {http.StatusOK, `["https://recipient.example/cb", 1]`, "/sector.json"},
		"more than 1 MiB": {
			http.StatusOK, `["https://recipient.example/cb", "` + strings.Repeat("a", 1<<20) + `"]`, "/sector.json",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			srv := serve(t, tc.status, tc.body)

			_, err := sectoridentifier.NewDocuments(srv.Client().Transport).RedirectURIs(t.Context(), srv.URL+tc.path)

			assert.ErrorIs(t, err, app.ErrSectorDocumentNotAvailable)
		})
	}
}

// Verifies: DH-2632
func TestTheDataHolderDoesNotTrustACAThatIsNotInItsTransport(t *testing.T) {
	srv := serve(t, http.StatusOK, `["https://recipient.example/cb"]`)

	_, err := sectoridentifier.NewDocuments(http.DefaultTransport).RedirectURIs(t.Context(), srv.URL+"/sector.json")

	assert.ErrorIs(t, err, app.ErrSectorDocumentNotAvailable)
}
