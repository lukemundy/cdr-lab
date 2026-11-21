package jtistore_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/jtistore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

func TestAJTIIsConsumedOneTime(t *testing.T) {
	store := jtistore.New(newPool(t))

	require.NoError(t, store.Consume(t.Context(), "jti-1"))
	require.NoError(t, store.Consume(t.Context(), "jti-2"))

	assert.ErrorIs(t, store.Consume(t.Context(), "jti-1"), app.ErrJTIUsed)
}

func TestAJTIStaysConsumedForANewStore(t *testing.T) {
	pool := newPool(t)
	require.NoError(t, jtistore.New(pool).Consume(t.Context(), "jti-1"))

	assert.ErrorIs(t, jtistore.New(pool).Consume(t.Context(), "jti-1"), app.ErrJTIUsed)
}

// Verifies: DH-0608
func TestAJTIThatIsOlderThanTheRetentionCanBeConsumedAgain(t *testing.T) {
	pool := newPool(t)
	store := jtistore.New(pool)
	require.NoError(t, store.Consume(t.Context(), "jti-1"))

	_, err := pool.Exec(t.Context(), "UPDATE consumed_jtis SET consumed_at = now() - interval '121 minutes'")
	require.NoError(t, err)

	assert.NoError(t, store.Consume(t.Context(), "jti-1"))
}

// Verifies: DH-0608, AUTH-0250
func TestAJTIStaysConsumedForTheRetention(t *testing.T) {
	pool := newPool(t)
	store := jtistore.New(pool)
	require.NoError(t, store.Consume(t.Context(), "jti-1"))

	_, err := pool.Exec(t.Context(), "UPDATE consumed_jtis SET consumed_at = now() - interval '119 minutes'")
	require.NoError(t, err)

	assert.ErrorIs(t, store.Consume(t.Context(), "jti-1"), app.ErrJTIUsed)
}

func TestTheAuthorisationServerAndTheRegistrationUseOneStore(t *testing.T) {
	store := jtistore.New(newPool(t))
	require.NoError(t, store.Consume(t.Context(), "jti-1"))

	err := jtistore.AuthServer{Store: store}.Consume(t.Context(), "jti-1")

	require.ErrorIs(t, err, authserver.ErrJTIUsed)
	assert.NoError(t, jtistore.AuthServer{Store: store}.Consume(t.Context(), "jti-2"))
}

// Verifies: DH-0620
func TestAnSSAJTIIsConsumedOneTimeUntilTheSSAExpires(t *testing.T) {
	store := jtistore.New(newPool(t))
	expiresAt := time.Now().Add(10 * time.Minute)

	require.NoError(t, store.ConsumeSSA(t.Context(), "ssa-1", expiresAt))
	require.NoError(t, store.ConsumeSSA(t.Context(), "ssa-2", expiresAt))

	assert.ErrorIs(t, store.ConsumeSSA(t.Context(), "ssa-1", expiresAt), app.ErrJTIUsed)
}

func TestAnSSAJTIIsKeptLongerThanTheRetentionOfARequestJTI(t *testing.T) {
	store := jtistore.New(newPool(t))
	require.NoError(t, store.ConsumeSSA(t.Context(), "ssa-1", time.Now().Add(3*time.Hour)))

	assert.ErrorIs(t, store.ConsumeSSA(t.Context(), "ssa-1", time.Now().Add(3*time.Hour)), app.ErrJTIUsed)
}

func TestAnSSAJTIOfAnExpiredSSACanBeConsumedAgain(t *testing.T) {
	store := jtistore.New(newPool(t))
	require.NoError(t, store.ConsumeSSA(t.Context(), "ssa-1", time.Now().Add(-time.Second)))

	assert.NoError(t, store.ConsumeSSA(t.Context(), "ssa-1", time.Now().Add(10*time.Minute)))
}

func TestTheJTIsOfSSAsAndOfOtherJWTsAreSeparate(t *testing.T) {
	store := jtistore.New(newPool(t))
	require.NoError(t, store.Consume(t.Context(), "jti-1"))

	assert.NoError(t, store.ConsumeSSA(t.Context(), "jti-1", time.Now().Add(10*time.Minute)))
}
