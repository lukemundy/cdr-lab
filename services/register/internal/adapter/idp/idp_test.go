package idp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/idp"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

type reader struct {
	client app.SoftwareProductClient
	err    error
}

func (r reader) SoftwareProductClient(context.Context, string) (app.SoftwareProductClient, error) {
	return r.client, r.err
}

// Verifies: REG-0457
func TestAClientIsActiveOnlyWhenTheSoftwareProductIsActive(t *testing.T) {
	for status, active := range map[domain.Status]bool{
		domain.Active:   true,
		domain.Inactive: false,
		domain.Removed:  false,
	} {
		t.Run(string(status), func(t *testing.T) {
			clients := idp.Clients{Reader: reader{client: app.SoftwareProductClient{
				SoftwareProductID:  "sp-1",
				Status:             status,
				JWKSURI:            "https://recipient.example/jwks",
				ClientCertificates: []string{"thumbprint"},
			}}}

			client, err := clients.Client(t.Context(), "sp-1")

			require.NoError(t, err)
			assert.Equal(t, authserver.Client{
				ID:                     "sp-1",
				JWKSURI:                "https://recipient.example/jwks",
				Active:                 active,
				CertificateThumbprints: []string{"thumbprint"},
			}, client)
		})
	}
}

func TestClientErrors(t *testing.T) {
	failure := errors.New("database failure")

	_, notFound := idp.Clients{
		Reader: reader{err: fmt.Errorf("software product %w", domain.ErrNotFound)},
	}.Client(
		t.Context(),
		"sp-1",
	)
	_, other := idp.Clients{Reader: reader{err: failure}}.Client(t.Context(), "sp-1")

	require.ErrorIs(t, notFound, authserver.ErrClientNotFound)
	require.ErrorIs(t, other, failure)
	assert.NotErrorIs(t, other, authserver.ErrClientNotFound)
}

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

func TestAJTIIsConsumedOneTime(t *testing.T) {
	store := idp.NewStore(newPool(t))

	require.NoError(t, store.Consume(t.Context(), "jti-1"))
	require.NoError(t, store.Consume(t.Context(), "jti-2"))

	assert.ErrorIs(t, store.Consume(t.Context(), "jti-1"), authserver.ErrJTIUsed)
}

// Verifies: REG-0455, AUTH-0250
func TestOldJTIsAreRemoved(t *testing.T) {
	pool := newPool(t)
	store := idp.NewStore(pool)
	_, err := pool.Exec(
		t.Context(),
		"INSERT INTO consumed_jtis (jti, consumed_at) "+
			"VALUES ('old', now() - interval '2 hours'), ('recent', now() - interval '30 minutes')",
	)
	require.NoError(t, err)

	require.NoError(t, store.Consume(t.Context(), "new"))

	var jtis []string
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT array_agg(jti ORDER BY jti) FROM consumed_jtis").Scan(&jtis))
	assert.Equal(t, []string{"new", "recent"}, jtis)
}

func TestTheSigningKeyIsCreatedOneTimeAndStaysTheSame(t *testing.T) {
	pool := newPool(t)
	keys := idp.NewStore(pool).SigningKeys(app.AccessTokenKeys)

	empty, err := keys(t.Context())
	require.NoError(t, err)
	require.NoError(t, idp.NewStore(pool).EnsureSigningKey(t.Context(), app.AccessTokenKeys))
	first, err := keys(t.Context())
	require.NoError(t, err)
	require.NoError(t, idp.NewStore(pool).EnsureSigningKey(t.Context(), app.AccessTokenKeys))
	second, err := keys(t.Context())
	require.NoError(t, err)

	assert.Empty(t, empty)
	require.Len(t, first, 1)
	assert.NotEmpty(t, first[0].ID)
	assert.GreaterOrEqual(t, first[0].Key.N.BitLen(), 2048)
	require.Len(t, second, 1)
	assert.Equal(t, first[0].ID, second[0].ID)
	assert.True(t, first[0].Key.Equal(second[0].Key))
}

// Verifies: REG-0465
func TestEachKeySetHasItsOwnKey(t *testing.T) {
	store := idp.NewStore(newPool(t))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.AccessTokenKeys))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))

	accessToken, err := store.SigningKeys(app.AccessTokenKeys)(t.Context())
	require.NoError(t, err)
	ssa, err := store.SigningKeys(app.SSAKeys)(t.Context())
	require.NoError(t, err)

	require.Len(t, accessToken, 1)
	require.Len(t, ssa, 1)
	assert.NotEqual(t, accessToken[0].ID, ssa[0].ID)
	assert.False(t, accessToken[0].Key.Equal(ssa[0].Key))
}

