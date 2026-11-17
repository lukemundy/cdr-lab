package console_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const (
	addBrand           = "/api/data-recipients/dr-1/brands"
	addSoftwareProduct = "/api/data-recipients/dr-1/brands/dr-1-b1/software-products"
	addedBrandFields   = `"brandName": "Gamma", "logoUri": "https://gamma.example/brand.png", "status": "ACTIVE"`
	addedProductFields = `"softwareProductName": "Gamma App", "softwareProductDescription": "An app",
		"logoUri": "https://gamma.example/app.png", "clientUri": "https://gamma.example",
		"sectorIdentifierUri": "https://gamma.example/sector", "redirectUris": ["https://gamma.example/callback"],
		"tosUri": "https://gamma.example/tos", "policyUri": "https://gamma.example/policy",
		"jwksUri": "https://gamma.example/jwks", "revocationUri": "https://gamma.example/revoke",
		"recipientBaseUri": "https://cdr.gamma.example", "scopes": ["openid", "profile"]`
)

func newBrandRequest(version, brandID string) string {
	return `{"version": ` + version + `, "dataRecipientBrandId": "` + brandID + `", ` + addedBrandFields + `}`
}

func newProductRequest(version, productID, status string) string {
	return `{"version": ` + version + `, "softwareProductId": "` + productID + `", "status": "` + status + `", ` +
		addedProductFields + `}`
}

func versionOf(t *testing.T, ports console.Ports) int64 {
	t.Helper()

	return detailOf(t, ports, "dr-1").Version
}

// Verifies: REG-1302, REG-1326
func TestAnAddOfABrandAtTheCurrentVersionMakesABrandWithNoSoftwareProductsAndGivesItsPath(t *testing.T) {
	ports := seedForCreates(t)

	resp := post(t, ports, addBrand, newBrandRequest("1", "brand-new"))

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.Equal(t, "/api/data-recipients/dr-1/brands/brand-new", resp.header.Get("Location"))
	assert.Empty(t, resp.body)

	var detail map[string]any
	require.NoError(
		t,
		json.Unmarshal([]byte(withoutActions(t, get(t, ports.Readers, "/api/data-recipients/dr-1").body)),
			&detail),
	)
	brands := detail["dataRecipientBrands"].([]any)
	require.Len(t, brands, 2)
	added, err := json.Marshal(brands[1])
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"dataRecipientBrandId": "brand-new", "brandName": "Gamma", "logoUri": "https://gamma.example/brand.png",
		"status": "ACTIVE", "softwareProducts": []
	}`, string(added))
	assert.Equal(t, int64(2), versionOf(t, ports))

	history := get(t, ports.Readers, "/api/data-recipients/dr-1/history?pageSize=1")
	assert.Contains(t, history.body, `"eventType":"DataRecipientBrandAdded"`)
}

// Verifies: REG-1304, REG-1326
func TestAnAddOfASoftwareProductAtTheCurrentVersionMakesItWithNoClientCertificatesAndGivesItsPath(t *testing.T) {
	ports := seedForCreates(t)

	resp := post(t, ports, addSoftwareProduct, newProductRequest("1", "sp-new", "INACTIVE"))

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.Equal(t, "/api/data-recipients/dr-1/software-products/sp-new", resp.header.Get("Location"))
	assert.Empty(t, resp.body)

	var detail map[string]any
	require.NoError(
		t,
		json.Unmarshal([]byte(withoutActions(t, get(t, ports.Readers, "/api/data-recipients/dr-1").body)),
			&detail),
	)
	products := detail["dataRecipientBrands"].([]any)[0].(map[string]any)["softwareProducts"].([]any)
	require.Len(t, products, 2)
	added, err := json.Marshal(products[1])
	require.NoError(t, err)
	assert.JSONEq(t, `{"softwareProductId": "sp-new", "status": "INACTIVE", "clientCertificateThumbprints": [], `+
		addedProductFields+`}`, string(added))
	assert.Equal(t, int64(2), versionOf(t, ports))

	history := get(t, ports.Readers, "/api/data-recipients/dr-1/history?pageSize=1")
	assert.Contains(t, history.body, `"eventType":"SoftwareProductAdded"`)
}

// Verifies: REG-1321
func TestAnAddGivesTheLocationOfAnIdentifierWithCharactersOfAPath(t *testing.T) {
	ports := seedForCreates(t)

	brand := post(t, ports, addBrand, newBrandRequest("1", "brand/new?x"))
	require.Equal(t, http.StatusCreated, brand.status, brand.body)
	assert.Equal(t, "/api/data-recipients/dr-1/brands/brand%2Fnew%3Fx", brand.header.Get("Location"))

	product := post(t, ports, addSoftwareProduct, newProductRequest("2", "sp/new?x", "ACTIVE"))
	require.Equal(t, http.StatusCreated, product.status, product.body)
	assert.Equal(t, "/api/data-recipients/dr-1/software-products/sp%2Fnew%3Fx", product.header.Get("Location"))
}

// Verifies: REG-1329
func TestAnAddWithAnIdentifierThatADifferentRecordHasGivesAnErrorAtTheIdentifier(t *testing.T) {
	// A data recipient, a data holder, a brand and a software product have these IDs.
	for _, id := range []string{"dr-1", "dh-1", "dr-1-b1", "sp-1"} {
		for _, add := range []struct{ target, body, field string }{
			{target: addBrand, body: newBrandRequest("1", id), field: "dataRecipientBrandId"},
			{target: addSoftwareProduct, body: newProductRequest("1", id, "ACTIVE"), field: "softwareProductId"},
		} {
			ports := seedForCreates(t)

			resp := post(t, ports, add.target, add.body)

			require.Equal(t, http.StatusConflict, resp.status, "%s %s", add.field, id)
			assert.JSONEq(t, `{
				"type": "`+fieldErrorsType+`",
				"title": "Values in the request are not valid",
				"status": 409,
				"detail": "One or more values in the request are not valid.",
				"errors": [
					{"pointer": "#/`+add.field+`", "detail": "`+add.field+` is not unique: a different record has it"}
				]
			}`, resp.body, "%s %s", add.field, id)
			assert.Equal(t, int64(1), versionOf(t, ports), "%s %s", add.field, id)
		}
	}
}

// Verifies: REG-1328, REG-1329
func TestAnAddWithValuesThatTheParticipantModelRefusesGivesAnErrorAtEachField(t *testing.T) {
	ports := seedForCreates(t)

	brand := post(t, ports, addBrand, `{"version": 1, "dataRecipientBrandId": "`+strings.Repeat("x", 37)+`",
		"brandName": "", "logoUri": "https://gamma.example/brand.png", "status": "SUSPENDED"}`)

	require.Equal(t, http.StatusBadRequest, brand.status, brand.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/dataRecipientBrandId", "detail": "dataRecipientBrandId is longer than 36 characters"},
			{"pointer": "#/brandName", "detail": "brandName is required"},
			{"pointer": "#/status", "detail": "status \"SUSPENDED\" is not ACTIVE, INACTIVE or REMOVED"}
		]
	}`, brand.body)

	product := post(t, ports, addSoftwareProduct, `{"version": 1, "softwareProductId": "sp-new",
		"softwareProductName": "Gamma App", "softwareProductDescription": "An app",
		"logoUri": "https://gamma.example/app.png", "clientUri": "https://gamma.example", "redirectUris": [],
		"jwksUri": "https://gamma.example/jwks", "revocationUri": "https://gamma.example/revoke",
		"recipientBaseUri": "https://cdr.gamma.example", "scopes": ["openid", "profile"], "status": "ACTIVE"}`)

	require.Equal(t, http.StatusBadRequest, product.status, product.body)

	var body fieldProblem
	require.NoError(t, json.Unmarshal([]byte(product.body), &body))
	require.Len(t, body.Errors, 1)
	assert.Equal(t, "#/redirectUris", body.Errors[0].Pointer)
	assert.Equal(t, "redirectUris is required", body.Errors[0].Detail)
	assert.Equal(t, int64(1), versionOf(t, ports))
}

