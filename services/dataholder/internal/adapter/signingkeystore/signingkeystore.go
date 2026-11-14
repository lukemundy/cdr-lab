// Package signingkeystore keeps the signing key of the data holder in PostgreSQL.
package signingkeystore

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) EnsureSigningKey(ctx context.Context) error {
	key, err := authserver.NewSigningKey()
	if err != nil {
		return err
	}

	der, err := key.MarshalPKCS8()
	if err != nil {
		return err
	}
	// The insert is a separate statement after the lock, so that its snapshot includes a key that another instance
	// committed during the wait.
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext('signing_keys'))"); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO signing_keys (key_id, private_key)
			SELECT $1, $2
			WHERE NOT EXISTS (SELECT FROM signing_keys)`,
			key.ID, der)

		return err
	})
}

func (s *Store) SigningKeys(ctx context.Context) ([]authserver.SigningKey, error) {
	rows, err := s.pool.Query(ctx, "SELECT key_id, private_key FROM signing_keys ORDER BY created_at, key_id")
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (authserver.SigningKey, error) {
		var (
			id  string
			der []byte
		)
		if err := row.Scan(&id, &der); err != nil {
			return authserver.SigningKey{}, err
		}

		return authserver.ParseSigningKey(id, der)
	})
}
