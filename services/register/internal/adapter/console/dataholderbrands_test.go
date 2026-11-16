package console_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const (
	addHolderBrand    = "/api/data-holders/dh-1/brands"
	changeHolderBrand = "/api/data-holders/dh-1/brands/dh-1-b1/change"
	holderBrandFields = `"brandName": "Delta Bank", "brandGroup": "Delta Group", "industries": ["banking", "energy"],
		"logoUri": "https://delta.example/brand.png",
		"endpointDetail": {
			"version": "1", "publicBaseUri": "https://api.delta.example",
			"resourceBaseUri": "https://mtls.delta.example",
			"infosecBaseUri": "https://id.delta.example", "websiteUri": "https://delta.example"
		},
		"jwksEndpoint": "https://id.delta.example/jwks"`
	// A brand with no dataHolderBrandId needs only its public details.
	publicHolderBrandFields = `"brandName": "Delta Products", "industries": ["banking"],
		"logoUri": "https://delta.example/products.png",
		"endpointDetail": {"publicBaseUri": "https://products.delta.example"}`
)

func newHolderBrandRequest(version, idFields string) string {
	return `{"version": ` + version + `, ` + idFields + `, "status": "ACTIVE", ` + holderBrandFields + `}`
}

// seedForHolderBrands gives dh-1 with the brand dh-1-b1 and the brand with only the interim ID dh-1-i2, and dh-2,
// which is REMOVED. dr-1 is a data recipient with the brand dr-1-b1 and the software product sp-1.
func seedForHolderBrands(t *testing.T) console.Ports {
	t.Helper()

	interim := domain.DataHolderBrand{
		InterimID: "dh-1-i2", BrandName: "Holder Products", Industries: []domain.Industry{domain.Telco},
		LogoURI: "https://holder.example/products.png", Status: domain.Active,
		EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://products.holder.example"},
	}
	store, queries := seedStore(t, domain.Participants{
		DataRecipients: []domain.DataRecipient{
			recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
				[]domain.SoftwareProduct{product("sp-1", domain.Active)}),
		},
		DataHolders: []domain.DataHolder{
			dataHolder("dh-1", "Holder Pty Ltd", domain.DataHolderActive,
				brand("dh-1-b1", "Holder Bank", domain.Active, domain.Banking), interim),
			dataHolder("dh-2", "Removed Pty Ltd", domain.DataHolderRemoved),
		},
	})

	return console.Ports{Readers: queries, Operator: app.NewOperator(store)}
}

type holderBrandsBody struct {
	Version int64             `json:"version"`
	Brands  []json.RawMessage `json:"dataHolderBrands"`
}

func holderBrandsOf(t *testing.T, ports console.Ports, legalEntityID string) holderBrandsBody {
	t.Helper()

	resp := get(t, ports.Readers, "/api/data-holders/"+legalEntityID)
	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var body holderBrandsBody
	require.NoError(t, json.Unmarshal([]byte(withoutHolderActions(t, resp.body)), &body))

	return body
}

func TestAnAddOfADataHolderBrandAtTheCurrentVersionMakesTheBrandAndGivesItsPath(t *testing.T) {
	ports := seedForHolderBrands(t)

	resp := post(t, ports, addHolderBrand, newHolderBrandRequest("1", `"dataHolderBrandId": "brand/new"`))

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.Equal(t, "/api/data-holders/dh-1/brands/brand%2Fnew", resp.header.Get("Location"))
	assert.Empty(t, resp.body)

	detail := holderBrandsOf(t, ports, "dh-1")
	assert.Equal(t, int64(2), detail.Version)
	require.Len(t, detail.Brands, 3)
	assert.JSONEq(t, `{"dataHolderBrandId": "brand/new", "status": "ACTIVE", `+holderBrandFields+`}`,
		string(detail.Brands[2]))

	history := get(t, ports.Readers, "/api/data-holders/dh-1/history?pageSize=1")
	assert.Contains(t, history.body, `"eventType":"DataHolderBrandAdded"`)
}

