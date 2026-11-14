package root

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Participants of seed/example.json.
const (
	seedDataHolder     = "a81975c6-1ac9-4fe4-88fc-68f50efd27cb"
	seedHolderBrand    = "293b45de-2a40-4e1b-ac77-324dacccb2c8"
	removedDataHolder  = "91bbf006-ca02-47b6-99c3-ddc63541e1c6"
	removedHolderBrand = "4ce2ed25-ed3f-429c-8cbd-24d5041a7c93"
)

// IDs that are not in seed/example.json.
const (
	newDataHolder = "5a0c3f6e-8d8e-4b8e-9d7e-3b3f1c0b2a01"
	newBrand      = "5a0c3f6e-8d8e-4b8e-9d7e-3b3f1c0b2a02"
	newInterim    = "5a0c3f6e-8d8e-4b8e-9d7e-3b3f1c0b2a03"
)

func newBrandJSON(id, name string) string {
	return `{
		"dataHolderBrandId": "` + id + `", "brandName": "` + name + `", "industries": ["banking"],
		"logoUri": "https://new.example/brand.png", "status": "ACTIVE",
		"endpointDetail": {
			"version": "1", "publicBaseUri": "https://api.new.example", "resourceBaseUri": "https://mtls.new.example",
			"infosecBaseUri": "https://id.new.example", "websiteUri": "https://new.example"
		},
		"jwksEndpoint": "https://new.example/jwks"
	}`
}

func newDataHolderJSON(legalEntityID, brand string) string {
	return `{
		"legalEntityId": "` + legalEntityID + `", "legalEntityName": "New Holder Ltd",
		"logoUri": "https://new.example/logo.png", "status": "ACTIVE",
		"dataHolderBrands": [` + brand + `]
	}`
}

func send(t *testing.T, handlers map[int]http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	return serveOperatorHost(t, handlers, operatorHost, method, path, authorisedOperator(), []byte(body))
}

// sendOK checks the success status, and gives the public view of the data holders after the request.
func sendOK(t *testing.T, handlers map[int]http.Handler, method, path, body string, want int) holderView {
	t.Helper()

	rec := send(t, handlers, method, path, body)
	require.Equal(t, want, rec.Code, rec.Body.String())

	return observeHolders(t, handlers)
}

// holderView is the data holders on the public APIs: the brand names on Get Data Holder Brands Summary by brand ID or
// interim ID, the statuses on Get Data Holder Statuses, and the ETag of each API.
type holderView struct {
	brandNames   map[string]string
	statuses     map[string]string
	summaryETag  string
	statusesETag string
}

