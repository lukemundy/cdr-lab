// Package jtistore keeps the IDs of the JWTs and SSAs that the data holder accepted, so that it accepts each one time.
package jtistore

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

// retention is longer than a request object of 60 minutes, the longest JWT that the data holder accepts, plus the
// leeway.
const retention = 2 * time.Hour

type Store struct {
	pool *pgxpool.Pool
}

var _ app.JTIStore = (*Store)(nil)

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Consume(ctx context.Context, jti string) error {
	if _, err := s.pool.Exec(
		ctx,
		"DELETE FROM consumed_jtis WHERE consumed_at < now() - $1::interval",
		retention,
	); err != nil {
		return err
	}

	tag, err := s.pool.Exec(ctx, "INSERT INTO consumed_jtis (jti) VALUES ($1) ON CONFLICT DO NOTHING", jti)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return app.ErrJTIUsed
	}

	return nil
}

// ConsumeSSA keeps the JWT ID also at expiresAt, which is the last time at which the data holder accepts the SSA.
func (s *Store) ConsumeSSA(ctx context.Context, jti string, expiresAt time.Time) error {
	if _, err := s.pool.Exec(ctx, "DELETE FROM consumed_ssa_jtis WHERE expires_at < now()"); err != nil {
		return err
	}

	tag, err := s.pool.Exec(
		ctx,
		"INSERT INTO consumed_ssa_jtis (jti, expires_at) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		jti,
		expiresAt,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return app.ErrJTIUsed
	}

	return nil
}

// AuthServer gives the authorisation server the same store, so that a JWT ID is used one time across all endpoints.
type AuthServer struct {
	Store *Store
}

var _ authserver.JTIStore = AuthServer{}

func (a AuthServer) Consume(ctx context.Context, jti string) error {
	err := a.Store.Consume(ctx, jti)
	if errors.Is(err, app.ErrJTIUsed) {
		return authserver.ErrJTIUsed
	}

	return err
}
