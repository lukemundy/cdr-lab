package castore_test

import (
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/castore"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

var signedAt = time.Date(2026, 10, 27, 6, 30, 15, 0, time.UTC)

func crl(number int64) app.CRL {
	return app.CRL{
		Number:     number,
		ThisUpdate: signedAt.Add(time.Duration(number) * time.Hour),
		NextUpdate: signedAt.Add(time.Duration(number)*time.Hour).AddDate(0, 0, 7),
		DER:        []byte{byte(number)},
	}
}

func revocation(serial int64) app.Revocation {
	return app.Revocation{
		Serial:     new(big.Int).Lsh(big.NewInt(serial), 100),
		Subject:    "CN=App,OU=Consumer Data Right,O=App,C=AU",
		Thumbprint: "thumbprint",
		Time:       signedAt.Add(time.Duration(serial) * time.Minute),
		Reason:     ca.KeyCompromise,
	}
}

func change(c app.CAChange) func(app.CAState) (app.CAChange, error) {
	return func(app.CAState) (app.CAChange, error) { return c, nil }
}

func state(t *testing.T, store *castore.Store) app.CAState {
	t.Helper()

	var got app.CAState

	require.NoError(t, store.ChangeCA(t.Context(), func(s app.CAState) (app.CAChange, error) {
		got = s
		return app.CAChange{}, nil
	}))

	return got
}

func TestANewCAHasNoCRLAndNoRevocations(t *testing.T) {
	store := castore.New(newPool(t))

	newest, err := store.NewestCRL(t.Context())

	require.NoError(t, err)
	assert.Zero(t, newest)
	assert.Zero(t, state(t, store).CRL)
	assert.Empty(t, state(t, store).Revoked)
}

func TestTheStoreKeepsOnlyTheNewestCRL(t *testing.T) {
	store := castore.New(newPool(t))

	require.NoError(t, store.ChangeCA(t.Context(), change(app.CAChange{CRL: new(crl(1))})))
	require.NoError(t, store.ChangeCA(t.Context(), change(app.CAChange{CRL: new(crl(2))})))

	newest, err := store.NewestCRL(t.Context())
	require.NoError(t, err)
	assert.Equal(t, crl(2), newest)
	assert.Equal(t, crl(2), state(t, store).CRL)
}

func TestTheStoreKeepsEachRevocationForANewStore(t *testing.T) {
	pool := newPool(t)
	store := castore.New(pool)

	require.NoError(t, store.ChangeCA(t.Context(), change(app.CAChange{Revocation: new(revocation(2))})))
	require.NoError(
		t,
		store.ChangeCA(t.Context(), change(app.CAChange{Revocation: new(revocation(1)), CRL: new(crl(1))})),
	)

	got := state(t, castore.New(pool))
	assert.Equal(t, crl(1), got.CRL)
	assert.Equal(t, []app.Revocation{revocation(1), revocation(2)}, got.Revoked, "in the sequence of their times")
}

func TestAChangeThatFailsStoresNothing(t *testing.T) {
	store := castore.New(newPool(t))
	failure := errors.New("sign failure")

	err := store.ChangeCA(t.Context(), func(app.CAState) (app.CAChange, error) {
		return app.CAChange{Revocation: new(revocation(1)), CRL: new(crl(1))}, failure
	})

	require.ErrorIs(t, err, failure)
	assert.Zero(t, state(t, store).CRL)
	assert.Empty(t, state(t, store).Revoked)
}

func TestChangesAtTheSameTimeOccurOneAfterTheOther(t *testing.T) {
	pool := newPool(t)

	const changes = 10

	var wg sync.WaitGroup

	for range changes {
		wg.Go(func() {
			// Two instances of the Register, each with its own store, sign a CRL with the next number.
			assert.NoError(t, castore.New(pool).ChangeCA(t.Context(), func(s app.CAState) (app.CAChange, error) {
				return app.CAChange{CRL: new(crl(s.CRL.Number + 1))}, nil
			}))
		})
	}

	wg.Wait()

	assert.Equal(t, int64(changes), state(t, castore.New(pool)).CRL.Number)
}

func revoke(t *testing.T, store *castore.Store, r app.Revocation) {
	t.Helper()

	require.NoError(t, store.ChangeCA(t.Context(), change(app.CAChange{Revocation: &r})))
}

func TestTheStoreGivesTheRevocationOfASerialNumber(t *testing.T) {
	store := castore.New(newPool(t))
	revoked := revocation(1)
	revoke(t, store, revoked)

	got, found, err := store.Revocation(t.Context(), revoked.Serial)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, revoked, got)

	_, found, err = store.Revocation(t.Context(), revocation(2).Serial)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestTheStoreGivesTheRevocationsInPagesNewestFirstThenBySerialNumber(t *testing.T) {
	store := castore.New(newPool(t))
	oldest, newest := revocation(1), revocation(3)
	// The two serial numbers have the same time. As text, "f" is after "10", but 15 is before 16.
	sameTimeLow, sameTimeHigh := revocation(2), revocation(2)
	sameTimeLow.Serial, sameTimeHigh.Serial = big.NewInt(15), big.NewInt(16)

	for _, r := range []app.Revocation{oldest, sameTimeHigh, newest, sameTimeLow} {
		revoke(t, store, r)
	}

	all, total, err := store.Revocations(t.Context(), app.RevocationPageQuery{Page: 1, PageSize: 25})
	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Equal(t, []app.Revocation{newest, sameTimeLow, sameTimeHigh, oldest}, all)

	second, total, err := store.Revocations(t.Context(), app.RevocationPageQuery{Page: 2, PageSize: 3})
	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Equal(t, []app.Revocation{oldest}, second)

	after, total, err := store.Revocations(t.Context(), app.RevocationPageQuery{Page: 3, PageSize: 3})
	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Empty(t, after)
}
