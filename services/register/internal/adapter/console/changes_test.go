package console_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const (
	changeDetails     = "/api/data-recipients/dr-1/change"
	fieldErrorsType   = "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002"
	newDetailsRequest = `{"version": 1, "legalEntityName": "Alpha Group Pty Ltd", "accreditationNumber": "ADR-9",
		"accreditationLevel": "SPONSORED", "logoUri": "https://recipient.example/new.png"}`
)

type legalEntityDetails struct {
	Version             int64  `json:"version"`
	LegalEntityName     string `json:"legalEntityName"`
	AccreditationNumber string `json:"accreditationNumber"`
	AccreditationLevel  string `json:"accreditationLevel"`
	LogoURI             string `json:"logoUri"`
}

func legalEntityOf(t *testing.T, ports console.Ports) legalEntityDetails {
	t.Helper()

	var detail legalEntityDetails
	require.NoError(t, json.Unmarshal([]byte(get(t, ports.Readers, "/api/data-recipients/dr-1").body), &detail))

	return detail
}

// Verifies: REG-1301, REG-1322, REG-1323
func TestAChangeOfTheDetailsAtTheCurrentVersionReplacesEachField(t *testing.T) {
	ports := seedForActions(t)

	resp := post(t, ports, changeDetails, newDetailsRequest)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)
	assert.Equal(t, legalEntityDetails{
		Version: 2, LegalEntityName: "Alpha Group Pty Ltd", AccreditationNumber: "ADR-9",
		AccreditationLevel: "SPONSORED", LogoURI: "https://recipient.example/new.png",
	}, legalEntityOf(t, ports))

	history := get(t, ports.Readers, "/api/data-recipients/dr-1/history?pageSize=1")
	assert.Contains(t, history.body, `"eventType":"DataRecipientDetailsChanged"`)

	recipients, err := ports.Readers.(*readmodel.Queries).DataRecipients(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "Alpha Group Pty Ltd", recipients[0].DataRecipient.LegalEntityName)
}

type fieldProblem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
	Errors []struct {
		Pointer string `json:"pointer"`
		Detail  string `json:"detail"`
	} `json:"errors"`
}

// Verifies: REG-1328
func TestAChangeWithValuesThatTheParticipantModelRefusesGivesAnErrorAtEachField(t *testing.T) {
	ports := seedForActions(t)

	resp := post(t, ports, changeDetails, `{"version": 1, "legalEntityName": "`+strings.Repeat("x", 201)+
		`", "accreditationNumber": "", "accreditationLevel": "OTHER"}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.Equal(t, "application/problem+json", resp.header.Get("Content-Type"))
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/legalEntityName", "detail": "legalEntityName is longer than 200 characters"},
			{"pointer": "#/accreditationNumber", "detail": "accreditationNumber is required"},
			{"pointer": "#/accreditationLevel",
			 "detail": "accreditationLevel \"OTHER\" is not UNRESTRICTED or SPONSORED"},
			{"pointer": "#/logoUri", "detail": "logoUri is required"}
		]
	}`, resp.body)
	assert.Equal(t, int64(1), legalEntityOf(t, ports).Version)
}

// Verifies: REG-1330
func TestAChangeWithAVersionThatIsNotCurrentChangesNothingAndHasNoFieldErrors(t *testing.T) {
	ports := seedForActions(t)
	require.Equal(t, http.StatusNoContent, post(t, ports, changeDetails, newDetailsRequest).status)

	// The values are not valid too, but the version comes first.
	resp := post(t, ports, changeDetails, `{"version": 1, "legalEntityName": ""}`)

	require.Equal(t, http.StatusConflict, resp.status)

	var body fieldProblem
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	assert.Equal(t, "about:blank", body.Type)
	assert.Equal(t, "The participant changed after the version in the request. The Register changed nothing.",
		body.Detail)
	assert.Nil(t, body.Errors)
	assert.Equal(t, int64(2), legalEntityOf(t, ports).Version)
}

