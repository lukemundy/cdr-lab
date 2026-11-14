package readmodel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func TestDataRecipientGivesTheFullDataRecipientAndTheTimeOfItsLastEvent(t *testing.T) {
	store, pool := newStore(t)
	dr := recipient("dr-1", domain.DataRecipientActive,
		[]domain.SoftwareProduct{product("sp-1", domain.Active), product("sp-2", domain.Removed)},
		[]domain.SoftwareProduct{product("sp-3", domain.Inactive)})
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		dr, recipient("dr-2", domain.DataRecipientRevoked),
	}}))

	got, err := readmodel.NewQueries(pool).DataRecipient(t.Context(), "dr-1")

	require.NoError(t, err)
	assert.Equal(t, dr, got.DataRecipient)
	assert.True(t, got.LastUpdated.Equal(lastRecordedAt(t, pool, "dr-1")))
}

func TestDataRecipientGivesTheStreamVersionOfEachChange(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}}))

	queries := readmodel.NewQueries(pool)

	seeded, err := queries.DataRecipient(t.Context(), "dr-1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), seeded.Version)

	// A suspension and its cascade are two events.
	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1", domain.SuspendAccreditation{}))

	suspended, err := queries.DataRecipient(t.Context(), "dr-1")
	require.NoError(t, err)
	assert.Equal(t, int64(3), suspended.Version)
}

func TestDataRecipientGivesNotFoundForAnUnknownLegalEntity(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{
		DataRecipients: []domain.DataRecipient{recipient("dr-1", domain.DataRecipientActive)},
		DataHolders:    []domain.DataHolder{holder("dh-1", domain.DataHolderActive)},
	}))

	for _, id := range []string{"dr-unknown", "dh-1"} {
		_, err := readmodel.NewQueries(pool).DataRecipient(t.Context(), id)

		assert.ErrorIs(t, err, domain.ErrNotFound, id)
	}
}

func TestDataHolderGivesTheFullDataHolderAndTheTimeOfItsLastEvent(t *testing.T) {
	store, pool := newStore(t)
	dh := holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking}, []domain.Industry{domain.Energy})
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{
		dh, holder("dh-2", domain.DataHolderRemoved),
	}}))

	got, err := readmodel.NewQueries(pool).DataHolder(t.Context(), "dh-1")

	require.NoError(t, err)
	assert.Equal(t, dh, got.DataHolder)
	assert.True(t, got.LastUpdated.Equal(lastRecordedAt(t, pool, "dh-1")))
}

func TestDataHolderGivesTheStreamVersionOfEachChange(t *testing.T) {
	store, pool := newStore(t)
	dh := holder("dh-1", domain.DataHolderActive)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{dh}}))

	queries := readmodel.NewQueries(pool)

	seeded, err := queries.DataHolder(t.Context(), "dh-1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), seeded.Version)

	require.NoError(t, app.NewOperator(store).ExecuteDataHolder(t.Context(), "dh-1", domain.ChangeDataHolderDetails{
		Details: domain.DataHolderDetails{LegalEntityName: "Changed", LogoURI: dh.LogoURI},
	}))

	changed, err := queries.DataHolder(t.Context(), "dh-1")
	require.NoError(t, err)
	assert.Equal(t, int64(2), changed.Version)

	page, _, err := queries.DataHolderPage(t.Context(), holderPageQuery(func(*app.DataHolderPageQuery) {}))
	require.NoError(t, err)
	assert.Equal(t, int64(2), page[0].Version)
}

func TestDataHolderGivesNotFoundForAnUnknownLegalEntity(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{
		DataRecipients: []domain.DataRecipient{recipient("dr-1", domain.DataRecipientActive)},
		DataHolders:    []domain.DataHolder{holder("dh-1", domain.DataHolderActive)},
	}))

	for _, id := range []string{"dh-unknown", "dr-1"} {
		_, err := readmodel.NewQueries(pool).DataHolder(t.Context(), id)

		assert.ErrorIs(t, err, domain.ErrNotFound, id)
	}
}