// Verifies: REG-1496, REG-1497
func TestAnAddOfASoftwareProductWithScopesThatTheSSANeedsGivesAnErrorAtTheScopes(t *testing.T) {
	ports := seedForCreates(t)

	tests := map[string]struct{ scopes, detail string }{
		"no profile":     {`["openid", "bank:accounts.basic:read"]`, "scopes does not contain openid and profile"},
		"an empty value": {`["openid", "profile", ""]`, "scopes has an empty value"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			product := post(t, ports, addSoftwareProduct, `{"version": 1, "softwareProductId": "sp-new",
				"softwareProductName": "Gamma App", "softwareProductDescription": "An app",
				"logoUri": "https://gamma.example/app.png", "clientUri": "https://gamma.example",
				"redirectUris": ["https://gamma.example/callback"],
				"jwksUri": "https://gamma.example/jwks", "revocationUri": "https://gamma.example/revoke",
				"recipientBaseUri": "https://cdr.gamma.example", "scopes": `+tt.scopes+`, "status": "ACTIVE"}`)

			require.Equal(t, http.StatusBadRequest, product.status, product.body)

			var body fieldProblem
			require.NoError(t, json.Unmarshal([]byte(product.body), &body))
			require.Len(t, body.Errors, 1)
			assert.Equal(t, "#/scopes", body.Errors[0].Pointer)
			assert.Equal(t, tt.detail, body.Errors[0].Detail)
			assert.Equal(t, int64(1), versionOf(t, ports))
		})
	}
}