func TestAChangeOfADataRecipientThatDoesNotExistGivesNotFoundWithNoIdentifier(t *testing.T) {
	resp := post(t, seedForActions(t), "/api/data-recipients/dr-unknown/change", newDetailsRequest)

	require.Equal(t, http.StatusNotFound, resp.status)
	assert.NotContains(t, resp.body, "unknown")
}

// Verifies: REG-1327
func TestAChangeRefusesABodyThatIsNotValid(t *testing.T) {
	ports := seedForActions(t)

	for _, body := range []string{
		``, `[]`, `null`, `{}`, `{"version": 0, "legalEntityName": "Alpha"}`, `{"version": "1"}`,
		`{"version": 1, "status": "SUSPENDED"}`, `{"version": 1, "legalEntityId": "dr-2"}`,
		`{"version": 1, "legalEntityName": 5}`, `{"version": 1} {}`,
	} {
		resp := post(t, ports, changeDetails, body)

		require.Equal(t, http.StatusBadRequest, resp.status, "body %.40s", body)
		assert.Contains(t, resp.body, `"type":"about:blank"`, "body %.40s", body)
	}

	assert.Equal(t, int64(1), legalEntityOf(t, ports).Version)
}

type erringOperator struct{ err error }

func (o erringOperator) Execute(context.Context, string, domain.DataRecipientCommand) error {
	return o.err
}

func (o erringOperator) ExecuteAt(context.Context, string, int64, domain.DataRecipientCommand) error {
	return o.err
}

func (o erringOperator) ExecuteDataHolder(context.Context, string, domain.DataHolderCommand) error {
	return o.err
}

func (o erringOperator) ExecuteDataHolderAt(context.Context, string, int64, domain.DataHolderCommand) error {
	return o.err
}

// Verifies: REG-1331
func TestAChangeGivesTheErrorsOfOtherRecordsInTheDetail(t *testing.T) {
	_, err := domain.HandleDataRecipient(
		domain.ApplyDataRecipient(domain.DataRecipientState{}, domain.DataRecipientImported{
			DataRecipient: recipient("dr-2", "Beta", domain.DataRecipientActive),
		}),
		domain.ChangeDataRecipientDetails{Details: domain.DataRecipientDetails{}},
	)
	require.Error(t, err)

	resp := post(t, console.Ports{Operator: erringOperator{err}}, changeDetails, newDetailsRequest)

	require.Equal(t, http.StatusBadRequest, resp.status)

	var body fieldProblem
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	assert.Equal(t, "about:blank", body.Type)
	assert.Contains(t, body.Detail, `data recipient "dr-2": legalEntityName is required`)
}

func TestAChangeGivesAProblemWithNoErrorMessageWhenTheChangeFails(t *testing.T) {
	resp := post(t, console.Ports{Operator: failingOperator{}}, changeDetails, newDetailsRequest)

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "Mock Budgeting")
}

const (
	changeBrand    = "/api/data-recipients/dr-1/brands/dr-1-b1/change"
	changeProduct  = "/api/data-recipients/dr-1/software-products/sp-1/change"
	newBrandFields = `"brandName": "Alpha Budgeting", "logoUri": "https://recipient.example/alpha.png"`
	// The request has no sectorIdentifierUri, tosUri or policyUri, so the change removes them.
	newProductFields = `"softwareProductName": "Alpha App", "softwareProductDescription": "A new description",
		"logoUri": "https://alpha.example/logo.png", "clientUri": "https://alpha.example",
		"redirectUris": ["https://alpha.example/cb1", "https://alpha.example/cb2"],
		"jwksUri": "https://alpha.example/jwks", "revocationUri": "https://alpha.example/revoke",
		"recipientBaseUri": "https://cdr.alpha.example", "scopes": ["openid", "profile"]`
)

func seedWithOptionalURIs(t *testing.T) console.Ports {
	t.Helper()

	sp := product("sp-1", domain.Active)
	sp.SectorIdentifierURI, sp.TOSURI, sp.PolicyURI = "https://recipient.example/sector",
		"https://recipient.example/tos", "https://recipient.example/policy"
	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive, []domain.SoftwareProduct{sp}),
	}})

	return console.Ports{Readers: queries, Operator: app.NewOperator(store)}
}

