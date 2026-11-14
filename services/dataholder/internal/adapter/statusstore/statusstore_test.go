package statusstore_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

func newStore(t *testing.T) *statusstore.Store {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return statusstore.New(pool, nil)
}

func TestAStoreWithNoStatusesHasNoETagAndHasNotGotTheStatuses(t *testing.T) {
	store := newStore(t)

	etag, err := store.ETag(t.Context())
	require.NoError(t, err)
	assert.Empty(t, etag)

	got, err := store.HasStatuses(t.Context())
	require.NoError(t, err)
	assert.False(t, got)

	_, found, err := store.Status(t.Context(), "software-1")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestSavedStatusesAndTheirETagStay(t *testing.T) {
	store := newStore(t)

	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive, "software-2": domain.SoftwareProductActive,
	}, `"v1"`))
	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-2": domain.SoftwareProductInactive,
	}, `"v2"`))

	etag, err := store.ETag(t.Context())
	require.NoError(t, err)
	assert.Equal(t, `"v2"`, etag)

	got, err := store.HasStatuses(t.Context())
	require.NoError(t, err)
	assert.True(t, got)

	for id, want := range map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive, "software-2": domain.SoftwareProductInactive,
	} {
		status, found, err := store.Status(t.Context(), id)
		require.NoError(t, err)
		assert.True(t, found, id)
		assert.Equal(t, want, status, id)
	}
}

func TestAResponseWithNoStatusesCountsAsTheStatuses(t *testing.T) {
	store := newStore(t)

	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{}, `"empty"`))

	got, err := store.HasStatuses(t.Context())
	require.NoError(t, err)
	assert.True(t, got, "the Register can have no software product")
}

// removals records each call of the hook, and makes the hook fail with err.
type removals struct {
	products []string
	at       []time.Time
	err      error
}

func (r *removals) hook(_ context.Context, _ pgx.Tx, softwareProductID string, removedAt time.Time) error {
	r.products, r.at = append(r.products, softwareProductID), append(r.at, removedAt)

	return r.err
}

func newStoreWithHook(t *testing.T, hook statusstore.RemovedHook) *statusstore.Store {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return statusstore.New(pool, hook)
}

func TestTheHookRunsOnlyWhenTheStatusOfASoftwareProductBecomesRemoved(t *testing.T) {
	calls := &removals{}
	store := newStoreWithHook(t, calls.hook)
	before := time.Now()

	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive, "software-2": domain.SoftwareProductRemoved,
	}, `"v1"`))
	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductRemoved, "software-2": domain.SoftwareProductRemoved,
	}, `"v2"`))
	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductRemoved, "software-2": domain.SoftwareProductRemoved,
	}, `"v3"`))

	assert.Equal(t, []string{"software-2", "software-1"}, calls.products, "a status that stays REMOVED gives no call")

	for _, at := range calls.at {
		assert.WithinRange(t, at, before, time.Now())
	}
}

func TestAnErrorOfTheHookUndoesTheSave(t *testing.T) {
	calls := &removals{err: errors.New("insert failed")}
	store := newStoreWithHook(t, calls.hook)

	err := store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductRemoved,
	}, `"v1"`)

	require.ErrorIs(t, err, calls.err)

	_, found, err := store.Status(t.Context(), "software-1")
	require.NoError(t, err)
	assert.False(t, found)

	etag, err := store.ETag(t.Context())
	require.NoError(t, err)
	assert.Empty(t, etag, "the next request gets the statuses again")
}

func TestAStoreWithNoStatusesGivesAnEmptyListWithNoTime(t *testing.T) {
	store := newStore(t)

	list, err := store.StatusList(t.Context(), byName)
	require.NoError(t, err)
	assert.Empty(t, list.Products)
	assert.Zero(t, list.ReceivedAt)
}

func TestTheListGivesEachStatusAndTheTimeOfTheLastResponse(t *testing.T) {
	store := newStore(t)

	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-2": domain.SoftwareProductActive, "software-1": domain.SoftwareProductActive,
	}, `"v1"`))

	before := time.Now()

	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-2": domain.SoftwareProductInactive,
	}, `"v2"`))

	list, err := store.StatusList(t.Context(), byName)
	require.NoError(t, err)
	assert.Equal(t, []app.ProductStatus{
		{SoftwareProductID: "software-1", Status: domain.SoftwareProductActive},
		{SoftwareProductID: "software-2", Status: domain.SoftwareProductInactive},
	}, list.Products)
	assert.WithinRange(t, list.ReceivedAt, before.Truncate(time.Microsecond), time.Now())
}

var byName = app.StatusListQuery{Sort: app.SortStatusesByName}