// Verifies: REG-1329
func TestAnActiveSoftwareProductOfAnInactiveBrandGivesAConflictAtTheStatus(t *testing.T) {
	dr := recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive, []domain.SoftwareProduct{
		product("sp-1", domain.Inactive),
	})
	dr.Brands[0].Status = domain.Inactive
	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{dr}})
	ports := console.Ports{Readers: queries, Operator: app.NewOperator(store)}

	resp := post(t, ports, addSoftwareProduct, newProductRequest("1", "sp-new", "ACTIVE"))

	require.Equal(t, http.StatusConflict, resp.status, resp.body)

	var body fieldProblem
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	assert.Equal(t, fieldErrorsType, body.Type)
	require.Len(t, body.Errors, 1)
	assert.Equal(t, "#/status", body.Errors[0].Pointer)
	assert.Contains(t, body.Errors[0].Detail, "status is ACTIVE, but data recipient brand")
	assert.Equal(t, int64(1), versionOf(t, ports))
}

// Verifies: REG-1330
func TestAnAddWithAVersionThatIsNotCurrentChangesNothingAndHasNoFieldErrors(t *testing.T) {
	ports := seedForCreates(t)
	require.Equal(t, http.StatusNoContent, post(t, ports, changeDetails, newDetailsRequest).status)

	// The brand name is not valid too, but the version comes first.
	for _, add := range []struct{ target, body string }{
		{target: addBrand, body: `{"version": 1, "dataRecipientBrandId": "brand-new", "brandName": "",
			"logoUri": "https://gamma.example/brand.png", "status": "ACTIVE"}`},
		{target: addSoftwareProduct, body: newProductRequest("1", "sp-new", "ACTIVE")},
	} {
		resp := post(t, ports, add.target, add.body)

		require.Equal(t, http.StatusConflict, resp.status, add.target)

		var body fieldProblem
		require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
		assert.Equal(t, "about:blank", body.Type)
		assert.Equal(t, "The participant changed after the version in the request. The Register changed nothing.",
			body.Detail)
		assert.Nil(t, body.Errors)
	}

	assert.Equal(t, int64(2), versionOf(t, ports))
}

func TestAnAddToARecordThatDoesNotExistGivesNotFoundWithNoIdentifier(t *testing.T) {
	ports := seedForCreates(t)

	for _, add := range []struct{ target, body string }{
		{target: "/api/data-recipients/dr-unknown/brands", body: newBrandRequest("1", "brand-new")},
		{
			target: "/api/data-recipients/dr-unknown/brands/dr-1-b1/software-products",
			body:   newProductRequest("1", "sp-new", "ACTIVE"),
		},
		{
			target: "/api/data-recipients/dr-1/brands/brand-unknown/software-products",
			body:   newProductRequest("1", "sp-new", "ACTIVE"),
		},
	} {
		resp := post(t, ports, add.target, add.body)

		require.Equal(t, http.StatusNotFound, resp.status, add.target)
		assert.NotContains(t, resp.body, "unknown", add.target)
	}
}

// Verifies: REG-1327
func TestAnAddRefusesABodyThatIsNotValid(t *testing.T) {
	ports := seedForCreates(t)

	for target, bodies := range map[string][]string{
		addBrand: {
			``, `[]`, `null`, `{}`, `{"dataRecipientBrandId": "brand-new", ` + addedBrandFields + `}`,
			`{"version": 0, "dataRecipientBrandId": "brand-new", ` + addedBrandFields + `}`,
			`{"version": "1", "dataRecipientBrandId": "brand-new", ` + addedBrandFields + `}`,
			newBrandRequest("1", "brand-new") + ` {}`,
			`{"version": 1, "dataRecipientBrandId": "brand-new", ` + addedBrandFields + `, "softwareProducts": []}`,
		},
		addSoftwareProduct: {
			`{"softwareProductId": "sp-new", "status": "ACTIVE", ` + addedProductFields + `}`,
			`{"version": 1, "softwareProductId": "sp-new", "status": "ACTIVE", ` + addedProductFields +
				`, "clientCertificates": []}`,
			`{"version": 1, "softwareProductId": 5}`,
		},
	} {
		for _, body := range bodies {
			resp := post(t, ports, target, body)

			require.Equal(t, http.StatusBadRequest, resp.status, "%s body %.60s", target, body)
			assert.Contains(t, resp.body, `"type":"about:blank"`, "%s body %.60s", target, body)
		}
	}

	assert.Equal(t, int64(1), versionOf(t, ports))
}

func TestAnAddGivesAProblemWithNoErrorMessageWhenTheAddFails(t *testing.T) {
	ports := console.Ports{Operator: failingOperator{}}

	for target, body := range map[string]string{
		addBrand:           newBrandRequest("1", "brand-new"),
		addSoftwareProduct: newProductRequest("1", "sp-new", "ACTIVE"),
	} {
		resp := post(t, ports, target, body)

		assert.Equal(t, http.StatusInternalServerError, resp.status, target)
		assert.NotContains(t, resp.body, "Mock Budgeting", target)
	}
}
