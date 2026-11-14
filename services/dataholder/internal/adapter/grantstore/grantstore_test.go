package grantstore_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/grantstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

var confirmedAt = time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)

// grant gives a grant after the consent and before the code exchange.
func grant(id string) authserver.Grant {
	return authserver.Grant{
		ID:              id,
		ClientID:        "client-1",
		Subject:         "0e0d7ac1-2a0f-4f4c-9c0c-6f2f6c7d3b11",
		Scopes:          []string{"openid", "bank:accounts.basic:read"},
		ArrangementID:   "arrangement-" + id,
		SharingDuration: 90 * 24 * time.Hour,
		AuthenticatedAt: confirmedAt.Add(-time.Minute),
		SharingEndsAt:   confirmedAt.Add(90 * 24 * time.Hour),
		CreatedAt:       confirmedAt,
		Code:            "code-" + id,
		CodeExpiresAt:   confirmedAt.Add(time.Minute),
		RedirectURI:     "https://recipient.example/callback",
		Nonce:           "nonce-1",
		CodeChallenge:   "challenge-1",
		RefreshToken:    "refresh-" + id,
	}
}

func TestAGrantIsFoundByItsIDItsCodeAndItsRefreshToken(t *testing.T) {
	store := grantstore.New(newPool(t))
	require.NoError(t, store.SaveGrant(t.Context(), grant("1")))
	require.NoError(t, store.SaveGrant(t.Context(), grant("2")))

	byID, err := store.Grant(t.Context(), "1")
	require.NoError(t, err)
	byCode, err := store.GrantByCode(t.Context(), "code-1")
	require.NoError(t, err)
	byRefreshToken, err := store.GrantByRefreshToken(t.Context(), "refresh-1")
	require.NoError(t, err)

	assert.Equal(t, grant("1"), byID)
	assert.Equal(t, grant("1"), byCode)
	assert.Equal(t, grant("1"), byRefreshToken)
}

func TestAGrantThatDoesNotExistIsNotFound(t *testing.T) {
	store := grantstore.New(newPool(t))
	require.NoError(t, store.SaveGrant(t.Context(), grant("1")))

	_, byID := store.Grant(t.Context(), "2")
	_, byCode := store.GrantByCode(t.Context(), "refresh-1")
	_, byRefreshToken := store.GrantByRefreshToken(t.Context(), "code-1")

	require.ErrorIs(t, byID, authserver.ErrGrantNotFound)
	require.ErrorIs(t, byCode, authserver.ErrGrantNotFound)
	require.ErrorIs(t, byRefreshToken, authserver.ErrGrantNotFound)
}

func TestAGrantWithNoRefreshTokenIsNotFoundByAnEmptyRefreshToken(t *testing.T) {
	store := grantstore.New(newPool(t))

	for _, id := range []string{"1", "2"} {
		oneCollection := grant(id)
		oneCollection.RefreshToken, oneCollection.SharingDuration = "", 0
		require.NoError(t, store.SaveGrant(t.Context(), oneCollection))
	}

	_, err := store.GrantByRefreshToken(t.Context(), "")

	require.ErrorIs(t, err, authserver.ErrGrantNotFound)
}

func TestASecondSaveReplacesTheStateOfTheGrant(t *testing.T) {
	pool := newPool(t)
	exchanged := grant("1")
	require.NoError(t, grantstore.New(pool).SaveGrant(t.Context(), exchanged))

	exchanged.CodeConsumedAt = confirmedAt.Add(10 * time.Second)
	exchanged.CertificateThumbprint = "thumbprint-1"
	exchanged.RevokedAt = confirmedAt.Add(20 * time.Second)
	require.NoError(t, grantstore.New(pool).SaveGrant(t.Context(), exchanged))

	saved, err := grantstore.New(pool).GrantByRefreshToken(t.Context(), "refresh-1")
	require.NoError(t, err)
	assert.Equal(t, exchanged, saved)
}

func TestTwoGrantsCannotHaveOneCodeOrOneRefreshToken(t *testing.T) {
	store := grantstore.New(newPool(t))
	require.NoError(t, store.SaveGrant(t.Context(), grant("1")))

	sameCode := grant("2")
	sameCode.Code = "code-1"
	sameRefreshToken := grant("3")
	sameRefreshToken.RefreshToken = "refresh-1"

	require.Error(t, store.SaveGrant(t.Context(), sameCode))
	require.Error(t, store.SaveGrant(t.Context(), sameRefreshToken))
}

func TestARevocationOfAnArrangementRevokesOnlyItsGrant(t *testing.T) {
	store := grantstore.New(newPool(t))
	require.NoError(t, store.SaveGrant(t.Context(), grant("1")))
	require.NoError(t, store.SaveGrant(t.Context(), grant("2")))

	require.NoError(t, store.RevokeArrangement(t.Context(), "arrangement-1"))

	revoked, err := store.Grant(t.Context(), "1")
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), revoked.RevokedAt, time.Minute)

	other, err := store.Grant(t.Context(), "2")
	require.NoError(t, err)
	assert.True(t, other.RevokedAt.IsZero())
}

func TestARevocationOfAnArrangementKeepsAnEarlierRevocation(t *testing.T) {
	store := grantstore.New(newPool(t))
	earlier := grant("1")
	earlier.RevokedAt = confirmedAt.Add(time.Minute)
	require.NoError(t, store.SaveGrant(t.Context(), earlier))

	require.NoError(t, store.RevokeArrangement(t.Context(), "arrangement-1"))

	got, err := store.Grant(t.Context(), "1")
	require.NoError(t, err)
	assert.Equal(t, earlier.RevokedAt, got.RevokedAt)
}
