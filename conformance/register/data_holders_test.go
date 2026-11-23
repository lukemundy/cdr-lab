//go:build conformance

package register

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// The data holders and the brands of services/register/seed/example.json, except the preset brands.
const (
	wattleBankBrand      = "d7385786-d8b1-4d4b-bb80-b21616df1a4d"
	mockEnergy           = "293b45de-2a40-4e1b-ac77-324dacccb2c8"
	mockLending          = "1d344684-ca92-463d-af1e-4239ed438c18" // INACTIVE
	mockMutual           = "d9530e60-cdc4-408c-a6f4-aba39e4158b6" // interim ID only
	brightEnergyBrand    = "fa92bb4f-0914-4631-91af-98a6d34e23d4"
	brightEnergyInterim  = "7e1b355e-af51-4c7e-8525-645c2a9da175"
	brightEnergyBusiness = "71badab5-29d3-435c-a336-66c4b969e423" // interim ID only
	quietMobile          = "51afeac6-c330-4d7e-891f-198e81524257" // INACTIVE

	mockHoldings    = "a81975c6-1ac9-4fe4-88fc-68f50efd27cb"
	formerBank      = "91bbf006-ca02-47b6-99c3-ddc63541e1c6"
	brightEnergy    = "48cc4432-47f8-43b4-98ca-1a5ec4698d28"
	formerUtilities = "51a7978a-53e3-4a1a-97d8-74d7ee2d0014"
	quietTelecoms   = "476cd829-6efb-4c0b-b450-d8ef8eb467c3"
)

// A preset brand is an ACTIVE banking brand of an ACTIVE data holder.
type presetBrand struct {
	legalEntityID, brandID string
}

var presetBrands = map[string]presetBrand{
	"wattle-bank":  {"dd4d6bc6-22f4-4f61-aae9-9580382718c8", wattleBankBrand},
	"redgum-bank":  {"830186b8-6060-4773-8b59-e8567a13df01", "4967dd80-2cd2-43c3-8b13-f6faafb684ad"},
	"harbour-bank": {"00b12e81-207f-4fb7-b01f-d0c50cb1a06c", "366f7b1c-93c7-4b84-9c92-ab2456614095"},
	"kestrel-bank": {"0ed9945f-022e-4dca-9d05-ce327566e3de", "1ee54312-624b-4679-bf33-cf4af113f934"},
}

// registerPresets gives the preset brands in the Register, which are different in compose and in the cluster job.
func registerPresets(t *testing.T) []presetBrand {
	t.Helper()

	var presets []presetBrand

	for name := range strings.SplitSeq(labtest.Env(t, "REGISTER_PRESET_BRANDS"), ",") {
		preset, ok := presetBrands[name]
		require.True(t, ok, "REGISTER_PRESET_BRANDS: %q is not a preset brand", name)

		presets = append(presets, preset)
	}

	return presets
}

type dataHolderStatus struct {
	LegalEntityID string `json:"legalEntityId"`
	Status        string `json:"status"`
}

// Verifies: REG-0310, REG-0338
func TestDataHolderStatusesGiveTheDataHoldersOfTheIndustryInAllStatuses(t *testing.T) {
	var presets []dataHolderStatus
	for _, preset := range registerPresets(t) {
		presets = append(presets, dataHolderStatus{preset.legalEntityID, "ACTIVE"})
	}

	var (
		mock    = dataHolderStatus{mockHoldings, "ACTIVE"}
		bank    = dataHolderStatus{formerBank, "REMOVED"}
		energy  = dataHolderStatus{brightEnergy, "ACTIVE"}
		utility = dataHolderStatus{formerUtilities, "REMOVED"}
		telco   = dataHolderStatus{quietTelecoms, "ACTIVE"}
	)

	tests := map[string][]dataHolderStatus{
		"all":              append(slices.Clone(presets), mock, bank, energy, utility, telco),
		"banking":          append(slices.Clone(presets), mock, bank),
		"energy":           {mock, energy, utility},
		"non-bank-lending": {mock, bank},
		"telco":            {mock, utility, telco},
	}
	for industry, want := range tests {
		t.Run(industry, func(t *testing.T) {
			resp := call(t, "/cdr-register/v1/"+industry+"/data-holders/status", xv("2"))

			require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
			assert.Equal(t, "2", resp.Header.Get("x-v"))

			var body struct {
				Data []dataHolderStatus `json:"data"`
			}
			require.NoError(t, json.Unmarshal(resp.body, &body))
			assert.ElementsMatch(t, want, body.Data)
		})
	}
}

func TestDataHolderStatusesOfDifferentIndustriesHaveDifferentETags(t *testing.T) {
	banking := call(t, "/cdr-register/v1/banking/data-holders/status", xv("2"))
	energy := call(t, "/cdr-register/v1/energy/data-holders/status", xv("2"))

	assert.NotEqual(t, banking.Header.Get("ETag"), energy.Header.Get("ETag"))
}

type brandSummary struct {
	DataHolderBrandID string `json:"dataHolderBrandId"`
	InterimID         string `json:"interimId"`
}

func brandsSummary(t *testing.T, industry string) []brandSummary {
	t.Helper()
	resp := call(t, "/cdr-register/v1/"+industry+"/data-holders/brands/summary", xv("2"))
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
	assert.Equal(t, "2", resp.Header.Get("x-v"))

	var body struct {
		Data []brandSummary `json:"data"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &body))

	return body.Data
}

// Verifies: REG-0341, REG-0342
func TestTheBrandsSummaryGivesTheBrandsOfTheIndustryThatAreNotRemoved(t *testing.T) {
	var presets []brandSummary
	for _, preset := range registerPresets(t) {
		presets = append(presets, brandSummary{DataHolderBrandID: preset.brandID})
	}

	var (
		energy    = brandSummary{DataHolderBrandID: mockEnergy}
		lending   = brandSummary{DataHolderBrandID: mockLending}
		mutual    = brandSummary{InterimID: mockMutual}
		bright    = brandSummary{DataHolderBrandID: brightEnergyBrand, InterimID: brightEnergyInterim}
		brightBiz = brandSummary{InterimID: brightEnergyBusiness}
		quiet     = brandSummary{DataHolderBrandID: quietMobile}
	)

	tests := map[string][]brandSummary{
		"all":              append(slices.Clone(presets), energy, lending, mutual, bright, brightBiz, quiet),
		"banking":          append(slices.Clone(presets), mutual),
		"energy":           {energy, bright, brightBiz},
		"non-bank-lending": {lending},
		"telco":            {energy, quiet},
	}
	for industry, want := range tests {
		t.Run(industry, func(t *testing.T) {
			assert.ElementsMatch(t, want, brandsSummary(t, industry))
		})
	}
}