// newStoreWithRegistrations has the statuses of software-1 to software-5. Each of software-1 to software-4 has a
// registration, and software-5 has none.
func newStoreWithRegistrations(t *testing.T) *statusstore.Store {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	events := eventstore.New(pool, registrationstore.Projections()...)
	for softwareID, name := range map[string]string{
		"software-1": "beta", "software-2": "Alpha", "software-3": "alpha", "software-4": "Gamma",
	} {
		require.NoError(t, app.ClientRegistrationType.Execute(t.Context(), events, softwareID, domain.RegisterClient{
			Registration: domain.Registration{
				ClientID:   "client-" + softwareID,
				SoftwareID: softwareID,
				ClientName: name,
			},
		}))
	}

	store := statusstore.New(pool, nil)
	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive, "software-2": domain.SoftwareProductRemoved,
		"software-3": domain.SoftwareProductInactive, "software-4": domain.SoftwareProductActive,
		"software-5": domain.SoftwareProductActive,
	}, `"v1"`))

	return store
}

func softwareIDs(t *testing.T, store *statusstore.Store, query app.StatusListQuery) []string {
	t.Helper()

	list, err := store.StatusList(t.Context(), query)
	require.NoError(t, err)

	ids := make([]string, 0, len(list.Products))
	for _, p := range list.Products {
		ids = append(ids, p.SoftwareProductID)
	}

	return ids
}

func TestEachStatusHasTheClientIDAndTheClientNameOfItsRegistration(t *testing.T) {
	store := newStoreWithRegistrations(t)

	list, err := store.StatusList(t.Context(), byName)
	require.NoError(t, err)
	assert.Equal(t, []app.ProductStatus{
		{
			SoftwareProductID: "software-2", Status: domain.SoftwareProductRemoved,
			ClientID: "client-software-2", ClientName: "Alpha",
		},
		{
			SoftwareProductID: "software-3", Status: domain.SoftwareProductInactive,
			ClientID: "client-software-3", ClientName: "alpha",
		},
		{
			SoftwareProductID: "software-1", Status: domain.SoftwareProductActive,
			ClientID: "client-software-1", ClientName: "beta",
		},
		{
			SoftwareProductID: "software-4", Status: domain.SoftwareProductActive,
			ClientID: "client-software-4", ClientName: "Gamma",
		},
		{SoftwareProductID: "software-5", Status: domain.SoftwareProductActive},
	}, list.Products)
}

func TestTheNameSortHasNoDifferenceForCaseAndPutsNoRegistrationLastInBothDirections(t *testing.T) {
	store := newStoreWithRegistrations(t)

	assert.Equal(t, []string{"software-4", "software-1", "software-2", "software-3", "software-5"},
		softwareIDs(t, store, app.StatusListQuery{Sort: app.SortStatusesByName, Descending: true}),
		"the software ID sorts the same names in the ascending order")
}

func TestTheStatusSortUsesTheSoftwareIDForTheSameStatus(t *testing.T) {
	store := newStoreWithRegistrations(t)

	assert.Equal(t, []string{"software-1", "software-4", "software-5", "software-3", "software-2"},
		softwareIDs(t, store, app.StatusListQuery{Sort: app.SortStatusesByStatus}))
	assert.Equal(t, []string{"software-2", "software-3", "software-1", "software-4", "software-5"},
		softwareIDs(t, store, app.StatusListQuery{Sort: app.SortStatusesByStatus, Descending: true}))
}

func TestTheListIsFilteredByTheRegistration(t *testing.T) {
	store := newStoreWithRegistrations(t)
	registered, notRegistered := true, false

	assert.Equal(t, []string{"software-2", "software-3", "software-1", "software-4"},
		softwareIDs(t, store, app.StatusListQuery{Sort: app.SortStatusesByName, Registered: &registered}))
	assert.Equal(t, []string{"software-5"},
		softwareIDs(t, store, app.StatusListQuery{Sort: app.SortStatusesByName, Registered: &notRegistered}))
}

func TestAnUnknownSortIsAnError(t *testing.T) {
	store := newStoreWithRegistrations(t)

	_, err := store.StatusList(t.Context(), app.StatusListQuery{Sort: "zebracorn"})
	require.Error(t, err)
}

func TestANotModifiedResponseChangesOnlyTheTimeOfTheLastResponse(t *testing.T) {
	store := newStore(t)

	require.NoError(t, store.Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive,
	}, `"v1"`))

	saved, err := store.StatusList(t.Context(), byName)
	require.NoError(t, err)

	before := time.Now()

	require.NoError(t, store.SaveNotModified(t.Context()))

	list, err := store.StatusList(t.Context(), byName)
	require.NoError(t, err)
	assert.Equal(t, saved.Products, list.Products)
	assert.WithinRange(t, list.ReceivedAt, before.Truncate(time.Microsecond), time.Now())

	etag, err := store.ETag(t.Context())
	require.NoError(t, err)
	assert.Equal(t, `"v1"`, etag)
}

func TestANotModifiedResponseBeforeTheStatusesDoesNotCountAsTheStatuses(t *testing.T) {
	store := newStore(t)

	require.NoError(t, store.SaveNotModified(t.Context()))

	got, err := store.HasStatuses(t.Context())
	require.NoError(t, err)
	assert.False(t, got, "the readiness probe still fails")
}
