// Package revokedtokenstore keeps the IDs of the revoked access tokens in PostgreSQL.
package revokedtokenstore

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

type Store struct {
	pool *pgxpool.Pool
}

var _ authserver.RevokedTokenStore = (*Store)(nil)

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) RevokeToken(ctx context.Context, id string, expiresAt time.Time) error {
	if _, err := s.pool.Exec(ctx, "DELETE FROM revoked_access_tokens WHERE expires_at < now()"); err != nil {
		return err
	}

	_, err := s.pool.Exec(
		ctx,
		"INSERT INTO revoked_access_tokens (jti, expires_at) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		id, expiresAt,
	)

	return err
}

func (s *Store) TokenRevoked(ctx context.Context, id string) (bool, error) {
	var revoked bool

	err := s.pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM revoked_access_tokens WHERE jti = $1)", id).Scan(&revoked)

	return revoked, err
}
