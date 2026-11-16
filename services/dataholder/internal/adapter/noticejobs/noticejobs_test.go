package noticejobs_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/noticejobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var thisAttemptAt = time.Date(2026, 10, 17, 1, 1, 30, 0, time.UTC)

// recorded has no status for an attempt after which the notice stays pending.
type recorded struct {
	status  domain.NoticeStatus
	attempt domain.NoticeAttempt
}

type fakeNotices struct {
	result app.NoticeAttemptResult
	err    error

	// mu guards recorded, which a River worker writes in a different goroutine.
	mu       sync.Mutex
	recorded []recorded
}

func (f *fakeNotices) records() []recorded {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.recorded
}

func (f *fakeNotices) Attempt(_ context.Context, _ string, attempt int) (app.NoticeAttemptResult, error) {
	result := f.result
	result.Attempt = attempt

	return result, f.err
}

func (f *fakeNotices) RecordAttempt(_ context.Context, _ string, attempt domain.NoticeAttempt) error {
	return f.record(recorded{attempt: attempt})
}

func (f *fakeNotices) Record(
	_ context.Context,
	_ string,
	status domain.NoticeStatus,
	last domain.NoticeAttempt,
) error {
	return f.record(recorded{status: status, attempt: last})
}

func (f *fakeNotices) record(r recorded) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.recorded = append(f.recorded, r)

	return nil
}

func attemptResult(result string, outcome domain.NoticeOutcome) app.NoticeAttemptResult {
	return app.NoticeAttemptResult{
		At: thisAttemptAt, Result: result,
		Outcome: outcome,
	}
}

func job(attempt int) *river.Job[noticejobs.Args] {
	return &river.Job[noticejobs.Args]{
		JobRow: &rivertype.JobRow{Attempt: attempt, MaxAttempts: 3},
		Args:   noticejobs.Args{ArrangementID: "arrangement-1"},
	}
}

func work(t *testing.T, notices *fakeNotices, j *river.Job[noticejobs.Args]) error {
	t.Helper()

	return noticejobs.Worker{Notices: notices}.Work(t.Context(), j)
}

func requireCancelled(t *testing.T, err error) {
	t.Helper()

	var cancel *river.JobCancelError
	require.ErrorAs(t, err, &cancel)
}

// Verifies: DH-0933, DH-0939
func TestADeliveredNoticeIsRecordedCompleteWithItsLastAttempt(t *testing.T) {
	notices := &fakeNotices{result: attemptResult("status 204", domain.NoticeDelivered)}

	require.NoError(t, work(t, notices, job(3)))

	assert.Equal(t, []recorded{
		{
			status:  domain.NoticeComplete,
			attempt: domain.NoticeAttempt{Attempt: 3, At: thisAttemptAt, Result: "status 204"},
		},
	}, notices.recorded)
}

// Verifies: DH-0935
func TestARefusedNoticeIsRecordedFailedWithNoNewAttempt(t *testing.T) {
	notices := &fakeNotices{result: attemptResult("status 400", domain.NoticeRefused)}

	err := work(t, notices, job(1))

	requireCancelled(t, err)
	assert.Equal(t, []recorded{{
		status: domain.NoticeFailed, attempt: domain.NoticeAttempt{Attempt: 1, At: thisAttemptAt, Result: "status 400"},
	}}, notices.recorded)
}

// Verifies: DH-0934, DH-0940
func TestANoticeToTryAgainRecordsTheAttemptAndGetsANewAttemptWithTheResultAsItsError(t *testing.T) {
	notices := &fakeNotices{result: attemptResult("status 503", domain.NoticeTryAgain)}

	err := work(t, notices, job(2))

	require.EqualError(t, err, "status 503")
	assert.Equal(t, []recorded{{
		attempt: domain.NoticeAttempt{Attempt: 2, At: thisAttemptAt, Result: "status 503"},
	}}, notices.recorded)
}

// Verifies: DH-0936
func TestTheThirdAttemptThatFailsIsRecordedFailed(t *testing.T) {
	notices := &fakeNotices{result: attemptResult("no connection", domain.NoticeTryAgain)}

	err := work(t, notices, job(3))

	requireCancelled(t, err)
	assert.Equal(t, []recorded{
		{
			status:  domain.NoticeFailed,
			attempt: domain.NoticeAttempt{Attempt: 3, At: thisAttemptAt, Result: "no connection"},
		},
	}, notices.recorded)
}

func TestAJobForANoticeThatIsNotPendingEndsWithNoRecord(t *testing.T) {
	notices := &fakeNotices{err: domain.ErrNoNoticePending}

	require.NoError(t, work(t, notices, job(1)))
	assert.Empty(t, notices.recorded)
}

// Verifies: DH-0936
func TestTheSecondAttemptIs10SecondsAndTheThird60SecondsAfterTheAttemptBefore(t *testing.T) {
	worker := noticejobs.Worker{}

	for attempt, wait := range map[int]time.Duration{1: 10 * time.Second, 2: time.Minute} {
		before := time.Now()

		next := worker.NextRetry(job(attempt))

		assert.WithinRange(t, next, before.Add(wait), time.Now().Add(wait), "after attempt %d", attempt)
	}
}

func TestEachAttemptHasATimeLimitLongerThanTheTimeoutOfTheRequest(t *testing.T) {
	assert.Greater(t, noticejobs.Worker{}.Timeout(job(1)), 10*time.Second)
}

func TestANoticeJobHas3AttemptsAndIsUniqueForItsArrangement(t *testing.T) {
	opts := noticejobs.Args{}.InsertOpts()

	assert.Equal(t, 3, opts.MaxAttempts)
	assert.True(t, opts.UniqueOpts.ByArgs)
}
