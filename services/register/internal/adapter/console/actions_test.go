package console_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// withoutActions removes the version and each list of actions from the body of a data recipient detail.
func withoutActions(t *testing.T, body string) string {
	t.Helper()

	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &detail))
	delete(detail, "version")
	delete(detail, "actions")

	for _, b := range detail["dataRecipientBrands"].([]any) {
		delete(b.(map[string]any), "actions")

		for _, sp := range b.(map[string]any)["softwareProducts"].([]any) {
			delete(sp.(map[string]any), "actions")
			delete(sp.(map[string]any), "clientCertificateActions")
		}
	}

	rest, err := json.Marshal(detail)
	require.NoError(t, err)

	return string(rest)
}

type statusActionsBody struct {
	Version int64           `json:"version"`
	Status  string          `json:"status"`
	Actions json.RawMessage `json:"actions"`
	Brands  []struct {
		DataRecipientBrandID string          `json:"dataRecipientBrandId"`
		Status               string          `json:"status"`
		Actions              json.RawMessage `json:"actions"`
		SoftwareProducts     []struct {
			SoftwareProductID string          `json:"softwareProductId"`
			Status            string          `json:"status"`
			Actions           json.RawMessage `json:"actions"`
		} `json:"softwareProducts"`
	} `json:"dataRecipientBrands"`
}

func detailOf(t *testing.T, ports console.Ports, legalEntityID string) statusActionsBody {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/data-recipients/"+legalEntityID, nil)
	resp := do(t, ports, req)
	require.Equal(t, http.StatusOK, resp.status, resp.body)

	var body statusActionsBody
	require.NoError(t, json.Unmarshal([]byte(resp.body), &body))

	return body
}

func post(t *testing.T, ports console.Ports, target, body string) response {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	return do(t, ports, req)
}

// The seed data recipient has an ACTIVE and an INACTIVE software product.
func seedForActions(t *testing.T) console.Ports {
	t.Helper()

	store, queries := seedStore(t, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", "Alpha Pty Ltd", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active), product("sp-2", domain.Inactive)}),
	}})

	return console.Ports{Readers: queries, Operator: app.NewOperator(store)}
}

// Verifies: REG-1131, REG-1132, REG-1134
func TestADataRecipientDetailGivesTheVersionAndThePermittedActionsWithTheirCascade(t *testing.T) {
	detail := detailOf(t, seedForActions(t), "dr-1")

	assert.Equal(t, int64(1), detail.Version)
	assert.JSONEq(t, `[
		{"action": "suspend", "status": "SUSPENDED", "irreversible": false, "changes": [
			{"softwareProductId": "sp-1", "softwareProductName": "Product sp-1", "from": "ACTIVE", "to": "INACTIVE"}
		]},
		{"action": "revoke", "status": "REVOKED", "irreversible": true, "changes": [
			{"softwareProductId": "sp-1", "softwareProductName": "Product sp-1", "from": "ACTIVE", "to": "REMOVED"},
			{"softwareProductId": "sp-2", "softwareProductName": "Product sp-2", "from": "INACTIVE", "to": "REMOVED"}
		]},
		{"action": "surrender", "status": "SURRENDERED", "irreversible": true, "changes": [
			{"softwareProductId": "sp-1", "softwareProductName": "Product sp-1", "from": "ACTIVE", "to": "REMOVED"},
			{"softwareProductId": "sp-2", "softwareProductName": "Product sp-2", "from": "INACTIVE", "to": "REMOVED"}
		]}
	]`, string(detail.Actions))

	assert.JSONEq(t, `[
		{"action": "deactivate", "status": "INACTIVE", "irreversible": false, "changes": [
			{"softwareProductId": "sp-1", "softwareProductName": "Product sp-1", "from": "ACTIVE", "to": "INACTIVE"}
		]},
		{"action": "remove", "status": "REMOVED", "irreversible": true, "changes": [
			{"softwareProductId": "sp-1", "softwareProductName": "Product sp-1", "from": "ACTIVE", "to": "REMOVED"},
			{"softwareProductId": "sp-2", "softwareProductName": "Product sp-2", "from": "INACTIVE", "to": "REMOVED"}
		]}
	]`, string(detail.Brands[0].Actions))

	products := detail.Brands[0].SoftwareProducts
	assert.JSONEq(t, `[
		{"action": "deactivate", "status": "INACTIVE", "irreversible": false, "changes": []},
		{"action": "remove", "status": "REMOVED", "irreversible": true, "changes": []}
	]`, string(products[0].Actions))
	assert.JSONEq(t, `[
		{"action": "activate", "status": "ACTIVE", "irreversible": false, "changes": []},
		{"action": "remove", "status": "REMOVED", "irreversible": true, "changes": []}
	]`, string(products[1].Actions))
}

