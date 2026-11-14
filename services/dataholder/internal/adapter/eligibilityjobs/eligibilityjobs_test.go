package eligibilityjobs_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/eligibilityjobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

var closedAt = time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)

func newInserter(t *testing.T) (*pgxpool.Pool, *river.Client[pgx.Tx]) {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	inserter, err := jobs.NewInsertOnlyClient(pool)
	require.NoError(t, err)

	return pool, inserter
}

func eligibilityJobs(t *testing.T, client *river.Client[pgx.Tx]) []*rivertype.JobRow {
	t.Helper()

	listed, err := client.JobList(t.Context(), river.NewJobListParams().Kinds(eligibilityjobs.Args{}.Kind()))
	require.NoError(t, err)

	return listed.Jobs
}

func TestTheHookInsertsOneJobForTheCloseThatRunsAfterTheRequestEndsTheArrangements(t *testing.T) {
	pool, inserter := newInserter(t)
	hook := eligibilityjobs.Hook(inserter)

	require.NoError(t, pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error {
		return errors.Join(
			hook(t.Context(), tx, "00000001", closedAt),
			hook(t.Context(), tx, "00000001", closedAt),
		)
	}))

	got := eligibilityJobs(t, inserter)
	require.Len(t, got, 1, "the job is unique for the close")
	assert.Equal(t, closedAt.Add(time.Minute), got[0].ScheduledAt.UTC())

	var args eligibilityjobs.Args
	require.NoError(t, json.Unmarshal(got[0].EncodedArgs, &args))
	assert.Equal(t, eligibilityjobs.Args{CustomerNumber: "00000001", ClosedAt: closedAt}, args)
}

func TestARollbackOfTheCloseRemovesTheJob(t *testing.T) {
	pool, inserter := newInserter(t)
	rollback := errors.New("the close failed")

	err := pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error {
		if err := eligibilityjobs.Hook(inserter)(t.Context(), tx, "00000001", closedAt); err != nil {
			return err
		}

		return rollback
	})

	require.ErrorIs(t, err, rollback)
	assert.Empty(t, eligibilityJobs(t, inserter))
}

type ends struct {
	customerNumber string
	closedAt       time.Time
	err            error
}

func (e *ends) EndForIneligibleCustomer(_ context.Context, customerNumber string, closedAt time.Time) error {
	e.customerNumber, e.closedAt = customerNumber, closedAt

	return e.err
}

func job(args eligibilityjobs.Args) *river.Job[eligibilityjobs.Args] {
	return &river.Job[eligibilityjobs.Args]{JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 25}, Args: args}
}

func TestTheWorkerEndsTheArrangementsOfTheCustomerAtTheTimeOfTheClose(t *testing.T) {
	ender := &ends{}

	err := eligibilityjobs.Worker{Arrangements: ender}.Work(
		t.Context(), job(eligibilityjobs.Args{CustomerNumber: "00000001", ClosedAt: closedAt}),
	)

	require.NoError(t, err)
	assert.Equal(t, "00000001", ender.customerNumber)
	assert.Equal(t, closedAt, ender.closedAt)
}

func TestAnErrorOfTheEndGivesRiverNoTextOfTheError(t *testing.T) {
	ender := &ends{err: errors.New(`duplicate key: customer_number "00000001"`)}

	err := eligibilityjobs.Worker{Arrangements: ender}.Work(
		t.Context(), job(eligibilityjobs.Args{CustomerNumber: "00000001", ClosedAt: closedAt}),
	)

	require.EqualError(t, err, "internal error", "River keeps the text in the job and in telemetry")
}
