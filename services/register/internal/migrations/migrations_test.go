package migrations_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

const (
	insufficientPrivilege = "42501"
	applicationRole       = "register_app"
)

func TestMigrationsGiveTheApplicationRoleDataAccessOnly(t *testing.T) {
	// The role is global to the server, thus the migrations of the second database find it.
	for range 2 {
		pool := databasetest.New(t)
		require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

		assert.Empty(t, databasetest.ApplicationRoleProblems(t, pool, applicationRole, "public.events"))
	}
}

func TestTheApplicationRoleCanReindexTheRiverJobTable(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	_, err := databasetest.AsRole(t, pool, applicationRole).
		Exec(t.Context(), "REINDEX INDEX CONCURRENTLY river_job_kind")

	require.NoError(t, err)
}

type probeArgs struct{}

func (probeArgs) Kind() string { return "probe" }

type probeWorker struct {
	river.WorkerDefaults[probeArgs]

	worked chan struct{}
}

func (w probeWorker) Work(context.Context, *river.Job[probeArgs]) error {
	close(w.worked)
	return nil
}

func TestRiverWorksAJobOnTheSchemaOfTheMigrationsAsTheApplicationRole(t *testing.T) {
	owner := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), owner, migrations.FS))

	pool := databasetest.AsRole(t, owner, applicationRole)

	worker := probeWorker{worked: make(chan struct{})}
	workers := river.NewWorkers()
	river.AddWorker(workers, worker)
	client, err := jobs.NewWorkerClient(pool, workers)
	require.NoError(t, err)

	_, err = client.Insert(t.Context(), probeArgs{}, &river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}})
	require.NoError(t, err)

	require.NoError(t, client.Start(t.Context()))
	t.Cleanup(func() { _ = client.StopAndCancel(context.WithoutCancel(t.Context())) })

	select {
	case <-worker.worked:
	case <-time.After(10 * time.Second):
		t.Fatal("River did not work the job")
	}
}

func TestTheApplicationRoleCanAppendEventsButNotChangeOrRemoveThem(t *testing.T) {
	owner := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), owner, migrations.FS))
	pool := databasetest.AsRole(t, owner, applicationRole)

	require.NoError(t, eventstore.New(pool).Append(t.Context(), eventsourcing.Append{
		StreamType: "test",
		StreamID:   "s-1",
		Events:     []eventsourcing.EventData{{EventType: "Tested", SchemaVersion: 1, Payload: []byte(`{}`)}},
	}))

	changes := []string{"UPDATE events SET event_type = 'Changed'", "DELETE FROM events", "TRUNCATE events"}
	for _, statement := range changes {
		_, err := pool.Exec(t.Context(), statement)

		var pgErr *pgconn.PgError
		require.ErrorAs(t, err, &pgErr, statement)
		assert.Equal(t, insufficientPrivilege, pgErr.Code, statement)
	}
}
