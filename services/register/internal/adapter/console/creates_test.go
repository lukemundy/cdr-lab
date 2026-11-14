package console_test

import (
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
	createDataRecipient = "/api/data-recipients"
	newRecipientFields  = `"legalEntityName": "Gamma Pty Ltd", "accreditationNumber": "ADR-G",
		"accreditationLevel": "SPONSORED", "logoUri": "https://gamma.example/logo.png", "status": "ACTIVE"`
)

func newRecipientRequest(legalEntityID string) string {
	return `{"legalEntityId": "` + legalEntityID + `", ` + newRecipientFields + `}`
}

func seedForCreates(t *testing.T) console.Ports {
	t.Helper()

	store, queries := seedStore(t, domain.Participants{
		DataRecipients: []domain.DataRecipient{
			recipient(
				"dr-1",
				"Alpha Pty Ltd",
				domain.DataRecipientActive,
				[]domain.SoftwareProduct{product("sp-1", domain.Active)},
			),
		},
		DataHolders: []domain.DataHolder{dataHolder("dh-1", "Holder Pty Ltd", domain.DataHolderActive)},
	})

	return console.Ports{Readers: queries, Operator: app.NewOperator(store)}
}

func dataRecipientCount(t *testing.T, ports console.Ports) int {
	t.Helper()

	recipients, err := ports.Readers.(*readmodel.Queries).DataRecipients(t.Context())
	require.NoError(t, err)

	return len(recipients)
}

func TestACreateMakesADataRecipientWithNoBrandsAndGivesItsPath(t *testing.T) {
	ports := seedForCreates(t)

	resp := post(t, ports, createDataRecipient, newRecipientRequest("dr-new"))

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.Equal(t, "/api/data-recipients/dr-new", resp.header.Get("Location"))
	assert.Empty(t, resp.body)

	detail := get(t, ports.Readers, "/api/data-recipients/dr-new")
	require.Equal(t, http.StatusOK, detail.status)
	assert.JSONEq(t, `{
		"legalEntityId": "dr-new", "legalEntityName": "Gamma Pty Ltd", "accreditationNumber": "ADR-G",
		"accreditationLevel": "SPONSORED", "logoUri": "https://gamma.example/logo.png", "status": "ACTIVE",
		"dataRecipientBrands": []
	}`, withoutActions(t, withoutLastUpdated(t, detail.body)))
	assert.Equal(t, int64(1), detailOf(t, ports, "dr-new").Version)

	history := get(t, ports.Readers, "/api/data-recipients/dr-new/history")
	assert.Contains(t, history.body, `"eventType":"DataRecipientCreated"`)
	assert.Equal(t, 2, dataRecipientCount(t, ports))
}

func TestACreateGivesTheLocationOfAnIdentifierWithCharactersOfAPath(t *testing.T) {
	resp := post(t, seedForCreates(t), createDataRecipient, newRecipientRequest("dr/new?x"))

	require.Equal(t, http.StatusCreated, resp.status, resp.body)
	assert.Equal(t, "/api/data-recipients/dr%2Fnew%3Fx", resp.header.Get("Location"))
}

func TestACreateWithALegalEntityIDThatADifferentRecordHasGivesAnErrorAtTheLegalEntityID(t *testing.T) {
	// A data recipient, a data holder, a brand and a software product have these IDs.
	for _, id := range []string{"dr-1", "dh-1", "dr-1-b1", "sp-1"} {
		ports := seedForCreates(t)

		resp := post(t, ports, createDataRecipient, newRecipientRequest(id))

		require.Equal(t, http.StatusConflict, resp.status, id)
		assert.JSONEq(t, `{
			"type": "`+fieldErrorsType+`",
			"title": "Values in the request are not valid",
			"status": 409,
			"detail": "One or more values in the request are not valid.",
			"errors": [
				{"pointer": "#/legalEntityId", "detail": "legalEntityId is not unique: a different record has it"}
			]
		}`, resp.body, id)
		assert.Equal(t, 1, dataRecipientCount(t, ports), id)
	}

	assert.Equal(t, "Alpha Pty Ltd", legalEntityOf(t, seedForCreates(t)).LegalEntityName)
}

