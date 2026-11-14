package database_test

import (
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
)

var (
	v1 = &fstest.MapFile{Data: []byte("-- +goose Up\nCREATE TABLE one (id int PRIMARY KEY);\n")}
	v2 = &fstest.MapFile{Data: []byte("-- +goose Up\nCREATE TABLE two (id int PRIMARY KEY);\n")}
	// The first statement succeeds and the second fails, so the migration must roll back.
	v2Fails = &fstest.MapFile{Data: []byte("-- +goose Up\nCREATE TABLE two (id int PRIMARY KEY);\nSELECT 1/0;\n")}
)

func version(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	var v int64
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT max(version_id) FROM goose_db_version").Scan(&v))

	return v
}

func tableExists(t *testing.T, pool *pgxpool.Pool, name string) bool {
	t.Helper()

	var exists bool
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT to_regclass($1) IS NOT NULL", name).Scan(&exists))

	return exists
}

func TestMigrateAppliesAllMigrationsToAnEmptyDatabase(t *testing.T) {
	pool := databasetest.New(t)

	require.NoError(t, database.Migrate(t.Context(), pool, fstest.MapFS{"00001_one.sql": v1, "00002_two.sql": v2}))

	assert.Equal(t, int64(2), version(t, pool))
	assert.True(t, tableExists(t, pool, "one"))
	assert.True(t, tableExists(t, pool, "two"))
}

func TestMigrateAppliesNewMigrationsToAnEarlierVersion(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, fstest.MapFS{"00001_one.sql": v1}))

	require.NoError(t, database.Migrate(t.Context(), pool, fstest.MapFS{"00001_one.sql": v1, "00002_two.sql": v2}))

	assert.Equal(t, int64(2), version(t, pool))
	assert.True(t, tableExists(t, pool, "two"))
}

func TestMigrateAgainChangesNothing(t *testing.T) {
	pool := databasetest.New(t)
	migrations := fstest.MapFS{"00001_one.sql": v1, "00002_two.sql": v2}
	require.NoError(t, database.Migrate(t.Context(), pool, migrations))

	require.NoError(t, database.Migrate(t.Context(), pool, migrations))

	assert.Equal(t, int64(2), version(t, pool))
}

func TestMigrateRollsBackAFailedMigrationAndCanTryAgain(t *testing.T) {
	pool := databasetest.New(t)

	err := database.Migrate(t.Context(), pool, fstest.MapFS{"00001_one.sql": v1, "00002_two.sql": v2Fails})
	require.Error(t, err)
	assert.Equal(t, int64(1), version(t, pool))
	assert.False(t, tableExists(t, pool, "two"))

	require.NoError(t, database.Migrate(t.Context(), pool, fstest.MapFS{"00001_one.sql": v1, "00002_two.sql": v2}))
	assert.Equal(t, int64(2), version(t, pool))
}

func TestMigrateFailsWithNoMigrations(t *testing.T) {
	pool := databasetest.New(t)

	assert.ErrorIs(t, database.Migrate(t.Context(), pool, fstest.MapFS{}), goose.ErrNoMigrations)
}

func TestMigrateOnStart(t *testing.T) {
	tests := []struct {
		value   string
		want    bool
		wantErr bool
	}{
		{value: "", want: false},
		{value: "true", want: true},
		{value: "false", want: false},
		{value: "yes", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Setenv(database.MigrateOnStartEnv, tt.value)

			got, err := database.MigrateOnStart()

			if tt.wantErr {
				assert.ErrorContains(t, err, database.MigrateOnStartEnv)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
