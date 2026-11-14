// Package disclosurestore keeps the record of disclosures of the data holder in PostgreSQL.
package disclosurestore

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

type Store struct {
	pool *pgxpool.Pool
}

var _ app.DisclosureStore = (*Store)(nil)

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) AddDisclosure(ctx context.Context, arrangementID, scope string, at time.Time) error {
	_, err := s.pool.Exec(
		ctx,
		"INSERT INTO disclosures (arrangement_id, scope, disclosed_at) VALUES ($1, $2, $3)",
		arrangementID, scope, at,
	)

	return err
}

func (s *Store) FirstAndLastDisclosures(
	ctx context.Context,
	arrangementIDs []string,
) (map[app.DisclosureKey]app.DisclosurePeriod, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT arrangement_id, scope, min(disclosed_at), max(disclosed_at) FROM disclosures
		WHERE arrangement_id = ANY($1)
		GROUP BY arrangement_id, scope`, arrangementIDs)
	if err != nil {
		return nil, err
	}

	type row struct {
		key    app.DisclosureKey
		period app.DisclosurePeriod
	}

	found, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (row, error) {
		var got row
		if err := r.Scan(&got.key.ArrangementID, &got.key.Scope, &got.period.First, &got.period.Last); err != nil {
			return row{}, err
		}

		return got, nil
	})
	if err != nil {
		return nil, err
	}

	periods := make(map[app.DisclosureKey]app.DisclosurePeriod, len(found))
	for _, f := range found {
		periods[f.key] = app.DisclosurePeriod{First: f.period.First.UTC(), Last: f.period.Last.UTC()}
	}

	return periods, nil
}