func keyIDs(t *testing.T, store *idp.Store, set app.KeySet) []string {
	t.Helper()
	keys, err := store.SigningKeys(set)(t.Context())
	require.NoError(t, err)

	ids := make([]string, 0, len(keys))
	for _, k := range keys {
		ids = append(ids, k.ID)
	}

	return ids
}

// Verifies: REG-0467, REG-0468
func TestAKeyChangePublishesTheNextKeyBeforeItSigns(t *testing.T) {
	pool := newPool(t)
	store := idp.NewStore(pool)
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.AccessTokenKeys))
	first := keyIDs(t, store, app.SSAKeys)[0]
	otherSet := keyIDs(t, store, app.AccessTokenKeys)

	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))
	published := keyIDs(t, store, app.SSAKeys)
	require.NoError(t, store.ActivateNextKey(t.Context(), app.SSAKeys))
	activated := keyIDs(t, store, app.SSAKeys)

	require.Len(t, published, 2)
	assert.Equal(t, first, published[0], "the first key still signs after the next key is published")
	second := published[1]
	assert.Equal(
		t,
		[]string{second, first},
		activated,
		"the next key signs after its activation, and the JWKS still has the previous key",
	)
	assert.Equal(
		t,
		activated,
		keyIDs(t, idp.NewStore(pool), app.SSAKeys),
		"a new store on the same database gives the same keys in the same roles",
	)
	assert.Equal(t, otherSet, keyIDs(t, store, app.AccessTokenKeys), "the other key set does not change")
}

// Verifies: REG-0469
func TestTheNextKeyChangeRemovesThePreviousKey(t *testing.T) {
	store := idp.NewStore(newPool(t))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))
	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))
	require.NoError(t, store.ActivateNextKey(t.Context(), app.SSAKeys))
	afterFirstChange := keyIDs(t, store, app.SSAKeys)

	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))
	published := keyIDs(t, store, app.SSAKeys)

	require.Len(t, published, 2)
	assert.Equal(t, afterFirstChange[0], published[0], "the signing key stays")
	assert.NotContains(t, published, afterFirstChange[1], "the previous key is removed")
}

func TestKeyChangeStepsThatAreNotValid(t *testing.T) {
	store := idp.NewStore(newPool(t))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))

	noNext := store.ActivateNextKey(t.Context(), app.SSAKeys)
	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))
	before := keyIDs(t, store, app.SSAKeys)
	inProgress := store.PublishNextKey(t.Context(), app.SSAKeys)

	require.ErrorIs(t, noNext, app.ErrNoNextKey)
	require.ErrorIs(t, inProgress, app.ErrKeyChangeInProgress)
	assert.Equal(t, before, keyIDs(t, store, app.SSAKeys), "a refused step changes nothing")
}

// createdAt gives the creation time of each key of the set, in the order of the JWKS.
func createdAt(t *testing.T, pool *pgxpool.Pool, set app.KeySet) []time.Time {
	t.Helper()

	var created []time.Time
	require.NoError(t, pool.QueryRow(t.Context(), `
		SELECT array_agg(created_at ORDER BY CASE status WHEN 'signing' THEN 0 WHEN 'next' THEN 1 ELSE 2 END, key_id)
		FROM signing_keys WHERE key_set = $1`, set).Scan(&created))

	return created
}

func TestKeysGiveTheStateAndTheCreationTimeOfEachKeyWithTheKeyThatSignsFirst(t *testing.T) {
	pool := newPool(t)
	store := idp.NewStore(pool)
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))
	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.AccessTokenKeys))
	require.NoError(t, store.PublishNextKey(t.Context(), app.AccessTokenKeys))
	require.NoError(t, store.ActivateNextKey(t.Context(), app.AccessTokenKeys))

	tests := map[app.KeySet][]app.KeyState{
		app.SSAKeys:         {app.KeySigning, app.KeyNext},
		app.AccessTokenKeys: {app.KeySigning, app.KeyPrevious},
	}
	for set, states := range tests {
		t.Run(string(set), func(t *testing.T) {
			ids := keyIDs(t, store, set)
			created := createdAt(t, pool, set)

			keySet, err := store.KeySet(t.Context(), set)
			require.NoError(t, err)

			keys := keySet.Keys

			require.Len(t, ids, len(states))

			want := make([]app.Key, 0, len(states))
			for i, state := range states {
				want = append(want, app.Key{ID: ids[i], State: state, CreatedAt: created[i].UTC()})
			}

			assert.Equal(t, want, keys)

			for _, k := range keys {
				assert.Equal(t, time.UTC, k.CreatedAt.Location())
			}
		})
	}
}