func TestAnAddOfADataHolderBrandWithOnlyAnInterimIDNeedsOnlyThePublicDetails(t *testing.T) {
	ports := seedForHolderBrands(t)

	resp := post(t, ports, addHolderBrand,
		`{"version": 1, "interimId": "interim-new", "status": "INACTIVE", `+publicHolderBrandFields+`}`)

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.Equal(t, "/api/data-holders/dh-1/brands/interim-new", resp.header.Get("Location"))

	detail := holderBrandsOf(t, ports, "dh-1")
	require.Len(t, detail.Brands, 3)
	assert.JSONEq(t, `{"interimId": "interim-new", "status": "INACTIVE", `+publicHolderBrandFields+`}`,
		string(detail.Brands[2]))
}

func TestAnAddOfADataHolderBrandWithAnIdentifierThatADifferentRecordHasGivesAnErrorAtThatIdentifier(t *testing.T) {
	// A data holder, a data holder brand, an interim ID, a data recipient, a brand and a software product have these
	// IDs.
	for _, id := range []string{"dh-1", "dh-1-b1", "dh-1-i2", "dh-2", "dr-1", "dr-1-b1", "sp-1"} {
		for _, add := range []struct{ idFields, field string }{
			{idFields: `"dataHolderBrandId": "` + id + `"`, field: "dataHolderBrandId"},
			{idFields: `"dataHolderBrandId": "brand-new", "interimId": "` + id + `"`, field: "interimId"},
		} {
			ports := seedForHolderBrands(t)

			resp := post(t, ports, addHolderBrand, newHolderBrandRequest("1", add.idFields))

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
			assert.Equal(t, int64(1), holderBrandsOf(t, ports, "dh-1").Version, "%s %s", add.field, id)
		}
	}
}

func TestAnAddOfADataHolderBrandWithValuesThatTheParticipantModelRefusesGivesAnErrorAtEachField(t *testing.T) {
	ports := seedForHolderBrands(t)

	resp := post(t, ports, addHolderBrand, `{"version": 1, "brandName": "", "industries": ["banking", "insurance"],
		"logoUri": "https://delta.example/brand.png", "status": "ACTIVE", "endpointDetail": {"publicBaseUri": ""}}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/dataHolderBrandId", "detail": "dataHolderBrandId or interimId is required"},
			{"pointer": "#/brandName", "detail": "brandName is required"},
			{"pointer": "#/endpointDetail/publicBaseUri", "detail": "endpointDetail.publicBaseUri is required"},
			{"pointer": "#/industries",
			 "detail": "industry \"insurance\" is not banking, energy, non-bank-lending or telco"}
		]
	}`, resp.body)

	secure := post(t, ports, addHolderBrand, `{"version": 1, "dataHolderBrandId": "brand-new", "status": "ACTIVE", `+
		publicHolderBrandFields+`}`)

	require.Equal(t, http.StatusBadRequest, secure.status, secure.body)

	var body fieldProblem
	require.NoError(t, json.Unmarshal([]byte(secure.body), &body))

	pointers := make([]string, 0, len(body.Errors))
	for _, e := range body.Errors {
		pointers = append(pointers, e.Pointer)
	}

	assert.Equal(t, []string{
		"#/endpointDetail/version", "#/endpointDetail/resourceBaseUri", "#/endpointDetail/infosecBaseUri",
		"#/endpointDetail/websiteUri", "#/jwksEndpoint",
	}, pointers)
	assert.Equal(t, int64(1), holderBrandsOf(t, ports, "dh-1").Version)
}

func TestAnActiveBrandOfARemovedDataHolderGivesAConflictAtTheStatus(t *testing.T) {
	ports := seedForHolderBrands(t)

	resp := post(t, ports, "/api/data-holders/dh-2/brands",
		newHolderBrandRequest("1", `"dataHolderBrandId": "brand-new"`))

	require.Equal(t, http.StatusConflict, resp.status, resp.body)

	var body fieldProblem
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	assert.Equal(t, fieldErrorsType, body.Type)
	require.Len(t, body.Errors, 1)
	assert.Equal(t, "#/status", body.Errors[0].Pointer)
	assert.Contains(t, body.Errors[0].Detail, "a brand of a REMOVED data holder must be REMOVED")
	assert.Equal(t, int64(1), holderBrandsOf(t, ports, "dh-2").Version)
}

