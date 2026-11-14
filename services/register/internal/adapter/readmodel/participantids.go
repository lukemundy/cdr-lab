package readmodel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// projectParticipantIDs gives a domain.IDNotUniqueError when a different stream has an ID of the stream. The insert
// waits for a concurrent append with the same ID, so that only one of the two appends gets it.
func projectParticipantIDs(streamType string, idsOf func(streamID string, recorded []eventsourcing.RecordedEvent) (
	[]string, error,
),
) func(context.Context, pgx.Tx, string, []eventsourcing.RecordedEvent) error {
	return func(ctx context.Context, tx pgx.Tx, streamID string, recorded []eventsourcing.RecordedEvent) error {
		ids, err := idsOf(streamID, recorded)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `DELETE FROM participant_ids WHERE stream_type = $1 AND legal_entity_id = $2`,
			streamType, streamID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `INSERT INTO participant_ids (id, stream_type, legal_entity_id)
			SELECT unnest($1::text[]), $2, $3 ON CONFLICT (id) DO NOTHING`, ids, streamType, streamID); err != nil {
			return err
		}

		var taken string

		err = tx.QueryRow(ctx, `SELECT id FROM participant_ids
			WHERE id = ANY ($1) AND (stream_type, legal_entity_id) <> ($2, $3) ORDER BY id LIMIT 1`,
			ids, streamType, streamID).Scan(&taken)

		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil
		case err != nil:
			return err
		default:
			return domain.IDNotUniqueError{ID: taken}
		}
	}
}
