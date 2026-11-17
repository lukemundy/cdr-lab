package console_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func withoutHolderActions(t *testing.T, body string) string {
	t.Helper()

	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &detail))
	delete(detail, "actions")

	for _, b := range detail["dataHolderBrands"].([]any) {
		delete(b.(map[string]any), "actions")
		delete(b.(map[string]any), "serverCertificateActions")
	}

	rest, err := json.Marshal(detail)
	require.NoError(t, err)

	return string(rest)
}

type holderActionsBody struct {
	Version int64           `json:"version"`
	Status  string          `json:"status"`
	Actions json.RawMessage `json:"actions"`
	Brands  []struct {
		Status  string          `json:"status"`
		Actions json.RawMessage `json:"actions"`
	} `json:"dataHolderBrands"`
}

func holderActionsOf(t *testing.T, ports console.Ports, legalEntityID string) holderActionsBody {
	t.Helper()

	resp := get(t, ports.Readers, "/api/data-holders/"+legalEntityID)
	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var body holderActionsBody
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

	return body
}

// Verifies: REG-1133, REG-1134
func TestADataHolderDetailGivesThePermittedActionsWithTheirCascade(t *testing.T) {
	detail := holderActionsOf(t, seedForHolderBrands(t), "dh-1")

	assert.JSONEq(t, `[
		{"action": "remove", "status": "REMOVED", "irreversible": true, "changes": [
			{"dataHolderBrandId": "dh-1-b1", "brandName": "Holder Bank", "from": "ACTIVE", "to": "REMOVED"},
			{"interimId": "dh-1-i2", "brandName": "Holder Products", "from": "ACTIVE", "to": "REMOVED"}
		]}
	]`, string(detail.Actions))

	for _, b := range detail.Brands {
		assert.JSONEq(t, `[
			{"action": "deactivate", "status": "INACTIVE", "irreversible": false, "changes": []},
			{"action": "remove", "status": "REMOVED", "irreversible": true, "changes": []}
		]`, string(b.Actions))
	}
}

func TestARemovedDataHolderDetailGivesNoAction(t *testing.T) {
	assert.JSONEq(t, `[]`, string(holderActionsOf(t, seedForHolderBrands(t), "dh-2").Actions))
}

// Verifies: REG-1102
func TestEachDataHolderBrandActionChangesTheBrand(t *testing.T) {
	ports := seedForHolderBrands(t)
	queries := ports.Readers.(*readmodel.Queries)

	// The interim ID is in the path of a brand that has no dataHolderBrandId.
	for _, brand := range []struct {
		id    string
		index int
	}{{id: "dh-1-b1", index: 0}, {id: "dh-1-i2", index: 1}} {
		// Get Data Holder Brands Summary has no REMOVED brand.
		for _, step := range []struct {
			action, want string
			summary      domain.Status
		}{
			{action: "deactivate", want: "INACTIVE", summary: domain.Inactive},
			{action: "activate", want: "ACTIVE", summary: domain.Active},
			{action: "remove", want: "REMOVED", summary: ""},
		} {
			version := holderActionsOf(t, ports, "dh-1").Version
			resp := post(t, ports, "/api/data-holders/dh-1/brands/"+brand.id+"/"+step.action,
				fmt.Sprintf(`{"version": %d}`, version))

			require.Equal(t, http.StatusNoContent, resp.status, "%s %s: %s", brand.id, step.action, resp.body)
			assert.Empty(t, resp.body)
			assert.Equal(t, step.want, holderActionsOf(t, ports, "dh-1").Brands[brand.index].Status, step.action)

			brands, err := queries.DataHolderBrands(t.Context(), nil)
			require.NoError(t, err)

			summary := map[string]domain.Status{}
			for _, b := range brands {
				summary[b.Brand.ID()] = b.Brand.Status
			}

			assert.Equal(t, step.summary, summary[brand.id], "%s %s", brand.id, step.action)
		}

		assert.JSONEq(t, `[]`, string(holderActionsOf(t, ports, "dh-1").Brands[brand.index].Actions))
	}
}

// Verifies: REG-1101
func TestTheRemovalOfADataHolderRemovesItAndEachOfItsBrands(t *testing.T) {
	ports := seedForHolderBrands(t)
	queries := ports.Readers.(*readmodel.Queries)

	resp := post(t, ports, "/api/data-holders/dh-1/remove", `{"version": 1}`)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)

	detail := holderActionsOf(t, ports, "dh-1")
	assert.Equal(t, "REMOVED", detail.Status)
	assert.JSONEq(t, `[]`, string(detail.Actions))

	for _, b := range detail.Brands {
		assert.Equal(t, "REMOVED", b.Status)
		assert.JSONEq(t, `[]`, string(b.Actions))
	}

	statuses, err := queries.DataHolderStatuses(t.Context(), nil)
	require.NoError(t, err)
	assert.Contains(t, statuses, app.DataHolderStatusItem{LegalEntityID: "dh-1", Status: domain.DataHolderRemoved})

	brands, err := queries.DataHolderBrands(t.Context(), nil)
	require.NoError(t, err)
	assert.Empty(t, brands)
}