func TestAnAddOrAChangeOfADataHolderBrandWithAVersionThatIsNotCurrentChangesNothing(t *testing.T) {
	ports := seedForHolderBrands(t)
	require.Equal(t, http.StatusNoContent, post(t, ports, changeHolderDetails, newHolderDetailsRequest).status)

	// The brand name is not valid too, but the version comes first.
	for target, body := range map[string]string{
		addHolderBrand: `{"version": 1, "dataHolderBrandId": "brand-new", "status": "ACTIVE", ` +
			publicHolderBrandFields + `, "brandName": ""}`,
		changeHolderBrand: `{"version": 1, ` + holderBrandFields + `, "brandName": ""}`,
	} {
		resp := post(t, ports, target, body)

		require.Equal(t, http.StatusConflict, resp.status, target)

		var problem fieldProblem
		require.NoError(t, json.Unmarshal([]byte(resp.body), &problem))
		assert.Equal(t, "about:blank", problem.Type, target)
		assert.Equal(t, "The participant changed after the version in the request. The Register changed nothing.",
			problem.Detail, target)
		assert.Nil(t, problem.Errors, target)
	}

	assert.Equal(t, int64(2), holderBrandsOf(t, ports, "dh-1").Version)
}

func TestAnAddOrAChangeOfADataHolderBrandThatDoesNotExistGivesNotFoundWithNoIdentifier(t *testing.T) {
	ports := seedForHolderBrands(t)

	newBrand := newHolderBrandRequest("1", `"dataHolderBrandId": "brand-new"`)
	// dr-1 is a data recipient, not a data holder, and dr-1-b1 is a brand of a data recipient.
	for target, body := range map[string]string{
		"/api/data-holders/dh-unknown/brands":                newBrand,
		"/api/data-holders/dr-1/brands":                      newBrand,
		"/api/data-holders/dh-unknown/brands/dh-1-b1/change": `{"version": 1, ` + holderBrandFields + `}`,
		"/api/data-holders/dh-1/brands/brand-unknown/change": `{"version": 1, ` + holderBrandFields + `}`,
		"/api/data-holders/dh-1/brands/dr-1-b1/change":       `{"version": 1, ` + holderBrandFields + `}`,
	} {
		resp := post(t, ports, target, body)

		require.Equal(t, http.StatusNotFound, resp.status, target)
		assert.NotContains(t, resp.body, "unknown", target)
		assert.NotContains(t, resp.body, "dr-1", target)
	}
}

func TestAnAddOrAChangeOfADataHolderBrandRefusesABodyThatIsNotValid(t *testing.T) {
	ports := seedForHolderBrands(t)

	for target, bodies := range map[string][]string{
		addHolderBrand: {
			``, `[]`, `null`, `{}`,
			`{"dataHolderBrandId": "brand-new", "status": "ACTIVE", ` + holderBrandFields + `}`,
			`{"version": 0, "dataHolderBrandId": "brand-new", "status": "ACTIVE", ` + holderBrandFields + `}`,
			newHolderBrandRequest("1", `"dataHolderBrandId": "brand-new"`) + ` {}`,
			newHolderBrandRequest("1", `"dataHolderBrandId": "brand-new", "legalEntityId": "dh-1"`),
			newHolderBrandRequest("1", `"dataHolderBrandId": "brand-new", "lastUpdated": "2026-10-24T00:00:00Z"`),
			newHolderBrandRequest("1", `"dataHolderBrandId": "brand-new", "endpointDetail": {"api": "x"}`),
			newHolderBrandRequest("1", `"dataHolderBrandId": 5`),
		},
		changeHolderBrand: {
			``, `null`, `{"version": "1"}`,
			`{"version": 1, "status": "INACTIVE", ` + holderBrandFields + `}`,
			`{"version": 1, "dataHolderBrandId": "dh-1-b1", ` + holderBrandFields + `}`,
			`{"version": 1, "interimId": "dh-1-i9", ` + holderBrandFields + `}`,
			`{"version": 1, "industries": "banking"}`,
		},
	} {
		for _, body := range bodies {
			resp := post(t, ports, target, body)

			require.Equal(t, http.StatusBadRequest, resp.status, "%s body %.60s", target, body)
			assert.Contains(t, resp.body, `"type":"about:blank"`, "%s body %.60s", target, body)
		}
	}

	assert.Equal(t, int64(1), holderBrandsOf(t, ports, "dh-1").Version)
}

