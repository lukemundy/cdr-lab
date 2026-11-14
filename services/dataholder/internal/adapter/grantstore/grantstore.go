// Package grantstore keeps the grants of the authorisation server in PostgreSQL.
package grantstore

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

type Store struct {
	pool *pgxpool.Pool
}

var (
	_ authserver.GrantStore = (*Store)(nil)
	_ app.GrantRevoker      = (*Store)(nil)
)

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

const columns = `grant_id, client_id, subject, scopes, arrangement_id, sharing_duration_seconds, authenticated_at,
	sharing_ends_at, created_at, revoked_at, code, code_expires_at, code_consumed_at, redirect_uri, nonce,
	code_challenge, refresh_token, certificate_thumbprint`

func (s *Store) SaveGrant(ctx context.Context, grant authserver.Grant) error {
	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO grants (`+columns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (grant_id) DO UPDATE SET
			revoked_at = EXCLUDED.revoked_at,
			code_consumed_at = EXCLUDED.code_consumed_at,
			certificate_thumbprint = EXCLUDED.certificate_thumbprint`,
		grant.ID, grant.ClientID, grant.Subject, grant.Scopes, grant.ArrangementID,
		int64(grant.SharingDuration/time.Second), grant.AuthenticatedAt, grant.SharingEndsAt, grant.CreatedAt,
		optionalTime(grant.RevokedAt), grant.Code, grant.CodeExpiresAt, optionalTime(grant.CodeConsumedAt),
		grant.RedirectURI, grant.Nonce, grant.CodeChallenge, optionalText(grant.RefreshToken),
		grant.CertificateThumbprint,
	)

	return err
}

func (s *Store) RevokeArrangement(ctx context.Context, arrangementID string) error {
	_, err := s.pool.Exec(
		ctx,
		"UPDATE grants SET revoked_at = now() WHERE arrangement_id = $1 AND revoked_at IS NULL",
		arrangementID,
	)

	return err
}

func (s *Store) Grant(ctx context.Context, id string) (authserver.Grant, error) {
	return s.grantWhere(ctx, "grant_id = $1", id)
}

func (s *Store) GrantByCode(ctx context.Context, code string) (authserver.Grant, error) {
	return s.grantWhere(ctx, "code = $1", code)
}

func (s *Store) GrantByRefreshToken(ctx context.Context, refreshToken string) (authserver.Grant, error) {
	return s.grantWhere(ctx, "refresh_token = $1", refreshToken)
}

// grantWhere needs a condition that is a constant of this package.
func (s *Store) grantWhere(ctx context.Context, condition, value string) (authserver.Grant, error) {
	var (
		grant                     authserver.Grant
		sharingSeconds            int64
		revokedAt, codeConsumedAt *time.Time
		refreshToken              *string
	)

	err := s.pool.QueryRow(ctx, "SELECT "+columns+" FROM grants WHERE "+condition, value).Scan(
		&grant.ID, &grant.ClientID, &grant.Subject, &grant.Scopes, &grant.ArrangementID, &sharingSeconds,
		&grant.AuthenticatedAt, &grant.SharingEndsAt, &grant.CreatedAt, &revokedAt, &grant.Code,
		&grant.CodeExpiresAt, &codeConsumedAt, &grant.RedirectURI, &grant.Nonce, &grant.CodeChallenge,
		&refreshToken, &grant.CertificateThumbprint,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return authserver.Grant{}, authserver.ErrGrantNotFound
	}

	if err != nil {
		return authserver.Grant{}, err
	}

	grant.SharingDuration = time.Duration(sharingSeconds) * time.Second
	grant.AuthenticatedAt = grant.AuthenticatedAt.UTC()
	grant.SharingEndsAt = grant.SharingEndsAt.UTC()
	grant.CreatedAt = grant.CreatedAt.UTC()
	grant.CodeExpiresAt = grant.CodeExpiresAt.UTC()

	if revokedAt != nil {
		grant.RevokedAt = revokedAt.UTC()
	}

	if codeConsumedAt != nil {
		grant.CodeConsumedAt = codeConsumedAt.UTC()
	}

	if refreshToken != nil {
		grant.RefreshToken = *refreshToken
	}

	return grant, nil
}

func optionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}

	return &t
}

func optionalText(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