// Verifies: REG-1140
func TestADataHolderActionWithAVersionThatIsNotCurrentChangesNothing(t *testing.T) {
	ports := seedForHolderBrands(t)
	require.Equal(t, http.StatusNoContent,
		post(t, ports, "/api/data-holders/dh-1/brands/dh-1-b1/deactivate", `{"version": 1}`).status)
	before := holderActionsOf(t, ports, "dh-1")

	for _, target := range []string{
		"/api/data-holders/dh-1/remove",
		"/api/data-holders/dh-1/brands/dh-1-i2/deactivate",
		// The status rules also refuse this change, but the version comes first.
		"/api/data-holders/dh-1/brands/dh-1-b1/deactivate",
	} {
		resp := post(t, ports, target, `{"version": 1}`)

		require.Equal(t, http.StatusConflict, resp.status, target)
		assert.Contains(t, resp.body, "The participant changed after the version in the request.", target)
	}

	assert.Equal(t, before, holderActionsOf(t, ports, "dh-1"))
}

func TestADataHolderActionThatTheStatusRulesDoNotPermitGivesTheRule(t *testing.T) {
	ports := seedForHolderBrands(t)

	tests := map[string]string{
		"/api/data-holders/dh-2/remove":                  "cannot change from REMOVED to REMOVED",
		"/api/data-holders/dh-1/brands/dh-1-b1/activate": "cannot change from ACTIVE to ACTIVE",
		"/api/data-holders/dh-1/brands/dh-1-i2/activate": "cannot change from ACTIVE to ACTIVE",
	}
	for target, rule := range tests {
		resp := post(t, ports, target, `{"version": 1}`)

		require.Equal(t, http.StatusConflict, resp.status, target)
		assert.Contains(t, resp.body, rule, target)
	}

	assert.Equal(t, int64(1), holderActionsOf(t, ports, "dh-1").Version)
}

func TestADataHolderActionForARecordThatDoesNotExistGivesNotFoundWithNoIdentifier(t *testing.T) {
	ports := seedForHolderBrands(t)

	for _, target := range []string{
		"/api/data-holders/dh-unknown/remove",
		"/api/data-holders/dh-unknown/brands/dh-1-b1/deactivate",
		"/api/data-holders/dh-1/brands/brand-unknown/deactivate",
		// A data recipient is not a data holder.
		"/api/data-holders/dr-1/remove",
	} {
		resp := post(t, ports, target, `{"version": 1}`)

		require.Equal(t, http.StatusNotFound, resp.status, target)
		assert.NotContains(t, resp.body, "unknown", target)
		assert.NotContains(t, resp.body, "dr-1", target)
	}
}

func TestADataHolderActionRefusesABodyThatIsNotValid(t *testing.T) {
	ports := seedForHolderBrands(t)

	for _, target := range []string{
		"/api/data-holders/dh-1/remove",
		"/api/data-holders/dh-1/brands/dh-1-b1/deactivate",
	} {
		for _, body := range []string{
			``, `{}`, `{"version": 0}`, `{"version": "1"}`, `{"version": 1, "status": "REMOVED"}`, `{"version": 1} {}`,
		} {
			resp := post(t, ports, target, body)

			assert.Equal(t, http.StatusBadRequest, resp.status, "%s body %.40s", target, body)
		}
	}

	assert.Equal(t, int64(1), holderActionsOf(t, ports, "dh-1").Version)
}

func TestADataHolderActionGivesAProblemWithNoErrorMessageWhenTheChangeFails(t *testing.T) {
	resp := post(t, console.Ports{Operator: failingOperator{}}, "/api/data-holders/dh-1/remove", `{"version": 1}`)

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "Mock Budgeting")
}

// Verifies: REG-1137
func TestADataHolderDetailGivesThePermittedServerCertificateActionsOfEachBrand(t *testing.T) {
	resp := get(t, seedForHolderBrands(t).Readers, "/api/data-holders/dh-1")
	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var body struct {
		Brands []struct {
			Actions []string `json:"serverCertificateActions"`
		} `json:"dataHolderBrands"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

	require.Len(t, body.Brands, 2)
	assert.Equal(t, []string{"issue"}, body.Brands[0].Actions)
	assert.Equal(t, []string{}, body.Brands[1].Actions, "a brand with no dataHolderBrandId cannot get a certificate")
}
