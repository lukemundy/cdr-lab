package disclosurestore_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/disclosurestore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

func newStore(t *testing.T) *disclosurestore.Store {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return disclosurestore.New(pool)
}

func TestTheFirstAndTheLastDisclosureOfEachArrangementAndScope(t *testing.T) {
	store := newStore(t)
	at := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

	for _, d := range []struct {
		arrangementID, scope string
		at                   time.Time
	}{
		{"arrangement-1", "bank:accounts.basic:read", at.Add(time.Hour)},
		{"arrangement-1", "bank:accounts.basic:read", at},
		{"arrangement-1", "bank:accounts.basic:read", at.Add(30 * time.Minute)},
		{"arrangement-1", "profile", at.Add(2 * time.Hour)},
		{"arrangement-2", "profile", at},
		{"arrangement-3", "profile", at},
	} {
		require.NoError(t, store.AddDisclosure(t.Context(), d.arrangementID, d.scope, d.at))
	}

	got, err := store.FirstAndLastDisclosures(t.Context(), []string{"arrangement-1", "arrangement-2", "arrangement-4"})

	require.NoError(t, err)
	assert.Equal(t, map[app.DisclosureKey]app.DisclosurePeriod{
		{ArrangementID: "arrangement-1", Scope: "bank:accounts.basic:read"}: {First: at, Last: at.Add(time.Hour)},
		{ArrangementID: "arrangement-1", Scope: "profile"}: {
			First: at.Add(2 * time.Hour),
			Last:  at.Add(2 * time.Hour),
		},
		{ArrangementID: "arrangement-2", Scope: "profile"}: {First: at, Last: at},
	}, got)
}

func TestAnArrangementWithNoDisclosureHasNoEntry(t *testing.T) {
	got, err := newStore(t).FirstAndLastDisclosures(t.Context(), []string{"arrangement-1"})

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestTheTimesAreInUTC(t *testing.T) {
	store := newStore(t)
	perth := time.FixedZone("AWST", 8*60*60)
	at := time.Date(2026, 11, 2, 17, 0, 0, 0, perth)
	require.NoError(t, store.AddDisclosure(t.Context(), "arrangement-1", "profile", at))

	got, err := store.FirstAndLastDisclosures(t.Context(), []string{"arrangement-1"})

	require.NoError(t, err)

	period := got[app.DisclosureKey{ArrangementID: "arrangement-1", Scope: "profile"}]
	assert.Equal(t, time.UTC, period.First.Location())
	assert.True(t, at.Equal(period.Last))
}
