package jobs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
)

const waitForJob = 10 * time.Second

type signalArgs struct{}

func (signalArgs) Kind() string { return "signal" }

type signalWorker struct {
	river.WorkerDefaults[signalArgs]

	started chan struct{}
	block   bool
}

func (w *signalWorker) Work(ctx context.Context, _ *river.Job[signalArgs]) error {
	close(w.started)

	if !w.block {
		return nil
	}

	<-ctx.Done()

	return ctx.Err()
}

func riverDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	require.NoError(t, err)
	_, err = migrator.Migrate(t.Context(), rivermigrate.DirectionUp, nil)
	require.NoError(t, err)

	return pool
}

func newClient(t *testing.T, pool *pgxpool.Pool, worker *signalWorker) *river.Client[pgx.Tx] {
	t.Helper()

	workers := river.NewWorkers()
	river.AddWorker(workers, worker)

	client, err := jobs.NewWorkerClient(pool, workers)
	require.NoError(t, err)

	return client
}

func job(t *testing.T, client *river.Client[pgx.Tx], id int64) *rivertype.JobRow {
	t.Helper()

	row, err := client.JobGet(t.Context(), id)
	require.NoError(t, err)

	return row
}

func TestRunWorksAJobAndStopsWhenTheContextEnds(t *testing.T) {
	pool := riverDatabase(t)
	worker := &signalWorker{started: make(chan struct{})}
	client := newClient(t, pool, worker)
	ctx, cancel := context.WithCancel(t.Context())

	inserted, err := client.Insert(t.Context(), signalArgs{}, nil)
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		done <- jobs.Run(ctx, client, func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		})
	}()

	select {
	case <-worker.started:
	case <-time.After(waitForJob):
		t.Fatal("the job did not start")
	}

	cancel()
	require.NoError(t, <-done)
	assert.Equal(t, rivertype.JobStateCompleted, job(t, client, inserted.Job.ID).State)
}

func TestRunCancelsARunningJobAndKeepsItForANewAttempt(t *testing.T) {
	pool := riverDatabase(t)
	worker := &signalWorker{started: make(chan struct{}), block: true}
	client := newClient(t, pool, worker)
	ctx, cancel := context.WithCancel(t.Context())

	inserted, err := client.Insert(t.Context(), signalArgs{}, nil)
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		done <- jobs.Run(ctx, client, func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		})
	}()

	<-worker.started
	cancel()
	require.NoError(t, <-done)

	// A stop before River records the attempt leaves the job running, and the rescue then gives the next attempt.
	assert.Nil(t, job(t, client, inserted.Job.ID).FinalizedAt)
}

func TestRunGivesTheErrorOfServeAndStopsTheClient(t *testing.T) {
	pool := riverDatabase(t)
	client := newClient(t, pool, &signalWorker{started: make(chan struct{})})
	serveErr := errors.New("serve failed")

	err := jobs.Run(t.Context(), client, func(context.Context) error { return serveErr })

	require.ErrorIs(t, err, serveErr)

	select {
	case <-client.Stopped():
	default:
		t.Fatal("the client did not stop")
	}
}

func TestAnInserterInsertsAJobInTheTransactionOfTheCaller(t *testing.T) {
	pool := riverDatabase(t)
	inserter, err := jobs.NewInsertOnlyClient(pool)
	require.NoError(t, err)

	tx, err := pool.Begin(t.Context())
	require.NoError(t, err)
	inserted, err := inserter.InsertTx(t.Context(), tx, signalArgs{}, nil)
	require.NoError(t, err)
	require.NoError(t, tx.Rollback(t.Context()))

	_, err = inserter.JobGet(t.Context(), inserted.Job.ID)
	require.ErrorIs(t, err, river.ErrNotFound, "the rollback removes the job")
}

func TestAPeriodicJobRunsAtTheStartOfTheClient(t *testing.T) {
	pool := riverDatabase(t)
	worker := &signalWorker{started: make(chan struct{})}
	workers := river.NewWorkers()
	river.AddWorker(workers, worker)

	client, err := jobs.NewWorkerClient(pool, workers, river.NewPeriodicJob(
		river.PeriodicInterval(time.Hour),
		func() (river.JobArgs, *river.InsertOpts) { return signalArgs{}, nil },
		&river.PeriodicJobOpts{RunOnStart: true},
	))
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))
	t.Cleanup(func() { _ = client.StopAndCancel(context.WithoutCancel(t.Context())) })

	select {
	case <-worker.started:
	case <-time.After(waitForJob):
		t.Fatal("the periodic job did not run")
	}
}
