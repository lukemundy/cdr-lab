package signingkeystore_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/signingkeystore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

func newStore(t *testing.T) *signingkeystore.Store {
	t.Helper()

	return signingkeystore.New(newPool(t))
}

func TestADataHolderThatDidNotStartHasNoSigningKeys(t *testing.T) {
	keys, err := newStore(t).SigningKeys(t.Context())

	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestEnsureSigningKeyCreatesOneKeyAndKeepsItAtTheNextStart(t *testing.T) {
	store := newStore(t)

	require.NoError(t, store.EnsureSigningKey(t.Context()))
	first, err := store.SigningKeys(t.Context())
	require.NoError(t, err)
	require.Len(t, first, 1)

	require.NoError(t, store.EnsureSigningKey(t.Context()))
	second, err := store.SigningKeys(t.Context())
	require.NoError(t, err)

	require.Len(t, second, 1)
	assert.Equal(t, first[0].ID, second[0].ID)
	assert.True(t, first[0].Key.Equal(second[0].Key))
}

// The test holds the lock until all instances wait for it, because the race occurs only when they wait together.
func TestInstancesThatStartTogetherCreateOneKey(t *testing.T) {
	const instances = 4

	pool := newPool(t)
	store := signingkeystore.New(pool)

	holder, err := pgx.ConnectConfig(t.Context(), pool.Config().ConnConfig.Copy())
	require.NoError(t, err)
	t.Cleanup(func() { _ = holder.Close(context.WithoutCancel(t.Context())) })

	_, err = holder.Exec(t.Context(), "SELECT pg_advisory_lock(hashtext('signing_keys'))")
	require.NoError(t, err)

	var starts sync.WaitGroup
	for range instances {
		starts.Go(func() { assert.NoError(t, store.EnsureSigningKey(t.Context())) })
	}

	require.Eventually(t, func() bool {
		var waiting int

		err := holder.QueryRow(t.Context(), `
			SELECT count(*) FROM pg_locks
			WHERE locktype = 'advisory' AND NOT granted
			AND database = (SELECT oid FROM pg_database WHERE datname = current_database())`).Scan(&waiting)

		return err == nil && waiting == instances
	}, 10*time.Second, 10*time.Millisecond, "the instances did not all wait for the lock")

	_, err = holder.Exec(t.Context(), "SELECT pg_advisory_unlock(hashtext('signing_keys'))")
	require.NoError(t, err)

	starts.Wait()

	keys, err := store.SigningKeys(t.Context())
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}
