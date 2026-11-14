//go:build conformance

package dataholder

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

type endpointDetail struct {
	PublicBaseURI  string `json:"publicBaseUri"`
	ProductBaseURI string `json:"productBaseUri"`
}

// registeredEndpoints gives the base URIs that the Register has for the brand, as a client finds them.
func registeredEndpoints(t *testing.T) endpointDetail {
	t.Helper()

	url := labtest.Env(t, "REGISTER_PUBLIC_URL") + "/cdr-register/v1/banking/data-holders/brands/summary"
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("x-v", "2")

	resp, err := labtest.Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var summary struct {
		Data []struct {
			DataHolderBrandID string `json:"dataHolderBrandId"`
			endpointDetail
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&summary))

	brandID := labtest.Env(t, "DATAHOLDER_BRAND_ID")
	for _, brand := range summary.Data {
		if brand.DataHolderBrandID == brandID {
			return brand.endpointDetail
		}
	}

	t.Fatalf("the Register has no banking brand %s", brandID)

	return endpointDetail{}
}

func publicBaseURI(t *testing.T) string {
	t.Helper()

	return registeredEndpoints(t).PublicBaseURI
}

type response struct {
	*http.Response
	body []byte
}

func get(t *testing.T, url string, header http.Header) response {
	t.Helper()

	return send(t, http.MethodGet, url, header)
}

func send(t *testing.T, method, url string, header http.Header) response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), method, url, nil)
	require.NoError(t, err)

	req.Header = header

	resp, err := labtest.Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return response{Response: resp, body: body}
}

func TestTheRegisterGivesOneBaseURIForThePublicAndTheProductEndpoints(t *testing.T) {
	endpoints := registeredEndpoints(t)

	assert.True(t, strings.HasPrefix(endpoints.PublicBaseURI, "https://"), endpoints.PublicBaseURI)
	assert.Equal(t, endpoints.PublicBaseURI, endpoints.ProductBaseURI)
}

func TestAClientFindsTheBrandThroughTheRegisterAndAnUnknownPathGivesResourceNotFound(t *testing.T) {
	resp := get(t, publicBaseURI(t)+"/", http.Header{})

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var body struct {
		Errors []struct {
			Code   string `json:"code"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		} `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &body), string(resp.body))
	require.Len(t, body.Errors, 1)
	assert.Equal(t, "urn:au-cds:error:cds-all:Resource/NotFound", body.Errors[0].Code)
	assert.Equal(t, "Resource Not Found", body.Errors[0].Title)
	assert.NotEmpty(t, body.Errors[0].Detail)
}

func TestEachResponseAllowsAnyOrigin(t *testing.T) {
	resp := get(t, publicBaseURI(t)+"/", http.Header{"Origin": {"https://app.example"}})

	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestEachResponseHasTheInteractionIDOfTheRequestOrANewUUID(t *testing.T) {
	t.Run("from the request", func(t *testing.T) {
		const id = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

		resp := get(t, publicBaseURI(t)+"/", http.Header{"X-Fapi-Interaction-Id": {id}})

		assert.Equal(t, id, resp.Header.Get("x-fapi-interaction-id"))
	})

	t.Run("new when the request has none", func(t *testing.T) {
		first := get(t, publicBaseURI(t)+"/", http.Header{}).Header.Get("x-fapi-interaction-id")
		second := get(t, publicBaseURI(t)+"/", http.Header{}).Header.Get("x-fapi-interaction-id")

		_, err := uuid.Parse(first)
		require.NoError(t, err)
		assert.NotEqual(t, first, second)
	})
}

func TestThePublicAPIsAcceptOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t *testing.T) {
	labtest.AssertAcceptsOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t, publicBaseURI(t))
}

func TestThePublicAPIsPresentACertificateFromThePublicCAOnly(t *testing.T) {
	labtest.AssertPresentsACertificateFromThePublicCAOnly(t, publicBaseURI(t))
}

func TestThePublicAPIsDoNotRequestAClientCertificate(t *testing.T) {
	labtest.AssertDoesNotRequestAClientCertificate(t, publicBaseURI(t))
}
