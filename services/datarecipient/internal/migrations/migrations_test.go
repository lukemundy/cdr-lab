package migrations_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/datarecipient/internal/migrations"
)

const applicationRole = "datarecipient_app"

// Verifies: DR-0604
func TestMigrationsGiveTheApplicationRoleDataAccessOnly(t *testing.T) {
	// The role is global to the server, thus the migrations of the second database find it.
	for range 2 {
		pool := databasetest.New(t)
		require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

		assert.Empty(t, databasetest.ApplicationRoleProblems(t, pool, applicationRole))
	}
}

func TestTheApplicationRoleGetsEachTableOfALaterMigration(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))
	_, err := pool.Exec(t.Context(), "CREATE TABLE later (id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY)")
	require.NoError(t, err)

	_, err = databasetest.AsRole(t, pool, applicationRole).Exec(t.Context(), "INSERT INTO later DEFAULT VALUES")

	require.NoError(t, err)
}
