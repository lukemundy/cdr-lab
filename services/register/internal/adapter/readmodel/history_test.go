package readmodel_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func eventsOf(t *testing.T, pool *pgxpool.Pool, streamType, streamID string) []eventsourcing.RecordedEvent {
	t.Helper()

	rows, err := pool.Query(t.Context(), `SELECT event_type, stream_version, recorded_at FROM events
		WHERE stream_type = $1 AND stream_id = $2 ORDER BY stream_version DESC`, streamType, streamID)
	require.NoError(t, err)

	var events []eventsourcing.RecordedEvent

	for rows.Next() {
		var e eventsourcing.RecordedEvent
		require.NoError(t, rows.Scan(&e.EventType, &e.StreamVersion, &e.RecordedAt))
		events = append(events, e)
	}

	require.NoError(t, rows.Err())

	return events
}

// seedSuspendedRecipient gives a stream of 4 events: the import, the suspension, its cascade to sp-1, and a client
// certificate of sp-2.
func seedSuspendedRecipient(t *testing.T) (*readmodel.Queries, []eventsourcing.RecordedEvent) {
	t.Helper()

	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{
		DataRecipients: []domain.DataRecipient{recipient("dr-1", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active), product("sp-2", domain.Inactive)})},
		DataHolders: []domain.DataHolder{holder("dh-1", domain.DataHolderActive)},
	}))
	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1", domain.SuspendAccreditation{}))

	data, err := app.DataRecipientType.Encode([]domain.DataRecipientEvent{
		domain.ClientCertificateRegistered{SoftwareProductID: "sp-2", Thumbprint: "thumbprint-1"},
	})
	require.NoError(t, err)
	require.NoError(t, store.Append(t.Context(), eventsourcing.Append{
		StreamType: app.DataRecipientType.StreamType, StreamID: "dr-1", ExpectedVersion: 3, Events: data,
	}))

	return readmodel.NewQueries(pool), eventsOf(t, pool, app.DataRecipientType.StreamType, "dr-1")
}

func TestHistoryGivesEachEventOfTheStreamNewestFirstWithOnlyTheFieldsOfTheTable(t *testing.T) {
	queries, events := seedSuspendedRecipient(t)
	require.Len(t, events, 4)

	got, total, err := queries.History(t.Context(), app.DataRecipientType.StreamType,
		app.HistoryPageQuery{LegalEntityID: "dr-1", Page: 1, PageSize: 25})

	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Equal(t, []app.HistoryEntry{
		{
			StreamVersion: 4, EventType: "ClientCertificateRegistered", RecordedAt: events[0].RecordedAt,
			SoftwareProduct: &app.HistorySoftwareProduct{ID: "sp-2", Name: "Product sp-2"}, Thumbprint: "thumbprint-1",
		},
		{
			StreamVersion: 3, EventType: "SoftwareProductDeactivated", RecordedAt: events[1].RecordedAt,
			SoftwareProduct: &app.HistorySoftwareProduct{ID: "sp-1", Name: "Product sp-1"},
		},
		{StreamVersion: 2, EventType: "AccreditationSuspended", RecordedAt: events[2].RecordedAt},
		{StreamVersion: 1, EventType: "DataRecipientImported", RecordedAt: events[3].RecordedAt},
	}, got)
	assert.True(t, got[1].RecordedAt.Equal(got[2].RecordedAt), "the events of one append have the same time")
}

func TestHistoryGivesOnePageAndTheTotal(t *testing.T) {
	queries, _ := seedSuspendedRecipient(t)

	for page, versions := range map[int][]int64{1: {4, 3, 2}, 2: {1}, 3: nil} {
		got, total, err := queries.History(t.Context(), app.DataRecipientType.StreamType,
			app.HistoryPageQuery{LegalEntityID: "dr-1", Page: page, PageSize: 3})

		require.NoError(t, err)
		assert.Equal(t, 4, total, "page %d", page)

		var gotVersions []int64
		for _, e := range got {
			gotVersions = append(gotVersions, e.StreamVersion)
		}

		assert.Equal(t, versions, gotVersions, "page %d", page)
	}
}

