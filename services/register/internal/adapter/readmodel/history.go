package readmodel

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var _ app.HistoryReader = (*Queries)(nil)

type historyOf func(streamID string, recorded []eventsourcing.RecordedEvent) ([]app.HistoryEntry, error)

// projectHistory adds a row only for each event after the last row of the stream, because the events of a stream do
// not change. The projection gets the full stream, because a name in an entry comes from the state.
func projectHistory(streamType string, entriesOf historyOf) func(
	context.Context, pgx.Tx, string, []eventsourcing.RecordedEvent,
) error {
	return func(ctx context.Context, tx pgx.Tx, streamID string, recorded []eventsourcing.RecordedEvent) error {
		var projected int64
		if err := tx.QueryRow(ctx, `SELECT coalesce(max(stream_version), 0) FROM participant_history
			WHERE stream_type = $1 AND legal_entity_id = $2`, streamType, streamID).Scan(&projected); err != nil {
			return err
		}

		entries, err := entriesOf(streamID, recorded)
		if err != nil {
			return err
		}

		var rows [][]any

		for _, e := range entries {
			if e.StreamVersion > projected {
				rows = append(rows, historyColumns(streamType, streamID, e))
			}
		}

		_, err = tx.CopyFrom(ctx, pgx.Identifier{"participant_history"}, []string{
			"stream_type", "legal_entity_id", "stream_version", "event_type", "recorded_at",
			"software_product_id", "software_product_name", "thumbprint", "brand_id", "brand_name",
		}, pgx.CopyFromRows(rows))

		return err
	}
}

func historyColumns(streamType, streamID string, e app.HistoryEntry) []any {
	var id, name, thumbprint *string
	if e.SoftwareProduct != nil {
		id, name = &e.SoftwareProduct.ID, &e.SoftwareProduct.Name
	}

	if e.Thumbprint != "" {
		thumbprint = &e.Thumbprint
	}

	var brandID, brandName *string
	if e.Brand != nil {
		brandID, brandName = &e.Brand.ID, &e.Brand.Name
	}

	return []any{
		streamType, streamID, e.StreamVersion, e.EventType, e.RecordedAt, id, name, thumbprint, brandID, brandName,
	}
}

type historyRow struct {
	StreamVersion       int64
	EventType           string
	RecordedAt          time.Time
	SoftwareProductID   *string
	SoftwareProductName *string
	Thumbprint          *string
	BrandID             *string
	BrandName           *string
}

func (q *Queries) History(
	ctx context.Context,
	streamType string,
	query app.HistoryPageQuery,
) ([]app.HistoryEntry, int, error) {
	const from = "FROM participant_history WHERE stream_type = $1 AND legal_entity_id = $2"

	var total int
	if err := q.pool.QueryRow(ctx, "SELECT count(*) "+from, streamType, query.LegalEntityID).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return nil, 0, fmt.Errorf("history %w", domain.ErrNotFound)
	}

	rows, err := q.pool.Query(ctx, `SELECT stream_version, event_type, recorded_at,
			software_product_id, software_product_name, thumbprint, brand_id, brand_name `+from+`
		ORDER BY stream_version DESC LIMIT $3 OFFSET $4`,
		streamType, query.LegalEntityID, query.PageSize, (query.Page-1)*query.PageSize)
	if err != nil {
		return nil, 0, err
	}

	found, err := pgx.CollectRows(rows, pgx.RowToStructByPos[historyRow])
	if err != nil {
		return nil, 0, err
	}

	page := make([]app.HistoryEntry, 0, len(found))
	for _, r := range found {
		entry := app.HistoryEntry{StreamVersion: r.StreamVersion, EventType: r.EventType, RecordedAt: r.RecordedAt}
		if r.SoftwareProductID != nil && r.SoftwareProductName != nil {
			entry.SoftwareProduct = &app.HistorySoftwareProduct{ID: *r.SoftwareProductID, Name: *r.SoftwareProductName}
		}

		if r.Thumbprint != nil {
			entry.Thumbprint = *r.Thumbprint
		}

		if r.BrandID != nil && r.BrandName != nil {
			entry.Brand = &app.HistoryBrand{ID: *r.BrandID, Name: *r.BrandName}
		}

		page = append(page, entry)
	}

	return page, total, nil
}