func TestACreateWithALogoURIThatIsNotAURIGivesAnErrorAtTheLogoURI(t *testing.T) {
	ports := seedForCreates(t)
	request := strings.Replace(newRecipientRequest("dr-new"), "https://gamma.example/logo.png", "foo", 1)

	resp := post(t, ports, createDataRecipient, request)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [{"pointer": "#/logoUri", "detail": "logoUri is not an https URI with a host"}]
	}`, resp.body)
	assert.Equal(t, 1, dataRecipientCount(t, ports))
}

func TestACreateWithValuesThatTheParticipantModelRefusesGivesAnErrorAtEachField(t *testing.T) {
	ports := seedForCreates(t)

	resp := post(t, ports, createDataRecipient, `{"legalEntityId": "`+strings.Repeat("x", 37)+`",
		"legalEntityName": "", "accreditationNumber": "ADR-G", "logoUri": "https://gamma.example/logo.png",
		"status": "INACTIVE"}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/legalEntityId", "detail": "legalEntityId is longer than 36 characters"},
			{"pointer": "#/legalEntityName", "detail": "legalEntityName is required"},
			{"pointer": "#/accreditationLevel",
			 "detail": "accreditationLevel \"\" is not UNRESTRICTED or SPONSORED"},
			{"pointer": "#/status", "detail": "status \"INACTIVE\" is not ACTIVE, SUSPENDED, REVOKED or SURRENDERED"}
		]
	}`, resp.body)
	assert.Equal(t, 1, dataRecipientCount(t, ports))
}

func TestACreateWithNoLegalEntityIDGivesAnErrorAtTheLegalEntityID(t *testing.T) {
	resp := post(t, seedForCreates(t), createDataRecipient, `{`+newRecipientFields+`}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)

	var body fieldProblem
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))
	require.Len(t, body.Errors, 1)
	assert.Equal(t, "#/legalEntityId", body.Errors[0].Pointer)
	assert.Equal(t, "legalEntityId is required", body.Errors[0].Detail)
}

func TestACreateRefusesABodyThatIsNotValid(t *testing.T) {
	ports := seedForCreates(t)

	for _, body := range []string{
		``, `[]`, `null`, `{"legalEntityId": 5}`, newRecipientRequest("dr-new") + ` {}`,
		`{"version": 1, "legalEntityId": "dr-new", ` + newRecipientFields + `}`,
		`{"legalEntityId": "dr-new", ` + newRecipientFields + `, "dataRecipientBrands": []}`,
		`{"legalEntityId": "dr-new", "logoUri": "` + strings.Repeat("x", 1<<20) + `"}`,
	} {
		resp := post(t, ports, createDataRecipient, body)

		require.Equal(t, http.StatusBadRequest, resp.status, "body %.60s", body)
		assert.Contains(t, resp.body, `"type":"about:blank"`, "body %.60s", body)
	}

	assert.Equal(t, 1, dataRecipientCount(t, ports))
}

func TestACreateGivesAProblemWithNoErrorMessageWhenTheCreateFails(t *testing.T) {
	resp := post(t, console.Ports{Operator: failingOperator{}}, createDataRecipient, newRecipientRequest("dr-new"))

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "Mock Budgeting")
}

const createDataHolder = "/api/data-holders"

func newHolderRequest(legalEntityID, status string) string {
	return `{"legalEntityId": "` + legalEntityID + `", "legalEntityName": "Delta Ltd",
		"logoUri": "https://delta.example/logo.png", "status": "` + status + `"}`
}

func dataHolderCount(t *testing.T, ports console.Ports) int {
	t.Helper()

	_, total, err := ports.Readers.(*readmodel.Queries).DataHolderPage(t.Context(), app.DataHolderPageQuery{
		Sort: app.SortByName, Page: 1, PageSize: 25,
	})
	require.NoError(t, err)

	return total
}