func firstBrandOf(t *testing.T, ports console.Ports) map[string]any {
	t.Helper()

	var detail struct {
		Brands []map[string]any `json:"dataRecipientBrands"`
	}
	require.NoError(t, json.Unmarshal([]byte(get(t, ports.Readers, "/api/data-recipients/dr-1").body), &detail))

	return detail.Brands[0]
}

// Verifies: REG-1303
func TestAChangeOfABrandAtTheCurrentVersionReplacesEachFieldOfTheBrand(t *testing.T) {
	ports := seedForActions(t)

	resp := post(t, ports, changeBrand, `{"version": 1, `+newBrandFields+`}`)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	brand := firstBrandOf(t, ports)
	assert.Equal(t, "Alpha Budgeting", brand["brandName"])
	assert.Equal(t, "https://recipient.example/alpha.png", brand["logoUri"])
	assert.Equal(t, "ACTIVE", brand["status"])
	assert.Len(t, brand["softwareProducts"], 2)
	assert.Equal(t, int64(2), legalEntityOf(t, ports).Version)

	history := get(t, ports.Readers, "/api/data-recipients/dr-1/history?pageSize=1")
	assert.Contains(t, history.body, `"eventType":"DataRecipientBrandDetailsChanged"`)
}

// Verifies: REG-1305, REG-1324
func TestAChangeOfASoftwareProductReplacesEachFieldAndRemovesEachOptionalFieldThatIsNotInTheBody(t *testing.T) {
	ports := seedWithOptionalURIs(t)

	resp := post(t, ports, changeProduct, `{"version": 1, `+newProductFields+`}`)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)

	got := firstBrandOf(t, ports)["softwareProducts"].([]any)[0].(map[string]any)
	delete(got, "actions")
	delete(got, "clientCertificateActions")
	assert.Equal(t, map[string]any{
		"softwareProductId": "sp-1", "softwareProductName": "Alpha App",
		"softwareProductDescription": "A new description", "logoUri": "https://alpha.example/logo.png",
		"status": "ACTIVE", "clientUri": "https://alpha.example",
		"redirectUris": []any{"https://alpha.example/cb1", "https://alpha.example/cb2"},
		"jwksUri":      "https://alpha.example/jwks", "revocationUri": "https://alpha.example/revoke",
		"recipientBaseUri": "https://cdr.alpha.example", "scopes": []any{"openid", "profile"},
		"clientCertificateThumbprints": []any{},
	}, got)

	history := get(t, ports.Readers, "/api/data-recipients/dr-1/history?pageSize=1")
	assert.Contains(t, history.body, `"eventType":"SoftwareProductDetailsChanged"`)
}

func TestAChangeOfASoftwareProductWithEmptyListsGivesAnErrorAtTheFieldOfEachList(t *testing.T) {
	ports := seedForActions(t)

	resp := post(t, ports, changeProduct, `{"version": 1, "softwareProductName": "Alpha App",
		"softwareProductDescription": "A product", "logoUri": "https://alpha.example/logo.png",
		"clientUri": "https://alpha.example", "redirectUris": [], "jwksUri": "https://alpha.example/jwks",
		"revocationUri": "https://alpha.example/revoke", "recipientBaseUri": "https://cdr.alpha.example",
		"scopes": []}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/redirectUris", "detail": "redirectUris is required"},
			{"pointer": "#/scopes", "detail": "scopes is required"}
		]
	}`, resp.body)
	assert.Equal(t, int64(1), legalEntityOf(t, ports).Version)
}

