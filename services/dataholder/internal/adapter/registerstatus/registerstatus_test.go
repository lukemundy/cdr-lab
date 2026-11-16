package registerstatus_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registerstatus"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

const statusPath = "/cdr-register/v1/all/data-recipients/brands/software-products/status"

type request struct{ path, version, ifNoneMatch string }

func register(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (registerstatus.Client, *[]request) {
	t.Helper()

	var got []request

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, request{r.URL.Path, r.Header.Get("x-v"), r.Header.Get("If-None-Match")})
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	return registerstatus.NewClient(server.URL+"/", server.Client().Transport), &got
}

// Verifies: DH-0679
func TestTheStatusesComeFromGetSoftwareProductsStatusesWithTheIndustryAll(t *testing.T) {
	client, got := register(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("ETag", `"v1"`)
		_, _ = w.Write([]byte(`{"data": [
			{"softwareProductId": "software-1", "status": "ACTIVE"},
			{"softwareProductId": "software-2", "status": "INACTIVE"}
		], "links": {"self": "https://register.example"}, "meta": {}}`))
	})

	response, err := client.SoftwareProductStatuses(t.Context(), "")

	require.NoError(t, err)
	assert.Equal(t, app.StatusResponse{
		Statuses: map[string]string{"software-1": "ACTIVE", "software-2": "INACTIVE"}, ETag: `"v1"`,
	}, response)
	assert.Equal(t, []request{{statusPath, "3", ""}}, *got, "a first request has no If-None-Match")
}

// Verifies: DH-0685
func TestARequestWithTheETagOfTheLastResponseCanGetNotModified(t *testing.T) {
	client, got := register(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotModified) })

	response, err := client.SoftwareProductStatuses(t.Context(), `"v1"`)

	require.NoError(t, err)
	assert.Equal(t, app.StatusResponse{NotModified: true}, response)
	assert.Equal(t, `"v1"`, (*got)[0].ifNoneMatch)
}

func TestAResponseThatIsNotSuccessfulIsAnError(t *testing.T) {
	for name, handler := range map[string]func(w http.ResponseWriter, r *http.Request){
		"server error": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) },
		"redirect": func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://other.example/status", http.StatusFound)
		},
		"body that is not JSON": func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) },
	} {
		t.Run(name, func(t *testing.T) {
			client, _ := register(t, handler)

			_, err := client.SoftwareProductStatuses(t.Context(), "")

			require.Error(t, err)
		})
	}
}
