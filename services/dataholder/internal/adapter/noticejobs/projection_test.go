package noticejobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/noticejobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

const waitForWork = 10 * time.Second

func newStore(t *testing.T) (*pgxpool.Pool, *eventstore.Store, *river.Client[pgx.Tx]) {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	inserter, err := jobs.NewInsertOnlyClient(pool)
	require.NoError(t, err)

	return pool, eventstore.New(pool, noticejobs.Projection(inserter)), inserter
}

func withdrawal() domain.ArrangementCommand {
	return domain.WithdrawArrangement{WithdrawnAt: time.Now(), Method: domain.WithdrawalMethodAlternative}
}

func endArrangement(t *testing.T, store *eventstore.Store, end domain.ArrangementCommand) {
	t.Helper()

	confirmed := time.Now().Add(-time.Minute)
	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "arrangement-1", domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-1", CustomerNumber: "00000001", ClientID: "client-1",
			ConfirmedAt: confirmed, SharingDurationSeconds: 3600,
		},
	}))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "arrangement-1", end))
}

func noticeJobs(t *testing.T, client *river.Client[pgx.Tx]) []*rivertype.JobRow {
	t.Helper()

	listed, err := client.JobList(t.Context(), river.NewJobListParams().Kinds(noticejobs.Args{}.Kind()))
	require.NoError(t, err)

	return listed.Jobs
}

// Verifies: DH-0924
func TestAnEndThatNeedsANoticeInsertsOneJob(t *testing.T) {
	_, store, inserter := newStore(t)

	endArrangement(t, store, withdrawal())
	require.NoError(t, store.ProjectAll(t.Context()))

	listed := noticeJobs(t, inserter)
	require.Len(t, listed, 1, "a new projection inserts no second job")
	assert.JSONEq(t, `{"arrangement_id":"arrangement-1"}`, string(listed[0].EncodedArgs))
	assert.Equal(t, 3, listed[0].MaxAttempts)
}

// Verifies: DH-0924
func TestAnEndThatNeedsNoNoticeInsertsNoJob(t *testing.T) {
	_, store, inserter := newStore(t)

	endArrangement(t, store, domain.RevokeArrangement{RevokedAt: time.Now()})

	assert.Empty(t, noticeJobs(t, inserter))
}

// sequenceNotices tries again at the first attempt, and delivers at the second.
type sequenceNotices struct {
	fakeNotices

	attempted chan int
}

func (s *sequenceNotices) Attempt(_ context.Context, _ string, attempt int) (app.NoticeAttemptResult, error) {
	s.attempted <- attempt

	outcome, result := domain.NoticeTryAgain, "status 503"
	if attempt > 1 {
		outcome, result = domain.NoticeDelivered, "status 204"
	}

	return app.NoticeAttemptResult{
		Attempt: attempt, At: time.Now(), Result: result, Outcome: outcome,
	}, nil
}

func startWorker(t *testing.T, pool *pgxpool.Pool, notices noticejobs.Notices) *river.Client[pgx.Tx] {
	t.Helper()

	workers := river.NewWorkers()
	river.AddWorker(workers, noticejobs.Worker{Notices: notices})
	client, err := jobs.NewWorkerClient(pool, workers)
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))

	return client
}

func requireAttempt(t *testing.T, attempted <-chan int, want int) {
	t.Helper()

	select {
	case got := <-attempted:
		require.Equal(t, want, got)
	case <-time.After(waitForWork):
		t.Fatalf("no attempt %d", want)
	}
}

// Verifies: DH-0938
func TestANoticeContinuesItsAttemptsAfterARestart(t *testing.T) {
	pool, store, inserter := newStore(t)
	notices := &sequenceNotices{attempted: make(chan int, 2)}

	endArrangement(t, store, withdrawal())

	first := startWorker(t, pool, notices)
	requireAttempt(t, notices.attempted, 1)
	require.NoError(t, first.Stop(t.Context()))

	// The test does not wait 10 seconds for the second attempt, or for the River scheduler to make the job available.
	tag, err := pool.Exec(t.Context(),
		"UPDATE river_job SET state = 'available', scheduled_at = now() WHERE state = 'retryable' AND attempt = 1")
	require.NoError(t, err)
	require.Equal(t, int64(1), tag.RowsAffected(), "the stop keeps the job for its second attempt")

	second := startWorker(t, pool, notices)
	t.Cleanup(func() { _ = second.StopAndCancel(context.WithoutCancel(t.Context())) })
	requireAttempt(t, notices.attempted, 2)

	require.Eventually(t, func() bool {
		listed := noticeJobs(t, inserter)
		return len(listed) == 1 && listed[0].State == rivertype.JobStateCompleted
	}, waitForWork, 50*time.Millisecond)

	records := notices.records()
	require.Len(t, records, 2)
	assert.Equal(t, recorded{attempt: domain.NoticeAttempt{Attempt: 1, Result: "status 503"}}, withoutTime(records[0]))
	assert.Equal(t, recorded{
		status: domain.NoticeComplete, attempt: domain.NoticeAttempt{Attempt: 2, Result: "status 204"},
	}, withoutTime(records[1]))
}

func withoutTime(r recorded) recorded {
	r.attempt.At = time.Time{}

	return r
}

func TestARecordedAttemptInsertsNoSecondJob(t *testing.T) {
	_, store, inserter := newStore(t)
	endArrangement(t, store, withdrawal())

	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "arrangement-1", domain.RecordNoticeAttempt{
		Attempt: domain.NoticeAttempt{Attempt: 1, At: time.Now(), Result: "status 503"},
	}))

	assert.Len(t, noticeJobs(t, inserter), 1)
}
