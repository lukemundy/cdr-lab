// Package statusstore keeps the statuses of the software products in PostgreSQL.
package statusstore

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var (
	_ app.StatusStore      = (*Store)(nil)
	_ app.StatusListReader = (*Store)(nil)
)

// RemovedHook runs in the transaction of the save of a status that becomes REMOVED.
type RemovedHook func(ctx context.Context, tx pgx.Tx, softwareProductID string, removedAt time.Time) error

type Store struct {
	pool      *pgxpool.Pool
	onRemoved RemovedHook
}

// New takes a nil onRemoved for a store that only reads the statuses.
func New(pool *pgxpool.Pool, onRemoved RemovedHook) *Store {
	return &Store{pool: pool, onRemoved: onRemoved}
}

func (s *Store) ETag(ctx context.Context) (string, error) {
	var etag string

	err := s.pool.QueryRow(ctx, "SELECT etag FROM software_product_statuses_response").Scan(&etag)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}

	return etag, err
}

func (s *Store) Save(ctx context.Context, statuses map[string]domain.SoftwareProductStatus, etag string) error {
	receivedAt := time.Now().UTC()

	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for id, status := range statuses {
			changed, err := upsert(ctx, tx, id, status)
			if err != nil {
				return err
			}

			if !changed || status != domain.SoftwareProductRemoved || s.onRemoved == nil {
				continue
			}

			if err := s.onRemoved(ctx, tx, id, receivedAt); err != nil {
				return err
			}
		}

		_, err := tx.Exec(ctx, `INSERT INTO software_product_statuses_response (etag, saved_at) VALUES ($1, $2)
			ON CONFLICT (only_row) DO UPDATE SET etag = excluded.etag, saved_at = excluded.saved_at`, etag, receivedAt)

		return err
	})
}

// upsert is true only for a status that is new or that changes.
func upsert(ctx context.Context, tx pgx.Tx, id string, status domain.SoftwareProductStatus) (bool, error) {
	var changed bool

	err := tx.QueryRow(ctx, `INSERT INTO software_product_statuses (software_product_id, status) VALUES ($1, $2)
		ON CONFLICT (software_product_id) DO UPDATE SET status = excluded.status
		WHERE software_product_statuses.status <> excluded.status
		RETURNING true`, id, status).Scan(&changed)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	return changed, err
}

// SaveNotModified changes saved_at, which is then the time of the last response, also with no change.
func (s *Store) SaveNotModified(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "UPDATE software_product_statuses_response SET saved_at = $1", time.Now().UTC())

	return err
}

var statusSorts = map[app.StatusSort]func(a, b app.ProductStatus) int{
	// The byte order of the lower case is the order of lower(name) COLLATE "C" in the other lists.
	app.SortStatusesByName: func(a, b app.ProductStatus) int {
		return strings.Compare(strings.ToLower(a.ClientName), strings.ToLower(b.ClientName))
	},
	app.SortStatusesByStatus: func(a, b app.ProductStatus) int {
		return strings.Compare(string(a.Status), string(b.Status))
	},
}

const statusRows = `SELECT s.software_product_id, s.status, r.registration
	FROM software_product_statuses s LEFT JOIN registrations r ON r.software_id = s.software_product_id
	WHERE $1::boolean IS NULL OR (r.client_id IS NOT NULL) = $1`

// StatusList reads the time before the statuses, so that the statuses are never older than the time.
func (s *Store) StatusList(ctx context.Context, query app.StatusListQuery) (app.StatusList, error) {
	var list app.StatusList

	compare, ok := statusSorts[query.Sort]
	if !ok {
		return list, fmt.Errorf("unknown sort %q for the statuses", query.Sort)
	}

	err := s.pool.QueryRow(ctx, "SELECT saved_at FROM software_product_statuses_response").Scan(&list.ReceivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return list, nil
	}

	if err != nil {
		return list, err
	}

	list.ReceivedAt = list.ReceivedAt.UTC()

	rows, err := s.pool.Query(ctx, statusRows, query.Registered)
	if err != nil {
		return list, err
	}

	list.Products, err = pgx.CollectRows(rows, scanProductStatus)
	// The sort is here, because no query reads the content of the registration column. The lab has few statuses.
	slices.SortFunc(list.Products, func(a, b app.ProductStatus) int {
		order := compare(a, b)
		if query.Descending {
			order = -order
		}

		if query.Sort == app.SortStatusesByName {
			order = cmp.Or(cmp.Compare(noRegistration(a), noRegistration(b)), order)
		}

		return cmp.Or(order, strings.Compare(a.SoftwareProductID, b.SoftwareProductID))
	})

	return list, err
}

func noRegistration(p app.ProductStatus) int {
	if p.ClientID == "" {
		return 1
	}

	return 0
}

func scanProductStatus(row pgx.CollectableRow) (app.ProductStatus, error) {
	var (
		status       app.ProductStatus
		registration *domain.Registration
	)

	if err := row.Scan(&status.SoftwareProductID, &status.Status, &registration); err != nil {
		return status, err
	}

	if registration != nil {
		status.ClientID, status.ClientName = registration.ClientID, registration.ClientName
	}

	return status, nil
}

func (s *Store) HasStatuses(ctx context.Context) (bool, error) {
	var found bool

	err := s.pool.QueryRow(ctx, "SELECT EXISTS (SELECT FROM software_product_statuses_response)").Scan(&found)

	return found, err
}

func (s *Store) Status(ctx context.Context, softwareProductID string) (domain.SoftwareProductStatus, bool, error) {
	var status domain.SoftwareProductStatus

	err := s.pool.QueryRow(
		ctx, "SELECT status FROM software_product_statuses WHERE software_product_id = $1", softwareProductID,
	).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}

	if err != nil {
		return "", false, err
	}

	return status, true, nil
}