func TestACreateMakesADataHolderWithNoBrandsAndGivesItsPath(t *testing.T) {
	for _, status := range []string{"ACTIVE", "REMOVED"} {
		ports := seedForCreates(t)

		resp := post(t, ports, createDataHolder, newHolderRequest("dh/new", status))

		require.Equal(t, http.StatusCreated, resp.status, resp.body)
		assert.Equal(t, "/api/data-holders/dh%2Fnew", resp.header.Get("Location"))
		assert.Empty(t, resp.body)

		detail := get(t, ports.Readers, "/api/data-holders/dh%2Fnew")
		require.Equal(t, http.StatusOK, detail.status)
		assert.JSONEq(t, `{
			"legalEntityId": "dh/new", "legalEntityName": "Delta Ltd", "logoUri": "https://delta.example/logo.png",
			"status": "`+status+`", "version": 1, "dataHolderBrands": []
		}`, withoutHolderActions(t, withoutLastUpdated(t, detail.body)))

		history := get(t, ports.Readers, "/api/data-holders/dh%2Fnew/history")
		assert.Contains(t, history.body, `"eventType":"DataHolderCreated"`)
		assert.Equal(t, 2, dataHolderCount(t, ports))
	}
}

func TestACreateOfADataHolderWithALegalEntityIDThatADifferentRecordHasGivesAnErrorAtTheLegalEntityID(t *testing.T) {
	// A data holder, a data recipient, a brand and a software product have these IDs.
	for _, id := range []string{"dh-1", "dr-1", "dr-1-b1", "sp-1"} {
		ports := seedForCreates(t)

		resp := post(t, ports, createDataHolder, newHolderRequest(id, "ACTIVE"))

		require.Equal(t, http.StatusConflict, resp.status, id)
		assert.JSONEq(t, `{
			"type": "`+fieldErrorsType+`",
			"title": "Values in the request are not valid",
			"status": 409,
			"detail": "One or more values in the request are not valid.",
			"errors": [
				{"pointer": "#/legalEntityId", "detail": "legalEntityId is not unique: a different record has it"}
			]
		}`, resp.body, id)
		assert.Equal(t, 1, dataHolderCount(t, ports), id)
	}
}

func TestACreateOfADataHolderWithValuesThatTheParticipantModelRefusesGivesAnErrorAtEachField(t *testing.T) {
	ports := seedForCreates(t)

	resp := post(t, ports, createDataHolder, `{"legalEntityId": "", "legalEntityName": "", "status": "INACTIVE"}`)

	require.Equal(t, http.StatusBadRequest, resp.status, resp.body)
	assert.JSONEq(t, `{
		"type": "`+fieldErrorsType+`",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [
			{"pointer": "#/legalEntityId", "detail": "legalEntityId is required"},
			{"pointer": "#/legalEntityName", "detail": "legalEntityName is required"},
			{"pointer": "#/logoUri", "detail": "logoUri is required"},
			{"pointer": "#/status", "detail": "status \"INACTIVE\" is not ACTIVE or REMOVED"}
		]
	}`, resp.body)
	assert.Equal(t, 1, dataHolderCount(t, ports))
}

func TestACreateOfADataHolderRefusesABodyThatIsNotValid(t *testing.T) {
	ports := seedForCreates(t)

	for _, body := range []string{
		``, `[]`, `null`, `{"legalEntityId": 5}`, newHolderRequest("dh-new", "ACTIVE") + ` {}`,
		`{"version": 1, "legalEntityId": "dh-new"}`,
		`{"legalEntityId": "dh-new", "dataHolderBrands": []}`,
		`{"legalEntityId": "dh-new", "accreditationNumber": "ADR-1"}`,
	} {
		resp := post(t, ports, createDataHolder, body)

		require.Equal(t, http.StatusBadRequest, resp.status, "body %.60s", body)
		assert.Contains(t, resp.body, `"type":"about:blank"`, "body %.60s", body)
	}

	assert.Equal(t, 1, dataHolderCount(t, ports))
}

func TestACreateOfADataHolderGivesAProblemWithNoErrorMessageWhenTheCreateFails(t *testing.T) {
	resp := post(t, console.Ports{Operator: failingOperator{}}, createDataHolder, newHolderRequest("dh-new", "ACTIVE"))

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "Mock Budgeting")
}
