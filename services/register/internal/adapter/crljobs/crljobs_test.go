package crljobs_test

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
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/crljobs"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

type renewals struct {
	done chan struct{}
	err  error
}

func (r *renewals) RenewCRL(context.Context) error {
	r.done <- struct{}{}

	return r.err
}

func TestTheCRLRenewalRunsAtTheStartAndThenAtEachInterval(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	crls := &renewals{done: make(chan struct{}, 10)}
	workers := river.NewWorkers()
	river.AddWorker(workers, crljobs.Worker{CRLs: crls})
	client, err := jobs.NewWorkerClient(pool, workers, crljobs.PeriodicJob(time.Second))
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))
	t.Cleanup(func() { _ = client.StopAndCancel(context.WithoutCancel(t.Context())) })

	for i := range 2 {
		select {
		case <-crls.done:
		case <-time.After(10 * time.Second):
			t.Fatalf("renewal %d did not occur", i+1)
		}
	}
}

func TestAFailedRenewalHasNoNewAttemptBecauseTheNextIntervalTriesAgain(t *testing.T) {
	opts := crljobs.Args{}.InsertOpts()

	assert.Equal(t, 1, opts.MaxAttempts)
	assert.True(t, opts.UniqueOpts.ByArgs, "a slow renewal does not overlap the next one")
	assert.NotContains(t, opts.UniqueOpts.ByState, rivertype.JobStateCompleted, "a renewal that ended permits the next")
}

func TestAFailedRenewalGivesRiverNoTextOfTheError(t *testing.T) {
	crls := &renewals{done: make(chan struct{}, 1), err: errors.New(`connect to "db.example" as "user"`)}

	err := crljobs.Worker{CRLs: crls}.Work(
		t.Context(), &river.Job[crljobs.Args]{JobRow: &rivertype.JobRow{Attempt: 1}},
	)

	require.EqualError(t, err, "the CRL was not renewed")
}
