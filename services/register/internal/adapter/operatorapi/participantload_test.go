package operatorapi_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const participantLoadsPath = "/v1/participant-loads"

// Verifies: REG-1531, REG-1532, REG-1537
func TestAParticipantLoadGivesTheAddedAndTheExistingIDs(t *testing.T) {
	operator := &fakeOperator{
		load: domain.ParticipantLoad{Added: []string{"dr-2", "drb-2"}, Existing: []string{"dh-1"}},
	}

	rec := serveBody(t, operator, participantLoadsPath, []byte(`{
		"dataHolders": [{"legalEntityId": "dh-1"}],
		"dataRecipients": [{"legalEntityId": "dr-2", "dataRecipientBrands": [{"dataRecipientBrandId": "drb-2"}]}]
	}`))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"added": ["dr-2", "drb-2"], "existing": ["dh-1"]}`, rec.Body.String())
	assert.Equal(t, "dh-1", operator.participants.DataHolders[0].LegalEntityID)
	assert.Equal(t, "drb-2", operator.participants.DataRecipients[0].Brands[0].DataRecipientBrandID)
}

// Verifies: REG-1537
func TestAParticipantLoadGivesEmptyListsAsArrays(t *testing.T) {
	rec := serveBody(t, &fakeOperator{}, participantLoadsPath, []byte(`{}`))

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, `{"added": [], "existing": []}`, rec.Body.String())
}

func TestAParticipantLoadRefusesABodyThatIsNotALoadFile(t *testing.T) {
	for name, body := range map[string]string{
		"unknown field":  `{"participants": []}`,
		"second value":   `{} {}`,
		"not an object":  `[]`,
		"unknown nested": `{"dataHolders": [{"legalEntityNumber": "dh-1"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			operator := &fakeOperator{}

			requireProblem(t, serveBody(t, operator, participantLoadsPath, []byte(body)), http.StatusBadRequest)
			assert.Empty(t, operator.participants)
		})
	}
}

func TestAParticipantLoadRefusesABodyLargerThan1MiB(t *testing.T) {
	body := append([]byte(`{"dataHolders": [{"legalEntityName": "`), bytes.Repeat([]byte("a"), 1<<20)...)

	requireProblem(t, serveBody(t, &fakeOperator{}, participantLoadsPath, append(body, `"}]}`...)),
		http.StatusRequestEntityTooLarge)
}

// Verifies: REG-1535, REG-1536
func TestAParticipantLoadGives400WhenOneRecordIsNotValidAlone(t *testing.T) {
	ruleErr := statusRuleError(t)
	notValid := domain.Participants{DataHolders: []domain.DataHolder{{LegalEntityID: "dh-1"}, {LegalEntityID: "dh-1"}}}.
		Validate()
	require.ErrorIs(t, notValid, domain.ErrNotValid)

	tests := map[string]struct {
		err        error
		status     int
		wantDetail string
	}{
		"status rule": {ruleErr, http.StatusConflict, ruleErr.Error()},
		"ID of a different record": {
			domain.IDNotUniqueError{ID: "sp-1"},
			http.StatusConflict,
			domain.IDNotUniqueError{ID: "sp-1"}.Error(),
		},
		"not valid": {notValid, http.StatusBadRequest, notValid.Error()},
		"status rule and not valid": {
			errors.Join(ruleErr, notValid),
			http.StatusBadRequest,
			errors.Join(ruleErr, notValid).Error(),
		},
		"version conflict": {
			fmt.Errorf("data recipient dr-1: %w", eventsourcing.ErrVersionConflict),
			http.StatusConflict,
			"The participants changed during the request. Send the request again.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := requireProblem(
				t,
				serveBody(t, &fakeOperator{err: tt.err}, participantLoadsPath, []byte(`{}`)),
				tt.status,
			)

			assert.Equal(t, tt.wantDetail, p.Detail)
		})
	}
}
