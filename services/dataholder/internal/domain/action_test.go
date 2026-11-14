package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func currentState(t *testing.T, id string) domain.ArrangementState {
	t.Helper()

	state := createdState(t, time.Hour)
	state.Arrangement.ArrangementID = id

	return state
}

func TestALossOfEligibilityEndsEachArrangementThatDidNotEnd(t *testing.T) {
	current, ended := currentState(t, "current"), endedState(t, domain.EndReasonWithdrawnByConsumer)
	ended.Arrangement.ArrangementID = "ended"
	expired := currentState(t, "expired")
	expired.Arrangement.SharingDurationSeconds = 60

	ends := domain.IneligibilityEnds(
		[]domain.ArrangementState{current, ended, expired}, confirmedAt.Add(2*time.Minute),
	)

	assert.Equal(t, []domain.ArrangementState{current}, ends)
}

func TestEachOpenAccountHasACloseThatEndsNoArrangementWhenAnotherAccountStaysOpenAndACreateOfATransaction(
	t *testing.T,
) {
	accounts := []domain.Account{
		{AccountID: "a", OpenStatus: domain.OpenStatusOpen},
		{AccountID: "b", OpenStatus: domain.OpenStatusOpen},
	}

	actions := domain.PermittedAccountActions(
		accounts, []domain.ArrangementState{currentState(t, "current")}, confirmedAt,
	)

	assert.Equal(t, map[string][]domain.Action[domain.AccountAction]{
		"a": {
			{Action: domain.AccountActionClose, Ends: []domain.ArrangementState{}},
			{Action: domain.AccountActionCreateTransaction, Ends: []domain.ArrangementState{}},
		},
		"b": {
			{Action: domain.AccountActionClose, Ends: []domain.ArrangementState{}},
			{Action: domain.AccountActionCreateTransaction, Ends: []domain.ArrangementState{}},
		},
	}, actions)
}

func TestTheCloseOfTheLastOpenAccountEndsEachArrangementThatDidNotEndAndAClosedAccountHasNoAction(t *testing.T) {
	accounts := []domain.Account{
		{AccountID: "closed", OpenStatus: domain.OpenStatusClosed},
		{AccountID: "last", OpenStatus: domain.OpenStatusOpen},
	}
	current, ended := currentState(t, "current"), endedState(t, domain.EndReasonWithdrawnByConsumer)

	actions := domain.PermittedAccountActions(accounts, []domain.ArrangementState{current, ended}, confirmedAt)

	assert.Equal(t, map[string][]domain.Action[domain.AccountAction]{
		"closed": {},
		"last": {
			{Action: domain.AccountActionClose, Ends: []domain.ArrangementState{current}},
			{Action: domain.AccountActionCreateTransaction, Ends: []domain.ArrangementState{}},
		},
	}, actions)
}

func TestEachArrangementThatDidNotEndHasAWithdrawThatEndsOnlyThatArrangement(t *testing.T) {
	current := currentState(t, "current")
	ended := endedState(t, domain.EndReasonWithdrawnByConsumer)
	ended.Arrangement.ArrangementID = "ended"
	expired := currentState(t, "expired")
	expired.Arrangement.SharingDurationSeconds = 60

	actions := domain.PermittedArrangementActions(
		[]domain.ArrangementState{current, ended, expired}, confirmedAt.Add(2*time.Minute),
	)

	assert.Equal(t, map[string][]domain.Action[domain.ArrangementAction]{
		"current": {{Action: domain.ArrangementActionWithdraw, Ends: []domain.ArrangementState{current}}},
		"ended":   {},
		"expired": {},
	}, actions)
}
