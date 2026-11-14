// Package transactionstore keeps the transactions of the accounts in PostgreSQL.
package transactionstore

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type Store struct {
	pool *pgxpool.Pool
}

var _ app.TransactionStore = (*Store)(nil)

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

const columns = `transaction_id, type, status, description, amount_cents, reference, posting_date_time,
	execution_date_time, value_date_time, merchant_name, merchant_category_code, biller_code, biller_name, crn,
	apca_number, extended_data`

func (s *Store) AddTransaction(ctx context.Context, accountID string, t domain.Transaction) error {
	return insert(ctx, s.pool, accountID, t)
}

// Insert adds the transaction in tx, so that a different store can add it in the same write as its own rows.
func Insert(ctx context.Context, tx pgx.Tx, accountID string, t domain.Transaction) error {
	return insert(ctx, tx, accountID, t)
}

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func insert(ctx context.Context, db execer, accountID string, t domain.Transaction) error {
	_, err := db.Exec(
		ctx,
		"INSERT INTO transactions (account_id, "+columns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		accountID, t.TransactionID, t.Type, t.Status, t.Description, int64(t.Amount), t.Reference,
		t.PostingDateTime, t.ExecutionDateTime, t.ValueDateTime, t.MerchantName, t.MerchantCategoryCode,
		t.BillerCode, t.BillerName, t.CRN, t.APCANumber, t.ExtendedData,
	)

	return err
}

func (s *Store) Transactions(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	rows, err := s.pool.Query(
		ctx,
		"SELECT "+columns+" FROM transactions WHERE account_id = $1 ORDER BY transaction_sequence",
		accountID,
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, scan)
}

// The text is a parameter of strpos, so that the characters of a LIKE pattern have no special meaning.
const matchingTransactions = `FROM transactions
	WHERE account_id = $1
	AND effective_time BETWEEN $2 AND $3
	AND ($4::bigint IS NULL OR amount_cents >= $4)
	AND ($5::bigint IS NULL OR amount_cents <= $5)
	AND ($6 = '' OR strpos(lower(description), $6) > 0 OR strpos(lower(reference), $6) > 0)`

func (s *Store) FindTransactions(ctx context.Context, q app.TransactionQuery) ([]domain.Transaction, int, error) {
	filters := []any{q.AccountID, q.Oldest, q.Newest, q.MinAmount, q.MaxAmount, strings.ToLower(q.Text)}

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) "+matchingTransactions, filters...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(
		ctx,
		"SELECT "+columns+" "+matchingTransactions+
			" ORDER BY effective_time DESC, transaction_sequence DESC LIMIT $7 OFFSET $8",
		append(filters, q.PageSize, (q.Page-1)*q.PageSize)...,
	)
	if err != nil {
		return nil, 0, err
	}

	page, err := pgx.CollectRows(rows, scan)

	return page, total, err
}

func (s *Store) Totals(ctx context.Context, accountIDs []string) (map[string]domain.TransactionTotals, error) {
	rows, err := s.pool.Query(ctx, `SELECT account_id,
		coalesce(sum(amount_cents) FILTER (WHERE status = 'POSTED'), 0)::bigint,
		coalesce(sum(amount_cents) FILTER (WHERE status = 'PENDING' AND amount_cents < 0), 0)::bigint
		FROM transactions WHERE account_id = ANY($1) GROUP BY account_id`, accountIDs)
	if err != nil {
		return nil, err
	}

	found, err := pgx.CollectRows(rows, pgx.RowToStructByPos[accountTotals])
	if err != nil {
		return nil, err
	}

	totals := make(map[string]domain.TransactionTotals, len(found))
	for _, row := range found {
		totals[row.AccountID] = domain.TransactionTotals{Posted: row.Posted, PendingDebits: row.PendingDebits}
	}

	return totals, nil
}

type accountTotals struct {
	AccountID     string
	Posted        domain.Amount
	PendingDebits domain.Amount
}

func scan(row pgx.CollectableRow) (domain.Transaction, error) {
	var t domain.Transaction

	err := row.Scan(
		&t.TransactionID, &t.Type, &t.Status, &t.Description, &t.Amount, &t.Reference, &t.PostingDateTime,
		&t.ExecutionDateTime, &t.ValueDateTime, &t.MerchantName, &t.MerchantCategoryCode, &t.BillerCode,
		&t.BillerName, &t.CRN, &t.APCANumber, &t.ExtendedData,
	)
	if err != nil {
		return domain.Transaction{}, err
	}

	return t.InUTC(), nil
}
