package root

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IDs that are not in seed/example.json.
const (
	newRecipient       = "7e1f0c2a-3b4d-4e5f-8a9b-0c1d2e3f4a01"
	newRecipientBrand  = "7e1f0c2a-3b4d-4e5f-8a9b-0c1d2e3f4a02"
	newProduct         = "7e1f0c2a-3b4d-4e5f-8a9b-0c1d2e3f4a03"
	secondNewProduct   = "7e1f0c2a-3b4d-4e5f-8a9b-0c1d2e3f4a04"
	mockBudgetInactive = "ffea3f81-674a-4ba8-bfd2-9bd576c2fe9d" // an INACTIVE brand of mockBudgetRecipient
)

func newProductJSON(id, name string) string {
	return `{
		"softwareProductId": "` + id + `", "softwareProductName": "` + name + `",
		"softwareProductDescription": "A new app", "logoUri": "https://new.example/app.png", "status": "ACTIVE",
		"clientUri": "https://new.example", "redirectUris": ["https://new.example/callback"],
		"jwksUri": "https://new.example/jwks", "revocationUri": "https://new.example/revoke",
		"recipientBaseUri": "https://cdr.new.example", "scopes": ["openid"]
	}`
}

func newRecipientJSON() string {
	return `{
		"legalEntityId": "` + newRecipient + `", "legalEntityName": "New Recipient Pty Ltd",
		"accreditationNumber": "ADRNEW000001", "accreditationLevel": "UNRESTRICTED",
		"logoUri": "https://new.example/logo.png", "status": "ACTIVE",
		"dataRecipientBrands": [{
			"dataRecipientBrandId": "` + newRecipientBrand + `", "brandName": "New App",
			"logoUri": "https://new.example/brand.png", "status": "ACTIVE",
			"softwareProducts": [` + newProductJSON(newProduct, "New App Web") + `]
		}]
	}`
}

// productNames gives the name of each software product on Get Data Recipients, by ID.
func productNames(t *testing.T, handlers map[int]http.Handler) map[string]string {
	t.Helper()

	var body struct {
		Data []struct {
			Brands []struct {
				SoftwareProducts []struct {
					SoftwareProductID   string `json:"softwareProductId"`
					SoftwareProductName string `json:"softwareProductName"`
				} `json:"softwareProducts"`
			} `json:"dataRecipientBrands"`
		} `json:"data"`
	}
	require.NoError(
		t,
		json.Unmarshal(getPublic(t, handlers, "/cdr-register/v1/all/data-recipients", "4").Body.Bytes(), &body),
	)

	names := map[string]string{}

	for _, dr := range body.Data {
		for _, b := range dr.Brands {
			for _, sp := range b.SoftwareProducts {
				names[sp.SoftwareProductID] = sp.SoftwareProductName
			}
		}
	}

	return names
}

// changeRecipient checks the success status, and that the public APIs show the statuses after the request.
func changeRecipient(
	t *testing.T,
	handlers map[int]http.Handler,
	method, path, body string,
	wantCode int,
	c statusChange,
) {
	t.Helper()

	before := observe(t, handlers)

	rec := send(t, handlers, method, path, body)
	require.Equal(t, wantCode, rec.Code, rec.Body.String())

	after := observe(t, handlers)

	for _, list := range []string{recipientStatusList, dataRecipientList} {
		for id, want := range c.recipients {
			assert.Equal(t, want, after.recipients[list][id], "%s: data recipient %s", list, id)
		}
	}

	for _, list := range []string{softwareProductList, dataRecipientList} {
		for id, want := range c.products {
			assert.Equal(t, want, after.products[list][id], "%s: software product %s", list, id)
		}
	}

	for _, api := range c.changedAPIs {
		assert.NotEqual(t, before.etags[api], after.etags[api], "the ETag of %s did not change", api)
	}
}

