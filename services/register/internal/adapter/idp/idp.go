// Package idp connects the Register authorisation server to the participant read models and to PostgreSQL.
package idp

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// jtiRetention is longer than the maximum lifetime of a client assertion that the authorisation server accepts.
const jtiRetention = time.Hour

type Clients struct {
	Reader app.SoftwareProductClientReader
}

var _ authserver.ClientStore = Clients{}

func (c Clients) Client(ctx context.Context, id string) (authserver.Client, error) {
	sp, err := c.Reader.SoftwareProductClient(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return authserver.Client{}, authserver.ErrClientNotFound
	}

	if err != nil {
		return authserver.Client{}, err
	}

	return authserver.Client{
		ID:                     sp.SoftwareProductID,
		JWKSURI:                sp.JWKSURI,
		Active:                 sp.Status == domain.Active,
		CertificateThumbprints: sp.ClientCertificates,
	}, nil
}

type Store struct {
	pool *pgxpool.Pool
}

var _ authserver.JTIStore = (*Store)(nil)

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Consume(ctx context.Context, jti string) error {
	if _, err := s.pool.Exec(
		ctx,
		"DELETE FROM consumed_jtis WHERE consumed_at < now() - $1::interval",
		jtiRetention,
	); err != nil {
		return err
	}

	tag, err := s.pool.Exec(ctx, "INSERT INTO consumed_jtis (jti) VALUES ($1) ON CONFLICT DO NOTHING", jti)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return authserver.ErrJTIUsed
	}

	return nil
}

var (
	_ app.KeyChanger = (*Store)(nil)
	_ app.KeyReader  = (*Store)(nil)
)

// jwksOrder sorts the keys of a set as the JWKS gives them: the key that signs first.
const jwksOrder = "ORDER BY CASE status WHEN 'signing' THEN 0 WHEN 'next' THEN 1 ELSE 2 END, key_id"

func newKey() (id string, der []byte, err error) {
	key, err := authserver.NewSigningKey()
	if err != nil {
		return "", nil, err
	}

	der, err = key.MarshalPKCS8()

	return key.ID, der, err
}

func (s *Store) EnsureSigningKey(ctx context.Context, set app.KeySet) error {
	id, der, err := newKey()
	if err != nil {
		return err
	}

	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			"INSERT INTO key_sets (key_set, version) VALUES ($1, 1) ON CONFLICT DO NOTHING", set); err != nil {
			return err
		}

		// The lock makes sure that two instances that start together create one key.
		_, err := tx.Exec(ctx, `
			WITH lock AS (SELECT pg_advisory_xact_lock(hashtext('signing_keys')))
			INSERT INTO signing_keys (key_id, key_set, private_key, status)
			SELECT $1, $2, $3, $4 FROM lock
			WHERE NOT EXISTS (SELECT FROM signing_keys WHERE key_set = $2)`,
			id, set, der, app.KeySigning)

		return err
	})
}

// PublishNextKey also removes the previous key of the last key change.
func (s *Store) PublishNextKey(ctx context.Context, set app.KeySet) error {
	return s.changeKey(ctx, set, app.PublishNext, nil)
}

func (s *Store) ActivateNextKey(ctx context.Context, set app.KeySet) error {
	return s.changeKey(ctx, set, app.ActivateNext, nil)
}

func (s *Store) ChangeKeyAt(ctx context.Context, set app.KeySet, action app.KeyAction, version int64) error {
	return s.changeKey(ctx, set, action, &version)
}

func (s *Store) changeKey(ctx context.Context, set app.KeySet, action app.KeyAction, version *int64) error {
	var (
		next string
		der  []byte
	)

	if action == app.PublishNext {
		var err error
		if next, der, err = newKey(); err != nil {
			return err
		}
	}

	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		// The update locks the row of the key set, so the steps of one key set occur one at a time.
		tag, err := tx.Exec(ctx, `
			UPDATE key_sets SET version = version + 1
			WHERE key_set = $1 AND ($2::bigint IS NULL OR version = $2)`, set, version)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return app.ErrKeySetChanged
		}

		keys, err := keysOf(ctx, tx, set)
		if err != nil {
			return err
		}

		changes, err := app.PlanKeyChange(keys, action)
		if err != nil {
			return err
		}

		for _, c := range changes {
			if err := applyKeyChange(ctx, tx, c); err != nil {
				return err
			}
		}

		if action != app.PublishNext {
			return nil
		}

		_, err = tx.Exec(ctx,
			"INSERT INTO signing_keys (key_id, key_set, private_key, status) VALUES ($1, $2, $3, $4)",
			next, set, der, app.KeyNext)

		return err
	})
}

func applyKeyChange(ctx context.Context, tx pgx.Tx, change app.KeyChange) error {
	if change.To == "" {
		_, err := tx.Exec(ctx, "DELETE FROM signing_keys WHERE key_id = $1", change.KeyID)
		return err
	}

	_, err := tx.Exec(ctx, "UPDATE signing_keys SET status = $2 WHERE key_id = $1", change.KeyID, change.To)

	return err
}

// SigningKeys gives a function that reads the keys of the set, with the key that signs first.
func (s *Store) SigningKeys(set app.KeySet) func(context.Context) ([]authserver.SigningKey, error) {
	return func(ctx context.Context) ([]authserver.SigningKey, error) { return s.signingKeys(ctx, set) }
}

func (s *Store) signingKeys(ctx context.Context, set app.KeySet) ([]authserver.SigningKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT key_id, private_key FROM signing_keys WHERE key_set = $1 `+jwksOrder, set)
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

// KeySet reads no key material. One query gives the version and the keys, so that they agree.
func (s *Store) KeySet(ctx context.Context, set app.KeySet) (app.KeySetState, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.version, k.key_id, k.status, k.created_at
		FROM key_sets v LEFT JOIN signing_keys k ON k.key_set = v.key_set
		WHERE v.key_set = $1 `+jwksOrder, set)
	if err != nil {
		return app.KeySetState{}, err
	}

	type row struct {
		version int64
		key     *app.Key
	}

	found, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (row, error) {
		var (
			result       row
			id, keyState *string
			created      *time.Time
		)

		if err := r.Scan(&result.version, &id, &keyState, &created); err != nil {
			return row{}, err
		}

		if id != nil {
			result.key = &app.Key{ID: *id, State: app.KeyState(*keyState), CreatedAt: created.UTC()}
		}

		return result, nil
	})
	if err != nil {
		return app.KeySetState{}, err
	}

	state := app.KeySetState{Keys: []app.Key{}}

	for _, r := range found {
		state.Version = r.version
		if r.key != nil {
			state.Keys = append(state.Keys, *r.key)
		}
	}

	return state, nil
}

func keysOf(ctx context.Context, tx pgx.Tx, set app.KeySet) ([]app.Key, error) {
	rows, err := tx.Query(ctx, "SELECT key_id, status, created_at FROM signing_keys WHERE key_set = $1 "+jwksOrder, set)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (app.Key, error) {
		var key app.Key

		err := row.Scan(&key.ID, &key.State, &key.CreatedAt)

		return key, err
	})
}
