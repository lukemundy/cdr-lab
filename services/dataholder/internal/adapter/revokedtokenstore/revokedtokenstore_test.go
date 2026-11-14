package revokedtokenstore_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/revokedtokenstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

func revoked(t *testing.T, store *revokedtokenstore.Store, id string) bool {
	t.Helper()

	ok, err := store.TokenRevoked(t.Context(), id)
	require.NoError(t, err)

	return ok
}

func TestOnlyARevokedTokenIsRevoked(t *testing.T) {
	store := revokedtokenstore.New(newPool(t))

	require.NoError(t, store.RevokeToken(t.Context(), "token-1", time.Now().Add(time.Minute)))

	assert.True(t, revoked(t, store, "token-1"))
	assert.False(t, revoked(t, store, "token-2"))
}

func TestARevocationOfARevokedTokenIsNotAnError(t *testing.T) {
	store := revokedtokenstore.New(newPool(t))
	expiry := time.Now().Add(time.Minute)

	require.NoError(t, store.RevokeToken(t.Context(), "token-1", expiry))
	require.NoError(t, store.RevokeToken(t.Context(), "token-1", expiry))

	assert.True(t, revoked(t, store, "token-1"))
}

func TestARevokedTokenStaysRevokedForANewStore(t *testing.T) {
	pool := newPool(t)
	require.NoError(t, revokedtokenstore.New(pool).RevokeToken(t.Context(), "token-1", time.Now().Add(time.Minute)))

	assert.True(t, revoked(t, revokedtokenstore.New(pool), "token-1"))
}

func TestARevocationRemovesTheTokensThatExpired(t *testing.T) {
	pool := newPool(t)
	store := revokedtokenstore.New(pool)
	require.NoError(t, store.RevokeToken(t.Context(), "token-1", time.Now().Add(-time.Second)))

	require.NoError(t, store.RevokeToken(t.Context(), "token-2", time.Now().Add(time.Minute)))

	var count int
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT count(*) FROM revoked_access_tokens").Scan(&count))
	assert.Equal(t, 1, count)
}
