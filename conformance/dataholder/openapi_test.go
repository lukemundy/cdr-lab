package dataholder

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThePinnedCommonSpecificationValidatesAResponseBelowTheBasePath(t *testing.T) {
	spec, err := LoadCommonSpec()
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-v", "1")
		_, _ = io.WriteString(w, `{"data":{"outages":[]},"links":{"self":"https://bank.example/x"},"meta":{}}`)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/cds-au/v1/discovery/outages", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)

	t.Cleanup(func() { _ = resp.Body.Close() })

	require.NoError(t, spec.ValidateResponse(resp))
}

func TestThePinnedBankingSpecificationRefusesAProductListWithAFieldThatIsNotValid(t *testing.T) {
	spec, err := LoadBankingSpec()
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-v", "5")
		_, _ = io.WriteString(w, `{
			"data": {"products": [{"productId": "saver", "productCategory": "SAVINGS"}]},
			"links": {"self": "https://bank.example/x"},
			"meta": {"totalRecords": 1, "totalPages": 1}
		}`)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/cds-au/v1/banking/products", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)

	t.Cleanup(func() { _ = resp.Body.Close() })

	err = spec.ValidateResponse(resp)

	require.ErrorContains(t, err, "productCategory")
	require.ErrorContains(t, err, "lastUpdated")
}

func TestThePinnedDCRSpecificationRefusesARegistrationWithAFieldThatIsNotValid(t *testing.T) {
	spec, err := LoadDCRSpec()
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"client_id": "client-1", "grant_types": ["implicit"]}`)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL+"/register", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)

	t.Cleanup(func() { _ = resp.Body.Close() })

	err = spec.ValidateResponse(resp)

	require.ErrorContains(t, err, "grant_types")
	require.ErrorContains(t, err, "client_name")
}
