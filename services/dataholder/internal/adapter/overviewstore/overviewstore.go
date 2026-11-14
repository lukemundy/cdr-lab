// Package overviewstore reads the overview of the operator console from the read models in PostgreSQL.
package overviewstore

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

var _ app.OverviewReader = Store{}

func New(pool *pgxpool.Pool) Store {
	return Store{pool: pool}
}

const counts = `SELECT
	(SELECT count(*) FROM customers),
	(SELECT count(*) FROM accounts WHERE open_status = 'OPEN'),
	(SELECT count(*) FROM accounts WHERE open_status = 'CLOSED'),
	(SELECT count(*) FROM arrangements WHERE ends_at > $1),
	(SELECT count(*) FROM arrangements WHERE ends_at <= $1),
	(SELECT count(*) FROM registrations),
	(SELECT count(*) FROM registrations r
		JOIN software_product_statuses s ON s.software_product_id = r.software_id
		WHERE s.status <> 'ACTIVE')`

// Each part gives its own newest rows first, so that the sort of the union reads few rows. The key makes the order of
// changes at the same time stable.
const changes = `SELECT c.at, c.kind, c.customer_number, c.client_id, c.client_name,
	EXISTS (SELECT FROM registrations r WHERE r.client_id = c.client_id), c.product_name, c.sharing_ends_at,
	c.end_reason, c.withdrawal_method
FROM (
	(SELECT created_at AS at, 'CUSTOMER_CREATED' AS kind, customer_number, '' AS client_id, '' AS client_name,
		'' AS product_name, NULL::timestamptz AS sharing_ends_at, '' AS end_reason, '' AS withdrawal_method,
		customer_number AS key
	FROM customers ORDER BY created_at DESC LIMIT $2)
	UNION ALL
	(SELECT created_at, 'ACCOUNT_CREATED', customer_number, '', '', product_name, NULL, '', '', account_id
	FROM accounts ORDER BY created_at DESC LIMIT $2)
	UNION ALL
	(SELECT closed_at, 'ACCOUNT_CLOSED', customer_number, '', '', product_name, NULL, '', '', account_id
	FROM accounts WHERE closed_at IS NOT NULL ORDER BY closed_at DESC LIMIT $2)
	UNION ALL
	(SELECT confirmed_at, 'ARRANGEMENT_CONFIRMED', customer_number, client_id, client_name, '', sharing_ends_at, '',
		'', arrangement_id
	FROM arrangements ORDER BY confirmed_at DESC LIMIT $2)
	UNION ALL
	(SELECT ends_at, 'ARRANGEMENT_ENDED', customer_number, client_id, client_name, '', NULL,
		coalesce(end_reason, ''), coalesce(withdrawal_method, ''), arrangement_id
	FROM arrangements WHERE ends_at <= $1 ORDER BY ends_at DESC LIMIT $2)
	UNION ALL
	(SELECT changed_at, kind, '', client_id, client_name, '', NULL, '', '', software_id || '/' || stream_version
	FROM registration_changes ORDER BY changed_at DESC LIMIT $2)
) c
ORDER BY c.at DESC, c.kind, c.key
LIMIT $2`

func (s Store) Overview(ctx context.Context, at time.Time, newest int) (app.Overview, error) {
	var overview app.Overview

	err := pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly},
		func(tx pgx.Tx) error {
			if err := tx.QueryRow(ctx, counts, at).Scan(
				&overview.Customers, &overview.OpenAccounts, &overview.ClosedAccounts, &overview.CurrentArrangements,
				&overview.EndedArrangements, &overview.Registrations, &overview.SoftwareProductsNotActive,
			); err != nil {
				return err
			}

			rows, err := tx.Query(ctx, changes, at, newest)
			if err != nil {
				return err
			}

			overview.Changes, err = pgx.CollectRows(rows, scanChange)

			return err
		})

	return overview, err
}

func scanChange(row pgx.CollectableRow) (app.Change, error) {
	var (
		c             app.Change
		sharingEndsAt *time.Time
	)

	err := row.Scan(
		&c.Time, &c.Kind, &c.CustomerNumber, &c.ClientID, &c.ClientName, &c.ClientRegistered, &c.ProductName,
		&sharingEndsAt, &c.EndReason, &c.WithdrawalMethod,
	)

	c.Time = c.Time.UTC()
	if sharingEndsAt != nil {
		c.SharingEndsAt = sharingEndsAt.UTC()
	}

	return c, err
}