func TestAChangeOfADataHolderBrandAtTheCurrentVersionReplacesEachField(t *testing.T) {
	ports := seedForHolderBrands(t)

	// The change has no brand group and no product base URI, so these have no value after the change.
	resp := post(t, ports, changeHolderBrand, `{"version": 1, "brandName": "Renamed Bank",
		"industries": ["telco", "banking"], "logoUri": "https://holder.example/new.png",
		"endpointDetail": {
			"version": "2", "publicBaseUri": "https://api2.holder.example",
			"resourceBaseUri": "https://mtls2.holder.example",
			"infosecBaseUri": "https://id2.holder.example", "extensionBaseUri": "https://ext.holder.example",
			"websiteUri": "https://holder.example/new"
		},
		"jwksEndpoint": "https://id2.holder.example/jwks"}`)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)

	detail := holderBrandsOf(t, ports, "dh-1")
	assert.Equal(t, int64(2), detail.Version)
	assert.JSONEq(t, `{"dataHolderBrandId": "dh-1-b1", "status": "ACTIVE", "brandName": "Renamed Bank",
		"industries": ["telco", "banking"], "logoUri": "https://holder.example/new.png",
		"endpointDetail": {
			"version": "2", "publicBaseUri": "https://api2.holder.example",
			"resourceBaseUri": "https://mtls2.holder.example",
			"infosecBaseUri": "https://id2.holder.example", "extensionBaseUri": "https://ext.holder.example",
			"websiteUri": "https://holder.example/new"
		},
		"jwksEndpoint": "https://id2.holder.example/jwks"}`, string(detail.Brands[0]))

	history := get(t, ports.Readers, "/api/data-holders/dh-1/history?pageSize=1")
	assert.Contains(t, history.body, `"eventType":"DataHolderBrandDetailsChanged"`)
}

// Verifies: REG-1104
func TestAChangeOfABrandWithOnlyAnInterimIDUsesTheInterimIDInThePath(t *testing.T) {
	ports := seedForHolderBrands(t)

	resp := post(t, ports, "/api/data-holders/dh-1/brands/dh-1-i2/change",
		`{"version": 1, "brandName": "Renamed Products", "industries": ["telco"],
		"logoUri": "https://holder.example/products.png",
		"endpointDetail": {"publicBaseUri": "https://p.holder.example"}}`)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.JSONEq(t, `{"interimId": "dh-1-i2", "status": "ACTIVE", "brandName": "Renamed Products",
		"industries": ["telco"], "logoUri": "https://holder.example/products.png",
		"endpointDetail": {"publicBaseUri": "https://p.holder.example"}}`,
		string(holderBrandsOf(t, ports, "dh-1").Brands[1]))
}

func TestAChangeOfADataHolderBrandWithValuesThatTheParticipantModelRefusesGivesAnErrorAtEachField(t *testing.T) {
	ports := seedForHolderBrands(t)

	resp := post(t, ports, changeHolderBrand, `{"version": 1, "brandName": "Holder Bank", "industries": [],
		"logoUri": "https://holder.example/brand.png",
		"endpointDetail": {
			"version": "1", "publicBaseUri": "", "resourceBaseUri": "https://mtls.holder.example",
			"infosecBaseUri": "https://id.holder.example", "websiteUri": "https://holder.example"
		},
		"jwksEndpoint": "https://holder.example/jwks"}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/endpointDetail/publicBaseUri", "detail": "endpointDetail.publicBaseUri is required"},
			{"pointer": "#/industries", "detail": "industries is required"}
		]
	}`, resp.body)
	assert.Equal(t, int64(1), holderBrandsOf(t, ports, "dh-1").Version)
}

func TestAnAddOrAChangeOfADataHolderBrandGivesAProblemWithNoErrorMessageWhenItFails(t *testing.T) {
	ports := console.Ports{Operator: failingOperator{}}

	for target, body := range map[string]string{
		addHolderBrand:    newHolderBrandRequest("1", `"dataHolderBrandId": "brand-new"`),
		changeHolderBrand: `{"version": 1, ` + holderBrandFields + `}`,
	} {
		resp := post(t, ports, target, body)

		assert.Equal(t, http.StatusInternalServerError, resp.status, target)
		assert.NotContains(t, resp.body, "Mock Budgeting", target)
	}
}
