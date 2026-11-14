package statusjobs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusjobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

type refreshes struct {
	done chan struct{}
	err  error
}

func (r *refreshes) Refresh(context.Context) error {
	r.done <- struct{}{}

	return r.err
}

func TestTheStatusesAreRefreshedAtTheStartAndThenAtEachInterval(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	statuses := &refreshes{done: make(chan struct{}, 10)}
	workers := river.NewWorkers()
	river.AddWorker(workers, statusjobs.Worker{Statuses: statuses})
	client, err := jobs.NewWorkerClient(pool, workers, statusjobs.PeriodicJob(time.Second))
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))
	t.Cleanup(func() { _ = client.StopAndCancel(context.WithoutCancel(t.Context())) })

	for i := range 2 {
		select {
		case <-statuses.done:
		case <-time.After(10 * time.Second):
			t.Fatalf("refresh %d did not occur", i+1)
		}
	}
}

func TestAFailedRefreshHasNoNewAttemptBecauseTheNextIntervalTriesAgain(t *testing.T) {
	opts := statusjobs.Args{}.InsertOpts()

	assert.Equal(t, 1, opts.MaxAttempts)
	assert.True(t, opts.UniqueOpts.ByArgs, "a slow refresh does not overlap the next one")
	assert.NotContains(t, opts.UniqueOpts.ByState, rivertype.JobStateCompleted, "a refresh that ended permits the next")
}

func TestAFailedRefreshGivesRiverNoTextOfTheError(t *testing.T) {
	statuses := &refreshes{done: make(chan struct{}, 1), err: errors.New(`decode: invalid character '<' in "secret"`)}

	err := statusjobs.Worker{Statuses: statuses}.Work(
		t.Context(), &river.Job[statusjobs.Args]{JobRow: &rivertype.JobRow{Attempt: 1}},
	)

	require.EqualError(t, err, "the statuses were not refreshed")
}
