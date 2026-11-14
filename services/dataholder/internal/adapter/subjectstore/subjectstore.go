// Package subjectstore keeps the pairwise subject identifiers of the customers in PostgreSQL.
package subjectstore

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

type Store struct {
	pool *pgxpool.Pool
}

var _ app.SubjectStore = (*Store)(nil)

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Subject uses two statements, because one statement does not see the row that a different connection adds at the
// same time.
func (s *Store) Subject(ctx context.Context, customerNumber, sectorIdentifier, candidate string) (string, error) {
	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO subject_identifiers (customer_number, sector_identifier, subject) VALUES ($1, $2, $3)
		ON CONFLICT (customer_number, sector_identifier) DO NOTHING`,
		customerNumber, sectorIdentifier, candidate,
	)
	if err != nil {
		return "", err
	}

	var subject string

	err = s.pool.QueryRow(
		ctx,
		"SELECT subject FROM subject_identifiers WHERE customer_number = $1 AND sector_identifier = $2",
		customerNumber, sectorIdentifier,
	).Scan(&subject)

	return subject, err
}
