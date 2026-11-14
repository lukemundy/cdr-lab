// Package eventstore keeps streams in PostgreSQL.
package eventstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
)

const (
	uniqueViolation     = "23505"
	streamVersionUnique = "events_stream_version_key"

	scope          = "github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	streamCountKey = "app.projection.stream.count"
	eventCountKey  = "app.projection.event.count"
)

type Store struct {
	pool        *pgxpool.Pool
	projections []Projection
}

// Projection updates a read model in the append transaction, so that the read model is correct when the append commits.
type Projection struct {
	StreamType string
	// Project gets all the events of one stream that the append changed.
	Project func(ctx context.Context, tx pgx.Tx, streamID string, recorded []eventsourcing.RecordedEvent) error
}

var _ eventsourcing.Store = (*Store)(nil)

func New(pool *pgxpool.Pool, projections ...Projection) *Store {
	return &Store{pool: pool, projections: projections}
}

func (s *Store) Read(ctx context.Context, streamType, streamID string) ([]eventsourcing.RecordedEvent, error) {
	return read(ctx, s.pool, streamType, streamID)
}

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func read(ctx context.Context, q querier, streamType, streamID string) ([]eventsourcing.RecordedEvent, error) {
	rows, err := q.Query(ctx, `
		SELECT event_type, schema_version, payload, stream_version, global_position, recorded_at
		FROM events
		WHERE stream_type = $1 AND stream_id = $2
		ORDER BY stream_version`, streamType, streamID)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (eventsourcing.RecordedEvent, error) {
		var e eventsourcing.RecordedEvent

		err := row.Scan(&e.EventType, &e.SchemaVersion, &e.Payload, &e.StreamVersion, &e.GlobalPosition, &e.RecordedAt)

		return e, err
	})
}

func (s *Store) Append(ctx context.Context, appends ...eventsourcing.Append) error {
	metadata, err := metadataFrom(ctx)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	var rows [][]any

	for _, a := range appends {
		var version int64
		if err := tx.QueryRow(ctx, `
			SELECT coalesce(max(stream_version), 0) FROM events WHERE stream_type = $1 AND stream_id = $2`,
			a.StreamType, a.StreamID).Scan(&version); err != nil {
			return err
		}

		if version != a.ExpectedVersion {
			return fmt.Errorf(
				"%s %s: version %d, expected %d: %w",
				a.StreamType,
				a.StreamID,
				version,
				a.ExpectedVersion,
				eventsourcing.ErrVersionConflict,
			)
		}

		for i, e := range a.Events {
			rows = append(
				rows,
				[]any{
					a.StreamType,
					a.StreamID,
					version + int64(i) + 1,
					e.EventType,
					e.SchemaVersion,
					e.Payload,
					metadata,
				},
			)
		}
	}

	// The unique constraint rejects a concurrent append that adds the same stream version after the check.
	_, err = tx.CopyFrom(ctx, pgx.Identifier{"events"},
		[]string{"stream_type", "stream_id", "stream_version", "event_type", "schema_version", "payload", "metadata"},
		pgx.CopyFromRows(rows))
	if err != nil {
		return conflictOr(err)
	}

	if err := s.project(ctx, tx, appends); err != nil {
		return err
	}

	return conflictOr(tx.Commit(ctx))
}

func (s *Store) project(ctx context.Context, tx pgx.Tx, appends []eventsourcing.Append) error {
	for _, a := range appends {
		for _, p := range s.projections {
			if p.StreamType != a.StreamType {
				continue
			}

			if _, err := projectStream(ctx, tx, p, a.StreamID); err != nil {
				return err
			}
		}
	}

	return nil
}

// projectStream gives the number of events of the stream.
func projectStream(ctx context.Context, tx pgx.Tx, p Projection, streamID string) (int, error) {
	recorded, err := read(ctx, tx, p.StreamType, streamID)
	if err != nil {
		return 0, err
	}

	if err := p.Project(ctx, tx, streamID, recorded); err != nil {
		return 0, fmt.Errorf("project %s %s: %w", p.StreamType, streamID, err)
	}

	return len(recorded), nil
}

// ProjectAll runs each projection for each stream of its stream type, so that a new read model is not empty.
func (s *Store) ProjectAll(ctx context.Context) error {
	ctx, span := otel.Tracer(scope).Start(ctx, "project all streams")
	defer span.End()

	streams, events, err := s.projectAll(ctx)
	if err != nil {
		// No description: errors can contain query arguments.
		span.SetStatus(codes.Error, "")
		span.SetAttributes(semconv.ErrorType(err))

		return err
	}

	span.SetAttributes(attribute.Int(streamCountKey, streams), attribute.Int(eventCountKey, events))
	slog.InfoContext(ctx, "read models projected", streamCountKey, streams, eventCountKey, events)

	return nil
}

func (s *Store) projectAll(ctx context.Context) (streams, events int, err error) {
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		// The lock blocks appends, so that no append changes a stream between the read and the projection.
		// It needs more than SELECT and INSERT on events, thus only the owner role can run ProjectAll.
		if _, err := tx.Exec(ctx, "LOCK TABLE events IN SHARE MODE"); err != nil {
			return err
		}

		for _, projection := range s.projections {
			rows, err := tx.Query(
				ctx,
				"SELECT DISTINCT stream_id FROM events WHERE stream_type = $1 ORDER BY stream_id",
				projection.StreamType,
			)
			if err != nil {
				return err
			}

			streamIDs, err := pgx.CollectRows(rows, pgx.RowTo[string])
			if err != nil {
				return err
			}

			for _, streamID := range streamIDs {
				projected, err := projectStream(ctx, tx, projection, streamID)
				if err != nil {
					return err
				}

				streams++
				events += projected
			}
		}

		return nil
	})

	return streams, events, err
}

func conflictOr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == streamVersionUnique {
		return eventsourcing.ErrVersionConflict
	}

	return err
}

func metadataFrom(ctx context.Context) ([]byte, error) {
	carrier := propagation.MapCarrier{}
	propagation.TraceContext{}.Inject(ctx, carrier)

	return json.Marshal(carrier)
}

func (s *Store) IsEmpty(ctx context.Context) (bool, error) {
	var empty bool

	err := s.pool.QueryRow(ctx, "SELECT NOT EXISTS (SELECT 1 FROM events)").Scan(&empty)

	return empty, err
}
