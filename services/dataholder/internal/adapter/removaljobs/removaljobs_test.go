package removaljobs_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/removaljobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

var removedAt = time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)

func TestTheHookInsertsOneJobForTheRemovalInTheTransactionOfTheStatus(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	inserter, err := jobs.NewInsertOnlyClient(pool)
	require.NoError(t, err)

	hook := removaljobs.Hook(inserter)

	require.NoError(t, pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error {
		return errors.Join(
			hook(t.Context(), tx, "software-1", removedAt),
			hook(t.Context(), tx, "software-1", removedAt),
		)
	}))

	rollback := errors.New("the save failed")
	err = pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error {
		return errors.Join(hook(t.Context(), tx, "software-2", removedAt), rollback)
	})
	require.ErrorIs(t, err, rollback)

	listed, err := inserter.JobList(t.Context(), river.NewJobListParams().Kinds(removaljobs.Args{}.Kind()))
	require.NoError(t, err)
	require.Len(t, listed.Jobs, 1)

	var args removaljobs.Args
	require.NoError(t, json.Unmarshal(listed.Jobs[0].EncodedArgs, &args))
	assert.Equal(t, removaljobs.Args{SoftwareProductID: "software-1", RemovedAt: removedAt}, args)
}

type removals struct {
	softwareID string
	removedAt  time.Time
	err        error
}

func (r *removals) RemoveSoftwareProduct(_ context.Context, softwareID string, removedAt time.Time) error {
	r.softwareID, r.removedAt = softwareID, removedAt

	return r.err
}

func job() *river.Job[removaljobs.Args] {
	return &river.Job[removaljobs.Args]{
		JobRow: &rivertype.JobRow{Attempt: 1},
		Args:   removaljobs.Args{SoftwareProductID: "software-1", RemovedAt: removedAt},
	}
}

func TestTheWorkerRemovesTheSoftwareProductAtTheTimeOfTheStatus(t *testing.T) {
	products := &removals{}

	require.NoError(t, removaljobs.Worker{Products: products}.Work(t.Context(), job()))

	assert.Equal(t, "software-1", products.softwareID)
	assert.Equal(t, removedAt, products.removedAt)
}

func TestAFailedRemovalGivesRiverNoTextOfTheError(t *testing.T) {
	products := &removals{err: errors.New(`duplicate key: client_id "client-1"`)}

	err := removaljobs.Worker{Products: products}.Work(t.Context(), job())

	require.EqualError(t, err, "internal error")
}
