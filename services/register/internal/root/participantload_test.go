package root

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/seedfile"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const participantLoads = "/v1/participant-loads"

type participantLoad struct {
	Added    []string `json:"added"`
	Existing []string `json:"existing"`
}

func loadParticipants(t *testing.T, handlers map[int]http.Handler, body string) participantLoad {
	t.Helper()

	rec := send(t, handlers, http.MethodPost, participantLoads, body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var load participantLoad
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &load))

	return load
}

func newParticipantsFile(holderBrand string) string {
	return `{
		"dataHolders": [` + newDataHolderJSON(newDataHolder, holderBrand) + `],
		"dataRecipients": [` + newRecipientJSON() + `]
	}`
}

// seedIDs gives the IDs of the participants in the sequence of the file.
func seedIDs(p domain.Participants) []string {
	var ids []string

	for _, dh := range p.DataHolders {
		ids = append(ids, dh.LegalEntityID)
		for _, b := range dh.Brands {
			ids = append(ids, b.ID())
		}
	}

	for _, dr := range p.DataRecipients {
		ids = append(ids, dr.IDs()...)
	}

	return ids
}

// Verifies: REG-1532, REG-1533, REG-1540
func TestALoadOfTheSeedFileIntoTheSeededRegisterChangesNothing(t *testing.T) {
	handlers := lab(t)
	data, err := os.ReadFile("../../seed/example.json")
	require.NoError(t, err)
	seed, err := seedfile.Read(bytes.NewReader(data))
	require.NoError(t, err)

	before, holdersBefore := observe(t, handlers), observeHolders(t, handlers)

	load := loadParticipants(t, handlers, string(data))

	assert.Empty(t, load.Added)
	assert.Equal(t, seedIDs(seed), load.Existing)
	assert.Equal(t, before, observe(t, handlers))
	assert.Equal(t, holdersBefore, observeHolders(t, handlers))
}

// Verifies: REG-1531, REG-1534, REG-1537, REG-1538, REG-1540
func TestALoadAddsNewParticipantsThatThePublicAPIsShowAndASecondLoadChangesNothing(t *testing.T) {
	handlers := lab(t)
	file := newParticipantsFile(newBrandJSON(newBrand, "New Bank"))

	changeRecipient(t, handlers, http.MethodPost, participantLoads, file, http.StatusOK, statusChange{
		recipients:  map[string]string{newRecipient: "ACTIVE"},
		products:    map[string]string{newProduct: "ACTIVE"},
		changedAPIs: []string{recipientStatusList, softwareProductList, dataRecipientList},
	})
	assert.Equal(t, "ACTIVE", observeHolders(t, handlers).statuses[newDataHolder])

	before, holdersBefore := observe(t, handlers), observeHolders(t, handlers)

	load := loadParticipants(t, handlers, file)

	assert.Empty(t, load.Added)
	assert.Equal(
		t,
		[]string{newDataHolder, newBrand, newRecipient, newRecipientBrand, newProduct},
		load.Existing,
	)
	assert.Equal(t, before, observe(t, handlers))
	assert.Equal(t, holdersBefore, observeHolders(t, handlers))
}

// Verifies: REG-1535, REG-1536, REG-1539
func TestALoadThatIsRefusedChangesNoRecord(t *testing.T) {
	handlers := lab(t)

	tests := map[string]struct {
		file   string
		status int
		detail string
	}{
		"an ID of a record of a different kind in a different legal entity": {
			file:   newParticipantsFile(newBrandJSON(mockBudgetWeb, "New Bank")),
			status: http.StatusConflict,
			detail: `the ID "` + mockBudgetWeb + `" is not unique`,
		},
		"a brand of a different legal entity": {
			file: `{"dataRecipients": [` + newRecipientJSON() + `, {
				"legalEntityId": "` + mockBudgetRecipient + `", "legalEntityName": "Mock Budgeting Pty Ltd",
				"accreditationNumber": "ADR1", "accreditationLevel": "UNRESTRICTED",
				"logoUri": "https://new.example/logo.png", "status": "ACTIVE",
				"dataRecipientBrands": [{
					"dataRecipientBrandId": "` + statusChangeBrand + `", "brandName": "Moved",
					"logoUri": "https://new.example/brand.png", "status": "ACTIVE", "softwareProducts": []
				}]
			}]}`,
			status: http.StatusConflict,
			detail: `the ID "` + statusChangeBrand + `" is not unique`,
		},
		"a record that is not valid": {
			file:   newParticipantsFile(newBrandJSON(newBrand, "")),
			status: http.StatusBadRequest,
			detail: "brandName is required",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			before, holdersBefore := observe(t, handlers), observeHolders(t, handlers)

			rec := send(t, handlers, http.MethodPost, participantLoads, tt.file)

			assert.Equal(t, tt.status, rec.Code, rec.Body.String())
			assert.Contains(t, problemOf(t, rec).Detail, tt.detail)
			assert.Equal(t, before, observe(t, handlers), "the refused load changes nothing")
			assert.Equal(t, holdersBefore, observeHolders(t, handlers), "the refused load changes nothing")
		})
	}
}
