// Package arrangementstore keeps the read model of the CDR arrangements in PostgreSQL and queries it.
package arrangementstore

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func Projections() []eventstore.Projection {
	return []eventstore.Projection{{StreamType: app.ArrangementType.StreamType, Project: project}}
}

// stateJSON keeps the times of the events with their full precision, which a timestamptz column does not.
type stateJSON struct {
	Arrangement      domain.Arrangement      `json:"arrangement"`
	EndedAt          *time.Time              `json:"endedAt,omitempty"`
	EndReason        domain.EndReason        `json:"endReason,omitempty"`
	WithdrawalMethod domain.WithdrawalMethod `json:"withdrawalMethod,omitempty"`
	RecordedNotice   domain.NoticeStatus     `json:"recordedNotice,omitempty"`
	NoticeAttempts   []domain.NoticeAttempt  `json:"noticeAttempts,omitempty"`
}

func project(ctx context.Context, tx pgx.Tx, arrangementID string, recorded []eventsourcing.RecordedEvent) error {
	state, _, err := app.ArrangementType.Fold(arrangementID, recorded)
	if err != nil || !state.Created {
		return err
	}

	stored := stateJSON{
		Arrangement: state.Arrangement, EndReason: state.EndReason, WithdrawalMethod: state.WithdrawalMethod,
		RecordedNotice: state.RecordedNotice, NoticeAttempts: state.NoticeAttempts,
	}
	if !state.EndedAt.IsZero() {
		stored.EndedAt = &state.EndedAt
	}

	a := state.Arrangement
	if _, err := tx.Exec(ctx, `
		INSERT INTO arrangements (
			arrangement_id, customer_number, client_id, software_id, client_name, confirmed_at, sharing_ends_at,
			ends_at, end_reason, withdrawal_method, state
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, nullif($9, ''), nullif($10, ''), $11)
		ON CONFLICT (arrangement_id) DO UPDATE SET
			client_name = excluded.client_name, sharing_ends_at = excluded.sharing_ends_at, ends_at = excluded.ends_at,
			end_reason = excluded.end_reason, withdrawal_method = excluded.withdrawal_method, state = excluded.state`,
		arrangementID, a.CustomerNumber, a.ClientID, a.SoftwareID, a.ClientName, a.ConfirmedAt, a.SharingEndsAt(),
		state.EndsAt(), state.EndReason, state.WithdrawalMethod, stored,
	); err != nil {
		return err
	}

	// The registration projection also writes this count, so the sequence of the projections has no effect.
	_, err = tx.Exec(ctx, `
		UPDATE registrations SET arrangement_count = (SELECT count(*) FROM arrangements WHERE client_id = $1)
		WHERE client_id = $1`, a.ClientID)

	return err
}

type Queries struct {
	pool *pgxpool.Pool
}

var _ app.ArrangementReader = Queries{}

func NewQueries(pool *pgxpool.Pool) Queries {
	return Queries{pool: pool}
}

func (q Queries) ArrangementsOf(ctx context.Context, customerNumber string) ([]domain.ArrangementState, error) {
	return q.states(ctx, `
		SELECT state FROM arrangements
		WHERE customer_number = $1
		ORDER BY confirmed_at, arrangement_id`, customerNumber)
}

func (q Queries) ArrangementsOfSoftwareProduct(
	ctx context.Context,
	softwareID string,
) ([]domain.ArrangementState, error) {
	return q.states(ctx, `
		SELECT state FROM arrangements
		WHERE software_id = $1
		ORDER BY confirmed_at, arrangement_id`, softwareID)
}

// ArrangementPageOfClient is newest first by the time of the confirmation, and then by the arrangement ID.
func (q Queries) ArrangementPageOfClient(
	ctx context.Context,
	clientID string,
	page, pageSize int,
) ([]app.ClientArrangement, int, error) {
	var total int
	if err := q.pool.QueryRow(ctx, "SELECT count(*) FROM arrangements WHERE client_id = $1", clientID).
		Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := q.pool.Query(ctx, `
		SELECT a.state, c.first_name, c.last_name
		FROM arrangements a JOIN customers c ON c.customer_number = a.customer_number
		WHERE a.client_id = $1
		ORDER BY a.confirmed_at DESC, a.arrangement_id
		LIMIT $2 OFFSET $3`, clientID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}

	found, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (app.ClientArrangement, error) {
		var (
			stored      stateJSON
			arrangement app.ClientArrangement
		)

		err := row.Scan(&stored, &arrangement.FirstName, &arrangement.LastName)
		arrangement.State = stored.state()

		return arrangement, err
	})

	return found, total, err
}

func (q Queries) states(ctx context.Context, query string, arg string) ([]domain.ArrangementState, error) {
	rows, err := q.pool.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ArrangementState, error) {
		var stored stateJSON
		if err := row.Scan(&stored); err != nil {
			return domain.ArrangementState{}, err
		}

		return stored.state(), nil
	})
}

func (stored stateJSON) state() domain.ArrangementState {
	state := domain.ArrangementState{
		Created:          true,
		Arrangement:      stored.Arrangement,
		EndReason:        stored.EndReason,
		WithdrawalMethod: stored.WithdrawalMethod,
		RecordedNotice:   stored.RecordedNotice,
		NoticeAttempts:   stored.NoticeAttempts,
	}
	if stored.EndedAt != nil {
		state.EndedAt = *stored.EndedAt
	}

	return state
}