func TestASuspensionAtTheCurrentVersionChangesTheDataRecipientAndItsCascade(t *testing.T) {
	ports := seedForActions(t)

	resp := post(t, ports, "/api/data-recipients/dr-1/suspend", `{"version": 1}`)

	require.Equal(t, http.StatusNoContent, resp.status, resp.body)
	assert.Empty(t, resp.body)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))

	detail := detailOf(t, ports, "dr-1")
	assert.Equal(t, "SUSPENDED", detail.Status)
	assert.Equal(t, int64(3), detail.Version)
	assert.Equal(t, "INACTIVE", detail.Brands[0].SoftwareProducts[0].Status)
	assert.Contains(t, string(detail.Actions), `"action":"reinstate"`)

	statuses := ports.Readers.(*readmodel.Queries)
	recipients, err := statuses.DataRecipientStatuses(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []app.DataRecipientStatusItem{{LegalEntityID: "dr-1", Status: domain.DataRecipientSuspended}},
		recipients)
	products, err := statuses.SoftwareProductStatuses(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []app.SoftwareProductStatusItem{
		{SoftwareProductID: "sp-1", Status: domain.Inactive}, {SoftwareProductID: "sp-2", Status: domain.Inactive},
	}, products)
}

func TestEachSoftwareProductActionChangesTheSoftwareProduct(t *testing.T) {
	ports := seedForActions(t)

	for i, step := range []struct{ action, want string }{
		{action: "deactivate", want: "INACTIVE"},
		{action: "activate", want: "ACTIVE"},
		{action: "remove", want: "REMOVED"},
	} {
		version := detailOf(t, ports, "dr-1").Version
		resp := post(t, ports, "/api/data-recipients/dr-1/software-products/sp-1/"+step.action,
			fmt.Sprintf(`{"version": %d}`, version))

		require.Equal(t, http.StatusNoContent, resp.status, "step %d %s: %s", i, step.action, resp.body)
		assert.Equal(t, step.want, detailOf(t, ports, "dr-1").Brands[0].SoftwareProducts[0].Status, step.action)
	}
}

// Verifies: REG-1100
func TestEachBrandActionChangesTheBrandAndItsSoftwareProducts(t *testing.T) {
	ports := seedForActions(t)
	queries := ports.Readers.(*readmodel.Queries)

	for i, step := range []struct {
		action, want string
		products     []app.SoftwareProductStatusItem
	}{
		{action: "deactivate", want: "INACTIVE", products: []app.SoftwareProductStatusItem{
			{SoftwareProductID: "sp-1", Status: domain.Inactive}, {SoftwareProductID: "sp-2", Status: domain.Inactive},
		}},
		{action: "activate", want: "ACTIVE", products: []app.SoftwareProductStatusItem{
			{SoftwareProductID: "sp-1", Status: domain.Inactive}, {SoftwareProductID: "sp-2", Status: domain.Inactive},
		}},
		{action: "remove", want: "REMOVED", products: []app.SoftwareProductStatusItem{
			{SoftwareProductID: "sp-1", Status: domain.Removed}, {SoftwareProductID: "sp-2", Status: domain.Removed},
		}},
	} {
		version := detailOf(t, ports, "dr-1").Version
		resp := post(t, ports, "/api/data-recipients/dr-1/brands/dr-1-b1/"+step.action,
			fmt.Sprintf(`{"version": %d}`, version))

		require.Equal(t, http.StatusNoContent, resp.status, "step %d %s: %s", i, step.action, resp.body)
		assert.Empty(t, resp.body)
		assert.Equal(t, step.want, detailOf(t, ports, "dr-1").Brands[0].Status, step.action)

		recipients, err := queries.DataRecipients(t.Context())
		require.NoError(t, err)
		require.Len(t, recipients, 1)
		assert.Equal(t, domain.Status(step.want), recipients[0].DataRecipient.Brands[0].Status, step.action)

		products, err := queries.SoftwareProductStatuses(t.Context())
		require.NoError(t, err)
		assert.Equal(t, step.products, products, step.action)
	}

	assert.JSONEq(t, `[]`, string(detailOf(t, ports, "dr-1").Brands[0].Actions))
}