func TestAKeySetWithNoKeyIsEmpty(t *testing.T) {
	keySet, err := idp.NewStore(newPool(t)).KeySet(t.Context(), app.SSAKeys)

	require.NoError(t, err)
	assert.Equal(t, app.KeySetState{Keys: []app.Key{}}, keySet)
}

func version(t *testing.T, store *idp.Store, set app.KeySet) int64 {
	t.Helper()

	keySet, err := store.KeySet(t.Context(), set)
	require.NoError(t, err)

	return keySet.Version
}

// Verifies: REG-0471
func TestEachStepOfAKeyChangeChangesTheVersionOfItsKeySetOnly(t *testing.T) {
	store := idp.NewStore(newPool(t))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.AccessTokenKeys))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))

	assert.Equal(t, int64(1), version(t, store, app.SSAKeys), "a new key set is at version 1")

	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))
	assert.Equal(t, int64(2), version(t, store, app.SSAKeys))

	require.NoError(t, store.ActivateNextKey(t.Context(), app.SSAKeys))
	assert.Equal(t, int64(3), version(t, store, app.SSAKeys))
	assert.Equal(t, int64(1), version(t, store, app.AccessTokenKeys), "the other key set does not change")
}

func TestAKeySetThatExistsBeforeItHasAVersionGetsOneAtTheNextStart(t *testing.T) {
	pool := newPool(t)
	store := idp.NewStore(pool)
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))
	before := keyIDs(t, store, app.SSAKeys)
	_, err := pool.Exec(t.Context(), "DELETE FROM key_sets")
	require.NoError(t, err)

	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))

	assert.Equal(t, int64(1), version(t, store, app.SSAKeys))
	assert.Equal(t, before, keyIDs(t, store, app.SSAKeys))
}

func TestAKeyChangeAtTheCurrentVersionDoesTheStep(t *testing.T) {
	store := idp.NewStore(newPool(t))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))
	first := keyIDs(t, store, app.SSAKeys)[0]

	require.NoError(t, store.ChangeKeyAt(t.Context(), app.SSAKeys, app.PublishNext, 1))
	published := keyIDs(t, store, app.SSAKeys)
	require.NoError(t, store.ChangeKeyAt(t.Context(), app.SSAKeys, app.ActivateNext, 2))

	require.Len(t, published, 2)
	assert.Equal(t, []string{published[1], first}, keyIDs(t, store, app.SSAKeys))
	assert.Equal(t, int64(3), version(t, store, app.SSAKeys))
}

func TestAKeyChangeAtAVersionThatIsNotCurrentChangesNothing(t *testing.T) {
	store := idp.NewStore(newPool(t))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))
	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))
	before, err := store.KeySet(t.Context(), app.SSAKeys)
	require.NoError(t, err)

	for _, v := range []int64{1, 3} {
		// The rules also refuse a publish now, but the version comes first.
		for _, action := range app.KeyActions {
			err := store.ChangeKeyAt(t.Context(), app.SSAKeys, action, v)

			require.ErrorIs(t, err, app.ErrKeySetChanged, "%s at version %d", action, v)
		}
	}

	after, err := store.KeySet(t.Context(), app.SSAKeys)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestAStepThatTheRulesRefuseDoesNotChangeTheVersion(t *testing.T) {
	store := idp.NewStore(newPool(t))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))

	err := store.ChangeKeyAt(t.Context(), app.SSAKeys, app.ActivateNext, 1)

	require.ErrorIs(t, err, app.ErrNoNextKey)
	assert.Equal(t, int64(1), version(t, store, app.SSAKeys))
}

func TestOnlyOneOfTwoKeyChangesAtTheSameVersionHappens(t *testing.T) {
	store := idp.NewStore(newPool(t))
	require.NoError(t, store.EnsureSigningKey(t.Context(), app.SSAKeys))

	const changes = 4

	errs := make(chan error, changes)
	for range changes {
		go func() { errs <- store.ChangeKeyAt(t.Context(), app.SSAKeys, app.PublishNext, 1) }()
	}

	succeeded := 0

	for range changes {
		err := <-errs
		if err == nil {
			succeeded++
			continue
		}

		require.ErrorIs(t, err, app.ErrKeySetChanged)
	}

	assert.Equal(t, 1, succeeded)
	assert.Len(t, keyIDs(t, store, app.SSAKeys), 2)
	assert.Equal(t, int64(2), version(t, store, app.SSAKeys))
}
