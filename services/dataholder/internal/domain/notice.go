package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

var ErrNoNoticePending = errors.New("the arrangement has no notice that is pending")

type NoticeStatus string

const (
	NoticePending  NoticeStatus = "PENDING"
	NoticeComplete NoticeStatus = "COMPLETE"
	NoticeFailed   NoticeStatus = "FAILED"
)

func (r EndReason) NeedsNotice() bool {
	return r == EndReasonWithdrawnByConsumer || r == EndReasonConsumerNotEligible
}

// NoticeAttempt has a Result that the data holder classifies, for example "status 503" or "timeout", and that
// contains no value of the response.
type NoticeAttempt struct {
	Attempt int       `json:"attempt"`
	At      time.Time `json:"at"`
	Result  string    `json:"result"`
}

// RecordNoticeAttempt records an attempt after which the notice stays pending.
type RecordNoticeAttempt struct {
	Attempt NoticeAttempt
}

func (RecordNoticeAttempt) isArrangementCommand() {}

// RecordNotice records the outcome of the notice, with its last attempt.
type RecordNotice struct {
	RecordedAt time.Time
	Status     NoticeStatus
	Attempt    NoticeAttempt
}

func (RecordNotice) isArrangementCommand() {}

type NoticeAttempted struct {
	NoticeAttempt
}

func (NoticeAttempted) isArrangementEvent() {}

// NoticeRecorded has each attempt of the notice, also those of the NoticeAttempted events before it, so that one
// event is the full record of the notice.
type NoticeRecorded struct {
	RecordedAt time.Time       `json:"recordedAt"`
	Status     NoticeStatus    `json:"status"`
	Attempts   []NoticeAttempt `json:"attempts"`
}

func (NoticeRecorded) isArrangementEvent() {}

func (s ArrangementState) NoticeStatus() NoticeStatus {
	switch {
	case !s.EndReason.NeedsNotice():
		return ""
	case s.RecordedNotice != "":
		return s.RecordedNotice
	default:
		return NoticePending
	}
}

func handleRecordNoticeAttempt(state ArrangementState, command RecordNoticeAttempt) ([]ArrangementEvent, error) {
	if state.NoticeStatus() != NoticePending {
		return nil, ErrNoNoticePending
	}

	return []ArrangementEvent{NoticeAttempted{NoticeAttempt: command.Attempt.utc()}}, nil
}

func handleRecordNotice(state ArrangementState, command RecordNotice) ([]ArrangementEvent, error) {
	if state.NoticeStatus() != NoticePending {
		return nil, ErrNoNoticePending
	}

	if command.Status != NoticeComplete && command.Status != NoticeFailed {
		return nil, fmt.Errorf("a notice cannot be recorded as %q", command.Status)
	}

	return []ArrangementEvent{NoticeRecorded{
		RecordedAt: command.RecordedAt.UTC(), Status: command.Status,
		Attempts: append(slices.Clone(state.NoticeAttempts), command.Attempt.utc()),
	}}, nil
}

func (a NoticeAttempt) utc() NoticeAttempt {
	a.At = a.At.UTC()

	return a
}

// The domain does no I/O, so it does not import the net/http names of these RFC 9110 statuses.
const (
	statusNoContent           = 204
	statusUnprocessableEntity = 422
	statusMinServerError      = 500
)

type NoticeOutcome int

const (
	NoticeDelivered NoticeOutcome = iota + 1
	NoticeTryAgain
	NoticeRefused
)

// NoticeResponseOutcome gives NoticeDelivered for 204 and 422, because the data recipient then knows that the
// arrangement ended.
func NoticeResponseOutcome(status int) NoticeOutcome {
	switch {
	case status == statusNoContent || status == statusUnprocessableEntity:
		return NoticeDelivered
	case status >= statusMinServerError:
		return NoticeTryAgain
	default:
		return NoticeRefused
	}
}