// Verifies: REG-0809, REG-0811, REG-0817, REG-0828, REG-0829
func TestTheOperatorCreatesAndChangesADataRecipientAndThePublicAPIsShowEachChange(t *testing.T) {
	handlers := lab(t)
	recipient := "/v1/data-recipients/" + newRecipient
	brand := recipient + "/brands/" + newRecipientBrand
	allAPIs := []string{recipientStatusList, softwareProductList, dataRecipientList}

	changeRecipient(t, handlers, http.MethodPost, "/v1/data-recipients", newRecipientJSON(), http.StatusCreated,
		statusChange{
			recipients:  map[string]string{newRecipient: "ACTIVE"},
			products:    map[string]string{newProduct: "ACTIVE"},
			changedAPIs: allAPIs,
		})

	changeRecipient(t, handlers, http.MethodPost, brand+"/software-products",
		newProductJSON(secondNewProduct, "New App Mobile"), http.StatusCreated, statusChange{
			products:    map[string]string{secondNewProduct: "ACTIVE"},
			changedAPIs: []string{softwareProductList, dataRecipientList},
		})

	rec := send(t, handlers, http.MethodPut, recipient+"/software-products/"+newProduct, `{
		"softwareProductName": "Renamed Web", "softwareProductDescription": "A new app",
		"logoUri": "https://new.example/app.png", "clientUri": "https://new.example",
		"redirectUris": ["https://new.example/callback"], "jwksUri": "https://new.example/jwks",
		"revocationUri": "https://new.example/revoke", "recipientBaseUri": "https://cdr.new.example",
		"scopes": ["openid"]
	}`)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	assert.Equal(t, "Renamed Web", productNames(t, handlers)[newProduct])

	changeRecipient(t, handlers, http.MethodPost, brand+"/deactivate", "", http.StatusNoContent, statusChange{
		products:    map[string]string{newProduct: "INACTIVE", secondNewProduct: "INACTIVE"},
		changedAPIs: []string{softwareProductList, dataRecipientList},
	})

	changeRecipient(t, handlers, http.MethodPost, brand+"/remove", "", http.StatusNoContent, statusChange{
		recipients:  map[string]string{newRecipient: "ACTIVE"},
		products:    map[string]string{newProduct: "REMOVED", secondNewProduct: "REMOVED"},
		changedAPIs: []string{softwareProductList, dataRecipientList},
	})
}

// Verifies: REG-0810, REG-0855, REG-0856, REG-0858
func TestTheOperatorCannotMakeADataRecipientChangeThatTheRulesDoNotAllow(t *testing.T) {
	handlers := lab(t)
	budget := "/v1/data-recipients/" + mockBudgetRecipient

	tests := []struct {
		name, method, path, body string
		status                   int
		detail                   string
	}{
		{
			name: "a software product with the ID of a removed software product", method: http.MethodPost,
			path:   budget + "/brands/" + mockBudgetInactive + "/software-products",
			body:   newProductJSON(mockBudgetBeta, "Beta"),
			status: http.StatusConflict, detail: `the ID "` + mockBudgetBeta + `" is not unique`,
		},
		{
			name: "an active software product of an inactive brand", method: http.MethodPost,
			path:   budget + "/brands/" + mockBudgetInactive + "/software-products",
			body:   newProductJSON(newProduct, "New"),
			status: http.StatusConflict, detail: "a software product of an INACTIVE data recipient brand",
		},
		{
			name: "details that are not complete", method: http.MethodPut, path: budget,
			body:   `{"legalEntityName": "Mock Budgeting Pty Ltd"}`,
			status: http.StatusBadRequest, detail: "accreditationNumber is required",
		},
		{
			name:   "a brand that does not exist",
			method: http.MethodPost,
			path:   budget + "/brands/" + unknownID + "/remove",
			status: http.StatusNotFound,
			detail: "does not exist",
		},
		{
			name: "a data recipient that exists", method: http.MethodPost, path: "/v1/data-recipients",
			body:   `{"legalEntityId": "` + mockBudgetRecipient + `"}`,
			status: http.StatusConflict, detail: "already exists",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := observe(t, handlers)

			rec := send(t, handlers, tt.method, tt.path, tt.body)

			assert.Equal(t, tt.status, rec.Code, rec.Body.String())
			assert.Contains(t, problemOf(t, rec).Detail, tt.detail)
			assert.Equal(t, before, observe(t, handlers), "the refused request changes nothing")
		})
	}
}
