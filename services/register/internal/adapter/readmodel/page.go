package readmodel

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

// The "C" collation gives the same order in each database locale.
type sortColumns map[app.ParticipantSort]string

type pageQuery struct {
	// selectList and from are constant SQL. The arguments of from are $1 to $n.
	selectList, from string
	args             []any
	sort             app.ParticipantSort
	descending       bool
	page, pageSize   int
}

// queryPage sorts by the legal entity ID after the sort column, so that each page is stable.
func queryPage[T any](ctx context.Context, q *Queries, columns sortColumns, query pageQuery) ([]T, int, error) {
	column, ok := columns[query.sort]
	if !ok {
		return nil, 0, fmt.Errorf("unknown sort %q for this list", query.sort)
	}

	direction := "ASC"
	if query.descending {
		direction = "DESC"
	}

	var total int
	if err := q.pool.QueryRow(ctx, "SELECT count(*) "+query.from, query.args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitArg := len(query.args) + 1

	rows, err := q.pool.Query(
		ctx,
		fmt.Sprintf("SELECT %s %s ORDER BY %s %s, legal_entity_id LIMIT $%d OFFSET $%d",
			query.selectList, query.from, column, direction, limitArg, limitArg+1),
		append(query.args, query.pageSize, (query.page-1)*query.pageSize)...,
	)
	if err != nil {
		return nil, 0, err
	}

	page, err := pgx.CollectRows(rows, pgx.RowToStructByPos[T])

	return page, total, err
}