// Verifies: PROB-0002, PROB-0003
func TestAChangeOfABrandWithValuesThatTheParticipantModelRefusesGivesAnErrorAtEachField(t *testing.T) {
	ports := seedForActions(t)

	resp := post(t, ports, changeBrand, `{"version": 1, "brandName": ""}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)

	var body fieldProblem
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	assert.Equal(t, fieldErrorsType, body.Type)
	assert.Len(t, body.Errors, 2)
	assert.Equal(t, "#/brandName", body.Errors[0].Pointer)
	assert.Equal(t, "#/logoUri", body.Errors[1].Pointer)
}

// Verifies: REG-1330
func TestAChangeOfABrandOrASoftwareProductWithAStaleVersionChangesNothing(t *testing.T) {
	for _, c := range []struct{ target, fields string }{
		{changeBrand, newBrandFields},
		{changeProduct, newProductFields},
	} {
		ports := seedForActions(t)
		require.Equal(t, http.StatusNoContent, post(t, ports, changeDetails, newDetailsRequest).status)

		resp := post(t, ports, c.target, `{"version": 1, `+c.fields+`}`)

		require.Equal(t, http.StatusConflict, resp.status, c.target)
		assert.Contains(t, resp.body, "The participant changed after the version in the request.", c.target)
		assert.NotContains(t, resp.body, `"errors"`, c.target)
		assert.Equal(t, int64(2), legalEntityOf(t, ports).Version, c.target)
	}
}

func TestAChangeOfABrandOrASoftwareProductThatDoesNotExistGivesNotFoundWithNoIdentifier(t *testing.T) {
	ports := seedForActions(t)

	for _, c := range []struct{ target, fields string }{
		{"/api/data-recipients/dr-1/brands/unknown-b/change", newBrandFields},
		{"/api/data-recipients/unknown-dr/brands/dr-1-b1/change", newBrandFields},
		{"/api/data-recipients/dr-1/software-products/unknown-sp/change", newProductFields},
		{"/api/data-recipients/unknown-dr/software-products/sp-1/change", newProductFields},
	} {
		resp := post(t, ports, c.target, `{"version": 1, `+c.fields+`}`)

		require.Equal(t, http.StatusNotFound, resp.status, c.target)
		assert.NotContains(t, resp.body, "unknown", c.target)
	}
}

// Verifies: REG-1327
func TestAChangeOfABrandOrASoftwareProductRefusesAFieldThatTheChangeDoesNotAccept(t *testing.T) {
	ports := seedForActions(t)

	for _, c := range []struct{ target, body string }{
		{changeBrand, `{"version": 1, ` + newBrandFields + `, "status": "INACTIVE"}`},
		{changeBrand, `{"version": 1, ` + newBrandFields + `, "dataRecipientBrandId": "b-2"}`},
		{changeBrand, `{"version": 1, ` + newBrandFields + `, "softwareProducts": []}`},
		{changeBrand, `{` + newBrandFields + `}`},
		{changeProduct, `{"version": 1, ` + newProductFields + `, "status": "INACTIVE"}`},
		{changeProduct, `{"version": 1, ` + newProductFields + `, "softwareProductId": "sp-9"}`},
		{changeProduct, `{"version": 1, ` + newProductFields + `, "clientCertificates": []}`},
		{changeProduct, `{"version": 1, "redirectUris": "https://alpha.example/cb"}`},
	} {
		resp := post(t, ports, c.target, c.body)

		require.Equal(t, http.StatusBadRequest, resp.status, "body %.60s", c.body)
		assert.Contains(t, resp.body, `"type":"about:blank"`, "body %.60s", c.body)
	}

	assert.Equal(t, int64(1), legalEntityOf(t, ports).Version)
}

const (
	changeHolderDetails     = "/api/data-holders/dh-1/change"
	newHolderDetailsRequest = `{"version": 1, "legalEntityName": "Holder Group Ltd",
		"logoUri": "https://holder.example/new.png"}`
)

type holderDetails struct {
	Version         int64  `json:"version"`
	LegalEntityName string `json:"legalEntityName"`
	LogoURI         string `json:"logoUri"`
	Status          string `json:"status"`
}

func holderDetailsOf(t *testing.T, ports console.Ports) holderDetails {
	t.Helper()

	var detail holderDetails
	require.NoError(t, json.Unmarshal([]byte(get(t, ports.Readers, "/api/data-holders/dh-1").body), &detail))

	return detail
}

// Verifies: REG-1307
func TestAChangeOfTheDetailsOfADataHolderAtTheCurrentVersionReplacesEachField(t *testing.T) {
	ports := seedForCreates(t)

	resp := post(t, ports, changeHolderDetails, newHolderDetailsRequest)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)
	assert.Equal(t, holderDetails{
		Version: 2, LegalEntityName: "Holder Group Ltd", LogoURI: "https://holder.example/new.png", Status: "ACTIVE",
	}, holderDetailsOf(t, ports))

	history := get(t, ports.Readers, "/api/data-holders/dh-1/history?pageSize=1")
	assert.Contains(t, history.body, `"eventType":"DataHolderDetailsChanged"`)
}

// Verifies: REG-1328
func TestAChangeOfADataHolderWithValuesThatTheParticipantModelRefusesGivesAnErrorAtEachField(t *testing.T) {
	ports := seedForCreates(t)

	resp := post(t, ports, changeHolderDetails, `{"version": 1, "legalEntityName": "`+strings.Repeat("x", 37)+`"}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/legalEntityName", "detail": "legalEntityName is longer than 36 characters"},
			{"pointer": "#/logoUri", "detail": "logoUri is required"}
		]
	}`, resp.body)
	assert.Equal(t, int64(1), holderDetailsOf(t, ports).Version)
}

// Verifies: REG-1330
func TestAChangeOfADataHolderWithAVersionThatIsNotCurrentChangesNothingAndHasNoFieldErrors(t *testing.T) {
	ports := seedForCreates(t)
	require.Equal(t, http.StatusNoContent, post(t, ports, changeHolderDetails, newHolderDetailsRequest).status)

	resp := post(t, ports, changeHolderDetails, `{"version": 1, "legalEntityName": ""}`)

	require.Equal(t, http.StatusConflict, resp.status)

	var body fieldProblem
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	assert.Equal(t, "about:blank", body.Type)
	assert.Equal(t, "The participant changed after the version in the request. The Register changed nothing.",
		body.Detail)
	assert.Nil(t, body.Errors)
	assert.Equal(t, int64(2), holderDetailsOf(t, ports).Version)
}

func TestAChangeOfADataHolderThatDoesNotExistGivesNotFoundWithNoIdentifier(t *testing.T) {
	ports := seedForCreates(t)

	// dr-1 is a data recipient, not a data holder.
	for _, id := range []string{"dh-unknown", "dr-1"} {
		resp := post(t, ports, "/api/data-holders/"+id+"/change", newHolderDetailsRequest)

		require.Equal(t, http.StatusNotFound, resp.status, id)
		assert.Contains(t, resp.body, "The Register has no data holder with this legal entity ID.", id)
		assert.NotContains(t, resp.body, id, id)
	}
}

// Verifies: REG-1327
func TestAChangeOfADataHolderRefusesABodyThatIsNotValid(t *testing.T) {
	ports := seedForCreates(t)

	for _, body := range []string{
		``, `[]`, `null`, `{}`, `{"version": 0, "legalEntityName": "Holder"}`, `{"version": "1"}`,
		`{"version": 1, "status": "REMOVED"}`, `{"version": 1, "legalEntityId": "dh-2"}`,
		`{"version": 1, "dataHolderBrands": []}`, `{"version": 1, "legalEntityName": 5}`, `{"version": 1} {}`,
	} {
		resp := post(t, ports, changeHolderDetails, body)

		require.Equal(t, http.StatusBadRequest, resp.status, "body %.40s", body)
		assert.Contains(t, resp.body, `"type":"about:blank"`, "body %.40s", body)
	}

	assert.Equal(t, int64(1), holderDetailsOf(t, ports).Version)
}

func TestAChangeOfADataHolderGivesAProblemWithNoErrorMessageWhenTheChangeFails(t *testing.T) {
	resp := post(t, console.Ports{Operator: failingOperator{}}, changeHolderDetails, newHolderDetailsRequest)

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "Mock Budgeting")
}