func observeHolders(t *testing.T, handlers map[int]http.Handler) holderView {
	t.Helper()

	summary := getPublic(t, handlers, "/cdr-register/v1/banking/data-holders/brands/summary", "2")
	statuses := getPublic(t, handlers, "/cdr-register/v1/all/data-holders/status", "2")

	var body struct {
		Data []struct {
			DataHolderBrandID string `json:"dataHolderBrandId"`
			InterimID         string `json:"interimId"`
			BrandName         string `json:"brandName"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(summary.Body.Bytes(), &body))

	v := holderView{
		brandNames:   map[string]string{},
		statuses:     statusesOf(t, statuses, "legalEntityId"),
		summaryETag:  summary.Header().Get("ETag"),
		statusesETag: statuses.Header().Get("ETag"),
	}

	for _, b := range body.Data {
		id := b.DataHolderBrandID
		if id == "" {
			id = b.InterimID
		}

		v.brandNames[id] = b.BrandName
	}

	return v
}

func TestTheOperatorCreatesChangesAndRemovesADataHolderAndThePublicAPIsShowEachChange(t *testing.T) {
	handlers := lab(t)
	holder := "/v1/data-holders/" + newDataHolder
	brand := holder + "/brands/" + newBrand

	before := observeHolders(t, handlers)

	rec := send(
		t,
		handlers,
		http.MethodPost,
		"/v1/data-holders",
		newDataHolderJSON(newDataHolder, newBrandJSON(newBrand, "New Bank")),
	)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, holder, rec.Header().Get("Location"))

	created := observeHolders(t, handlers)
	assert.Equal(t, "New Bank", created.brandNames[newBrand])
	assert.Equal(t, "ACTIVE", created.statuses[newDataHolder])
	assert.NotEqual(t, before.summaryETag, created.summaryETag)
	assert.NotEqual(t, before.statusesETag, created.statusesETag)

	renamed := sendOK(t, handlers, http.MethodPut, brand, `{
		"brandName": "Renamed Bank", "industries": ["banking"], "logoUri": "https://new.example/brand.png",
		"endpointDetail": {
			"version": "1", "publicBaseUri": "https://api.new.example", "resourceBaseUri": "https://mtls.new.example",
			"infosecBaseUri": "https://id.new.example", "websiteUri": "https://new.example"
		},
		"jwksEndpoint": "https://new.example/jwks"
	}`, http.StatusNoContent)
	assert.Equal(t, "Renamed Bank", renamed.brandNames[newBrand])
	assert.NotEqual(t, created.summaryETag, renamed.summaryETag)

	added := sendOK(t, handlers, http.MethodPost, holder+"/brands", `{
		"interimId": "`+newInterim+`", "brandName": "New Products", "industries": ["banking"],
		"logoUri": "https://new.example/products.png", "status": "ACTIVE",
		"endpointDetail": {"publicBaseUri": "https://products.new.example"}, "jwksEndpoint": ""
	}`, http.StatusCreated)
	assert.Equal(t, "New Products", added.brandNames[newInterim])

	deactivated := sendOK(t, handlers, http.MethodPost, brand+"/deactivate", "", http.StatusNoContent)
	assert.Contains(t, deactivated.brandNames, newBrand, "the summary gives INACTIVE brands")

	removedBrand := sendOK(t, handlers, http.MethodPost, brand+"/remove", "", http.StatusNoContent)
	assert.NotContains(t, removedBrand.brandNames, newBrand, "the summary does not give REMOVED brands")
	assert.Contains(t, removedBrand.brandNames, newInterim)

	removed := sendOK(t, handlers, http.MethodPost, holder+"/remove", "", http.StatusNoContent)
	assert.Equal(t, "REMOVED", removed.statuses[newDataHolder])
	assert.NotContains(t, removed.brandNames, newInterim, "the removal of the data holder removes its brands")
}

func TestTheOperatorCannotMakeADataHolderChangeThatTheRulesDoNotAllow(t *testing.T) {
	handlers := lab(t)

	tests := []struct {
		name, method, path, body string
		status                   int
		detail                   string
	}{
		{
			name: "a data holder that exists", method: http.MethodPost, path: "/v1/data-holders",
			body:   newDataHolderJSON(seedDataHolder, newBrandJSON(newBrand, "New Bank")),
			status: http.StatusConflict, detail: "already exists",
		},
		{
			name: "an ID of a removed brand", method: http.MethodPost, path: "/v1/data-holders",
			body:   newDataHolderJSON(newDataHolder, newBrandJSON(removedHolderBrand, "New Bank")),
			status: http.StatusConflict, detail: `the ID "` + removedHolderBrand + `" is not unique`,
		},
		{
			name: "a data holder that is not valid", method: http.MethodPost, path: "/v1/data-holders",
			body:   newDataHolderJSON(newDataHolder, newBrandJSON(newBrand, "")),
			status: http.StatusBadRequest, detail: "brandName is required",
		},
		{
			name: "a status in a change", method: http.MethodPut, path: "/v1/data-holders/" + seedDataHolder,
			body:   `{"legalEntityName": "Holder", "logoUri": "https://holder.example/logo.png", "status": "REMOVED"}`,
			status: http.StatusBadRequest, detail: `unknown field "status"`,
		},
		{
			name:   "activate a brand of a removed data holder",
			method: http.MethodPost,
			path:   "/v1/data-holders/" + removedDataHolder + "/brands/" + removedHolderBrand + "/activate",
			status: http.StatusConflict,
			detail: "REMOVED is final",
		},
		{
			name:   "remove a removed data holder",
			method: http.MethodPost, path: "/v1/data-holders/" + removedDataHolder + "/remove",
			status: http.StatusConflict, detail: "REMOVED is final",
		},
		{
			name:   "a brand that does not exist",
			method: http.MethodPost, path: "/v1/data-holders/" + seedDataHolder + "/brands/" + unknownID + "/remove",
			status: http.StatusNotFound, detail: "does not exist",
		},
		{
			name:   "a data holder that does not exist",
			method: http.MethodPost, path: "/v1/data-holders/" + unknownID + "/remove",
			status: http.StatusNotFound, detail: "does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := observeHolders(t, handlers)

			rec := send(t, handlers, tt.method, tt.path, tt.body)

			assert.Equal(t, tt.status, rec.Code, rec.Body.String())
			assert.Contains(t, problemOf(t, rec).Detail, tt.detail)
			assert.Equal(t, before, observeHolders(t, handlers), "the refused request changes nothing")
		})
	}
}