func TestHistoryGivesNotFoundForAnUnknownStream(t *testing.T) {
	queries, _ := seedSuspendedRecipient(t)

	for streamType, id := range map[string]string{
		app.DataRecipientType.StreamType: "dh-1",
		app.DataHolderType.StreamType:    "dr-1",
	} {
		_, _, err := queries.History(t.Context(), streamType,
			app.HistoryPageQuery{LegalEntityID: id, Page: 1, PageSize: 25})

		assert.ErrorIs(t, err, domain.ErrNotFound, streamType)
	}
}

func TestHistoryOfADataHolderGivesItsImport(t *testing.T) {
	queries, _ := seedSuspendedRecipient(t)

	got, total, err := queries.History(t.Context(), app.DataHolderType.StreamType,
		app.HistoryPageQuery{LegalEntityID: "dh-1", Page: 1, PageSize: 25})

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, got, 1)
	assert.Equal(t, "DataHolderImported", got[0].EventType)
	assert.Nil(t, got[0].SoftwareProduct)
}

// The events of a stream do not change, so the projection adds the rows of the new events only.
func TestAnAppendDoesNotWriteTheHistoryRowsOfEarlierEventsAgain(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		recipient("dr-1", domain.DataRecipientActive, []domain.SoftwareProduct{product("sp-1", domain.Active)}),
	}}))

	rowVersion := func() string {
		var xmin string
		require.NoError(t, pool.QueryRow(t.Context(), `SELECT xmin::text FROM participant_history
			WHERE legal_entity_id = 'dr-1' AND stream_version = 1`).Scan(&xmin))

		return xmin
	}
	before := rowVersion()

	require.NoError(t, app.NewOperator(store).Execute(t.Context(), "dr-1", domain.SuspendAccreditation{}))

	assert.Equal(t, before, rowVersion())

	_, total, err := readmodel.NewQueries(pool).History(t.Context(), app.DataRecipientType.StreamType,
		app.HistoryPageQuery{LegalEntityID: "dr-1", Page: 1, PageSize: 25})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
}

func TestHistoryOfADataHolderGivesTheBrandOfABrandEvent(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{
		DataHolders: []domain.DataHolder{holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking})},
	}))
	require.NoError(t, app.NewOperator(store).ExecuteDataHolder(t.Context(), "dh-1",
		domain.DeactivateDataHolderBrand{BrandID: "dh-1-b1"}))

	got, total, err := readmodel.NewQueries(pool).History(t.Context(), app.DataHolderType.StreamType,
		app.HistoryPageQuery{LegalEntityID: "dh-1", Page: 1, PageSize: 1})

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, got, 1)
	assert.Equal(t, "DataHolderBrandDeactivated", got[0].EventType)
	assert.Equal(t, &app.HistoryBrand{ID: "dh-1-b1", Name: "Brand"}, got[0].Brand)
	assert.Nil(t, got[0].SoftwareProduct)
}

func TestHistoryOfADataHolderGivesTheBrandAndTheThumbprintOfAServerCertificateIssue(t *testing.T) {
	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{
		DataHolders: []domain.DataHolder{holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking})},
	}))

	thumbprint := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	require.NoError(t, app.NewOperator(store).ExecuteDataHolder(t.Context(), "dh-1",
		domain.IssueServerCertificate{BrandID: "dh-1-b1", Thumbprint: thumbprint}))

	got, _, err := readmodel.NewQueries(pool).History(t.Context(), app.DataHolderType.StreamType,
		app.HistoryPageQuery{LegalEntityID: "dh-1", Page: 1, PageSize: 1})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "ServerCertificateIssued", got[0].EventType)
	assert.Equal(t, &app.HistoryBrand{ID: "dh-1-b1", Name: "Brand"}, got[0].Brand)
	assert.Equal(t, thumbprint, got[0].Thumbprint)
	assert.Nil(t, got[0].SoftwareProduct)
}