// Verifies: REG-1140
func TestAnActionWithAVersionThatIsNotCurrentChangesNothing(t *testing.T) {
	ports := seedForActions(t)
	require.Equal(t, http.StatusNoContent, post(t, ports, "/api/data-recipients/dr-1/suspend", `{"version": 1}`).status)
	before := detailOf(t, ports, "dr-1")

	for _, target := range []string{
		"/api/data-recipients/dr-1/revoke",
		// The status rules also refuse this change, but the version comes first.
		"/api/data-recipients/dr-1/suspend",
		"/api/data-recipients/dr-1/software-products/sp-2/remove",
		"/api/data-recipients/dr-1/brands/dr-1-b1/deactivate",
	} {
		resp := post(t, ports, target, `{"version": 1}`)

		require.Equal(t, http.StatusConflict, resp.status, target)
		assert.Contains(t, resp.body, "The participant changed after the version in the request.", target)
	}

	assert.Equal(t, before, detailOf(t, ports, "dr-1"))
}

func TestAnActionThatTheStatusRulesDoNotPermitGivesTheRule(t *testing.T) {
	ports := seedForActions(t)

	tests := map[string]string{
		"/api/data-recipients/dr-1/reinstate":                       "cannot change from ACTIVE to ACTIVE",
		"/api/data-recipients/dr-1/software-products/sp-1/activate": "cannot change from ACTIVE to ACTIVE",
		"/api/data-recipients/dr-1/brands/dr-1-b1/activate":         "cannot change from ACTIVE to ACTIVE",
	}
	for target, rule := range tests {
		resp := post(t, ports, target, `{"version": 1}`)

		require.Equal(t, http.StatusConflict, resp.status, target)
		assert.Contains(t, resp.body, rule, target)
	}

	assert.Equal(t, int64(1), detailOf(t, ports, "dr-1").Version)
}

func TestAnActionForARecordThatDoesNotExistGivesNotFoundWithNoIdentifier(t *testing.T) {
	ports := seedForActions(t)

	for _, target := range []string{
		"/api/data-recipients/dr-unknown/suspend",
		"/api/data-recipients/dr-unknown/software-products/sp-1/deactivate",
		"/api/data-recipients/dr-1/software-products/sp-unknown/deactivate",
		"/api/data-recipients/dr-unknown/brands/dr-1-b1/deactivate",
		"/api/data-recipients/dr-1/brands/brand-unknown/deactivate",
	} {
		resp := post(t, ports, target, `{"version": 1}`)

		require.Equal(t, http.StatusNotFound, resp.status, target)
		assert.NotContains(t, resp.body, "unknown", target)
	}
}

func TestAnActionRefusesABodyThatIsNotValid(t *testing.T) {
	ports := seedForActions(t)

	for _, target := range []string{
		"/api/data-recipients/dr-1/suspend",
		"/api/data-recipients/dr-1/brands/dr-1-b1/remove",
	} {
		for _, body := range []string{
			``, `{}`, `[]`, `null`, `{"version": null}`, `{"version": 0}`, `{"version": -1}`, `{"version": 1.5}`,
			`{"version": "1"}`, `{"version": 1, "status": "SUSPENDED"}`, `{"version": 1} {}`,
			`{"version": 1, "padding": "` + strings.Repeat("x", 2048) + `"}`,
		} {
			resp := post(t, ports, target, body)

			assert.Equal(t, http.StatusBadRequest, resp.status, "%s body %.40s", target, body)
		}
	}

	assert.Equal(t, int64(1), detailOf(t, ports, "dr-1").Version)
}

type failingOperator struct{}

func (failingOperator) Execute(context.Context, string, domain.DataRecipientCommand) error {
	return errors.New("a message that contains Mock Budgeting")
}

func (failingOperator) ExecuteAt(context.Context, string, int64, domain.DataRecipientCommand) error {
	return errors.New("a message that contains Mock Budgeting")
}

func (failingOperator) ExecuteDataHolder(context.Context, string, domain.DataHolderCommand) error {
	return errors.New("a message that contains Mock Budgeting")
}

func (failingOperator) ExecuteDataHolderAt(context.Context, string, int64, domain.DataHolderCommand) error {
	return errors.New("a message that contains Mock Budgeting")
}

func TestAnActionGivesAProblemWithNoErrorMessageWhenTheChangeFails(t *testing.T) {
	resp := post(t, console.Ports{Operator: failingOperator{}}, "/api/data-recipients/dr-1/suspend", `{"version": 1}`)

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "Mock Budgeting")
}
