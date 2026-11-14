package domain_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func endedState(t *testing.T, reason domain.EndReason) domain.ArrangementState {
	t.Helper()

	state := createdState(t, time.Hour)
	events, err := domain.HandleArrangement(state, domain.EndArrangement{
		EndedAt: confirmedAt.Add(time.Minute), Reason: reason,
	})
	require.NoError(t, err)

	return domain.ApplyArrangement(state, events[0])
}

var (
	firstAttempt = domain.NoticeAttempt{Attempt: 1, At: confirmedAt.Add(time.Minute), Result: "status 503"}
	lastAttempt  = domain.NoticeAttempt{Attempt: 2, At: confirmedAt.Add(2 * time.Minute), Result: "status 204"}
)

func recordNotice(status domain.NoticeStatus) domain.RecordNotice {
	return domain.RecordNotice{RecordedAt: confirmedAt.Add(2 * time.Minute), Status: status, Attempt: lastAttempt}
}

func apply(t *testing.T, state domain.ArrangementState, command domain.ArrangementCommand) domain.ArrangementState {
	t.Helper()

	events, err := domain.HandleArrangement(state, command)
	require.NoError(t, err)

	for _, event := range events {
		state = domain.ApplyArrangement(state, event)
	}

	return state
}

func TestOnlyAnEndThatNeedsANoticeHasANoticeStatus(t *testing.T) {
	assert.Equal(t, domain.NoticePending, endedState(t, domain.EndReasonWithdrawnByConsumer).NoticeStatus())
	assert.Equal(t, domain.NoticePending, endedState(t, domain.EndReasonConsumerNotEligible).NoticeStatus())
	assert.Empty(t, endedState(t, domain.EndReasonRevokedByDataRecipient).NoticeStatus())
	assert.Empty(t, endedState(t, domain.EndReasonSoftwareProductRemoved).NoticeStatus())
	assert.Empty(t, createdState(t, time.Hour).NoticeStatus())
}

func TestAnAttemptThatLeavesTheNoticePendingIsRecorded(t *testing.T) {
	state := endedState(t, domain.EndReasonWithdrawnByConsumer)

	events, err := domain.HandleArrangement(state, domain.RecordNoticeAttempt{Attempt: firstAttempt})

	require.NoError(t, err)
	assert.Equal(t, []domain.ArrangementEvent{domain.NoticeAttempted{NoticeAttempt: firstAttempt}}, events)

	state = domain.ApplyArrangement(state, events[0])
	assert.Equal(t, domain.NoticePending, state.NoticeStatus())
	assert.Equal(t, []domain.NoticeAttempt{firstAttempt}, state.NoticeAttempts)
}

func TestARecordedNoticeHasItsStatusAndEachAttempt(t *testing.T) {
	state := apply(t, endedState(t, domain.EndReasonWithdrawnByConsumer),
		domain.RecordNoticeAttempt{Attempt: firstAttempt})

	events, err := domain.HandleArrangement(state, recordNotice(domain.NoticeComplete))

	require.NoError(t, err)
	assert.Equal(t, []domain.ArrangementEvent{domain.NoticeRecorded{
		RecordedAt: confirmedAt.Add(2 * time.Minute), Status: domain.NoticeComplete,
		Attempts: []domain.NoticeAttempt{firstAttempt, lastAttempt},
	}}, events)

	state = domain.ApplyArrangement(state, events[0])
	assert.Equal(t, domain.NoticeComplete, state.NoticeStatus())
	assert.Equal(t, []domain.NoticeAttempt{firstAttempt, lastAttempt}, state.NoticeAttempts)
}

func TestTheAttemptsOfANoticeRecordedWithNoEarlierEventAreThoseOfTheRecord(t *testing.T) {
	state := domain.ApplyArrangement(endedState(t, domain.EndReasonWithdrawnByConsumer), domain.NoticeRecorded{
		Status: domain.NoticeFailed, Attempts: []domain.NoticeAttempt{firstAttempt},
	})

	assert.Equal(t, []domain.NoticeAttempt{firstAttempt}, state.NoticeAttempts)
}

func TestANoticeAndItsAttemptsAreRecordedOnlyWhenItIsPending(t *testing.T) {
	pending := endedState(t, domain.EndReasonWithdrawnByConsumer)
	events, err := domain.HandleArrangement(pending, recordNotice(domain.NoticeFailed))
	require.NoError(t, err)

	tests := map[string]domain.ArrangementState{
		"notice recorded":     domain.ApplyArrangement(pending, events[0]),
		"end with no notice":  endedState(t, domain.EndReasonRevokedByDataRecipient),
		"arrangement not end": createdState(t, time.Hour),
	}
	for name, state := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleArrangement(state, recordNotice(domain.NoticeComplete))
			require.ErrorIs(t, err, domain.ErrNoNoticePending)

			_, err = domain.HandleArrangement(state, domain.RecordNoticeAttempt{Attempt: firstAttempt})
			require.ErrorIs(t, err, domain.ErrNoNoticePending)
		})
	}
}

func TestANoticeIsRecordedOnlyAsCompleteOrFailed(t *testing.T) {
	_, err := domain.HandleArrangement(
		endedState(t, domain.EndReasonWithdrawnByConsumer), recordNotice(domain.NoticePending),
	)

	require.Error(t, err)
}

func TestTheOutcomeOfANoticeResponse(t *testing.T) {
	tests := map[int]domain.NoticeOutcome{
		http.StatusNoContent:           domain.NoticeDelivered,
		http.StatusUnprocessableEntity: domain.NoticeDelivered,
		http.StatusOK:                  domain.NoticeRefused,
		http.StatusFound:               domain.NoticeRefused,
		http.StatusBadRequest:          domain.NoticeRefused,
		http.StatusUnauthorized:        domain.NoticeRefused,
		http.StatusInternalServerError: domain.NoticeTryAgain,
		http.StatusServiceUnavailable:  domain.NoticeTryAgain,
	}
	for status, want := range tests {
		assert.Equal(t, want, domain.NoticeResponseOutcome(status), "status %d", status)
	}
}
