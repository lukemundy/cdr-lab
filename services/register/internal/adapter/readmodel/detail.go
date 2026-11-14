package readmodel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var (
	_ app.DataRecipientDetailReader = (*Queries)(nil)
	_ app.DataHolderDetailReader    = (*Queries)(nil)
)

func (q *Queries) DataRecipient(ctx context.Context, legalEntityID string) (app.DataRecipientItem, error) {
	return queryOne[app.DataRecipientItem](ctx, q, "data recipient",
		"SELECT data_recipient, last_updated, version FROM data_recipients WHERE legal_entity_id = $1", legalEntityID)
}

func (q *Queries) DataHolder(ctx context.Context, legalEntityID string) (app.DataHolderItem, error) {
	return queryOne[app.DataHolderItem](ctx, q, "data holder",
		"SELECT data_holder, last_updated, version FROM data_holders WHERE legal_entity_id = $1", legalEntityID)
}

func queryOne[T any](ctx context.Context, q *Queries, kind, sql string, args ...any) (T, error) {
	rows, err := q.pool.Query(ctx, sql, args...)
	if err != nil {
		var zero T
		return zero, err
	}

	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[T])
	if errors.Is(err, pgx.ErrNoRows) {
		return item, fmt.Errorf("%s %w", kind, domain.ErrNotFound)
	}

	return item, err
}
