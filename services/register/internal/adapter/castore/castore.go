// Package castore keeps the revocations and the newest CRL of the mock CDR intermediate in PostgreSQL.
package castore

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

type Store struct {
	pool *pgxpool.Pool
}

var _ app.CAStore = (*Store)(nil)

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

const selectCRL = "SELECT crl_number, this_update, next_update, crl FROM cdr_crl"

func (s *Store) NewestCRL(ctx context.Context) (app.CRL, error) {
	return newestCRL(ctx, s.pool)
}

func (s *Store) ChangeCA(ctx context.Context, change func(app.CAState) (app.CAChange, error)) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		// The lock also covers a CA that has no CRL row yet, so that two instances that start together do not sign
		// two CRLs with the same number.
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext('cdr_crl'))"); err != nil {
			return err
		}

		crl, err := newestCRL(ctx, tx)
		if err != nil {
			return err
		}

		revoked, err := revocations(ctx, tx)
		if err != nil {
			return err
		}

		c, err := change(app.CAState{CRL: crl, Revoked: revoked})
		if err != nil {
			return err
		}

		if r := c.Revocation; r != nil {
			if _, err := tx.Exec(ctx, `
				INSERT INTO cdr_revocations (serial, subject, thumbprint, revoked_at, reason)
				VALUES ($1, $2, $3, $4, $5)`,
				r.Serial.Text(16), r.Subject, r.Thumbprint, r.Time, int(r.Reason)); err != nil {
				return err
			}
		}

		if c.CRL == nil {
			return nil
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO cdr_crl (crl_number, this_update, next_update, crl) VALUES ($1, $2, $3, $4)
			ON CONFLICT (one_row) DO UPDATE SET
				crl_number = excluded.crl_number, this_update = excluded.this_update,
				next_update = excluded.next_update, crl = excluded.crl`,
			c.CRL.Number, c.CRL.ThisUpdate, c.CRL.NextUpdate, c.CRL.DER)

		return err
	})
}

type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func newestCRL(ctx context.Context, q querier) (app.CRL, error) {
	var crl app.CRL

	err := q.QueryRow(ctx, selectCRL).Scan(&crl.Number, &crl.ThisUpdate, &crl.NextUpdate, &crl.DER)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.CRL{}, nil
	}

	if err != nil {
		return app.CRL{}, err
	}

	crl.ThisUpdate, crl.NextUpdate = crl.ThisUpdate.UTC(), crl.NextUpdate.UTC()

	return crl, nil
}

const selectRevocations = "SELECT serial, subject, thumbprint, revoked_at, reason FROM cdr_revocations"

func (s *Store) Revocation(ctx context.Context, serial *big.Int) (app.Revocation, bool, error) {
	rows, err := s.pool.Query(ctx, selectRevocations+" WHERE serial = $1", serial.Text(16))
	if err != nil {
		return app.Revocation{}, false, err
	}

	r, err := pgx.CollectExactlyOneRow(rows, scanRevocation)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Revocation{}, false, nil
	}

	return r, err == nil, err
}

// Revocations sorts by the length of the serial number first, which makes the sort numeric with no leading zeros.
func (s *Store) Revocations(ctx context.Context, query app.RevocationPageQuery) ([]app.Revocation, int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM cdr_revocations").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, selectRevocations+`
		ORDER BY revoked_at DESC, length(serial), serial LIMIT $1 OFFSET $2`,
		query.PageSize, (query.Page-1)*query.PageSize)
	if err != nil {
		return nil, 0, err
	}

	page, err := pgx.CollectRows(rows, scanRevocation)

	return page, total, err
}

func revocations(ctx context.Context, q querier) ([]app.Revocation, error) {
	rows, err := q.Query(ctx, selectRevocations+" ORDER BY revoked_at, serial")
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scanRevocation)
}

func scanRevocation(row pgx.CollectableRow) (app.Revocation, error) {
	var (
		serial string
		r      app.Revocation
		at     time.Time
		reason int
	)

	if err := row.Scan(&serial, &r.Subject, &r.Thumbprint, &at, &reason); err != nil {
		return app.Revocation{}, err
	}

	var ok bool
	if r.Serial, ok = new(big.Int).SetString(serial, 16); !ok {
		return app.Revocation{}, fmt.Errorf("the serial number %q of a revocation is not hexadecimal", serial)
	}

	r.Time, r.Reason = at.UTC(), ca.Reason(reason)

	return r, nil
}
