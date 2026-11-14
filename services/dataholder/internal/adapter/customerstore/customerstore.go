// Package customerstore keeps the customers and the accounts of the data holder in PostgreSQL.
package customerstore

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
	// The other unique constraint of the accounts table is for the account identifier.
	accountNumberConstraint = "accounts_bsb_account_number_key"
	emailAddressConstraint  = "customers_email_address_key"
)

// NoOpenAccountHook runs in the transaction of the close of the last open account of a customer.
type NoOpenAccountHook func(ctx context.Context, tx pgx.Tx, customerNumber string, closedAt time.Time) error

type TransactionInserter func(ctx context.Context, tx pgx.Tx, accountID string, t domain.Transaction) error

type Store struct {
	pool              *pgxpool.Pool
	onNoOpenAccount   NoOpenAccountHook
	insertTransaction TransactionInserter
}

var _ app.CustomerStore = (*Store)(nil)

// New takes a nil onNoOpenAccount when nothing must occur at the close of the last open account, and a nil
// insertTransaction when no account is added with a transaction.
func New(pool *pgxpool.Pool, onNoOpenAccount NoOpenAccountHook, insertTransaction TransactionInserter) *Store {
	return &Store{pool: pool, onNoOpenAccount: onNoOpenAccount, insertTransaction: insertTransaction}
}

func violates(err error, code string) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == code
}

// violatesConstraint is for a table with more than one unique constraint.
func violatesConstraint(err error, constraint string) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == constraint
}

func (s *Store) AddCustomer(ctx context.Context, customer domain.Customer) error {
	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO customers (customer_number, first_name, last_name, email_address, last_updated, created_at)
		VALUES ($1, $2, $3, $4, $5, $5)`,
		customer.CustomerNumber, customer.FirstName, customer.LastName, customer.EmailAddress, customer.LastUpdated,
	)
	switch {
	case violatesConstraint(err, emailAddressConstraint):
		return app.ErrEmailAddressUsed
	case violates(err, uniqueViolation):
		return app.ErrCustomerNumberUsed
	default:
		return err
	}
}

func (s *Store) CustomerNumberOf(ctx context.Context, emailAddress string) (string, error) {
	var customerNumber string

	err := s.pool.QueryRow(
		ctx, "SELECT customer_number FROM customers WHERE lower(email_address) = lower($1)", emailAddress,
	).Scan(&customerNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", app.ErrCustomerNotFound
	}

	return customerNumber, err
}

func (s *Store) AddAccount(ctx context.Context, customerNumber string, account domain.Account) error {
	return addAccount(ctx, s.pool, customerNumber, account)
}

var errNoTransactionInserter = errors.New("the customer store has no transaction inserter")

func (s *Store) AddAccountWithTransaction(
	ctx context.Context,
	customerNumber string,
	account domain.Account,
	transaction domain.Transaction,
) error {
	if s.insertTransaction == nil {
		return errNoTransactionInserter
	}

	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := addAccount(ctx, tx, customerNumber, account); err != nil {
			return err
		}

		return s.insertTransaction(ctx, tx, account.AccountID, transaction)
	})
}

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func addAccount(ctx context.Context, db execer, customerNumber string, account domain.Account) error {
	_, err := db.Exec(
		ctx,
		`INSERT INTO accounts (
			account_id, customer_number, display_name, product_category, product_name, bsb, account_number,
			credit_limit_cents, open_status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		account.AccountID, customerNumber, account.DisplayName, account.ProductCategory, account.ProductName,
		account.BSB, account.AccountNumber, int64(account.CreditLimit), account.OpenStatus, account.CreatedAt,
	)

	switch {
	case violates(err, foreignKeyViolation):
		return app.ErrCustomerNotFound
	case violatesConstraint(err, accountNumberConstraint):
		return app.ErrAccountExists
	default:
		return err
	}
}

func (s *Store) Customer(ctx context.Context, customerNumber string) (app.CustomerRecord, error) {
	record := app.CustomerRecord{Customer: domain.Customer{CustomerNumber: customerNumber}}

	err := s.pool.QueryRow(
		ctx,
		"SELECT first_name, last_name, email_address, last_updated FROM customers WHERE customer_number = $1",
		customerNumber,
	).Scan(
		&record.Customer.FirstName, &record.Customer.LastName, &record.Customer.EmailAddress,
		&record.Customer.LastUpdated,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.CustomerRecord{}, app.ErrCustomerNotFound
	}

	if err != nil {
		return app.CustomerRecord{}, err
	}

	record.Customer.LastUpdated = record.Customer.LastUpdated.UTC()

	rows, err := s.pool.Query(
		ctx,
		`SELECT account_id, display_name, product_category, product_name, bsb, account_number, credit_limit_cents,
			open_status, created_at
		FROM accounts WHERE customer_number = $1 ORDER BY account_sequence`,
		customerNumber,
	)
	if err != nil {
		return app.CustomerRecord{}, err
	}

	record.Accounts, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Account, error) {
		var a domain.Account

		err := row.Scan(
			&a.AccountID, &a.DisplayName, &a.ProductCategory, &a.ProductName, &a.BSB, &a.AccountNumber,
			&a.CreditLimit, &a.OpenStatus, &a.CreatedAt,
		)
		a.CreatedAt = a.CreatedAt.UTC()

		return a, err
	})
	if err != nil {
		return app.CustomerRecord{}, err
	}

	return record, nil
}

// CloseAccount locks the customer, so that of two closes at the same time the second sees the first, also in its check.
func (s *Store) CloseAccount(
	ctx context.Context,
	account domain.Account,
	closedAt time.Time,
	check app.CloseCheck,
) (bool, error) {
	noOpenAccount := false

	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var customerNumber string

		err := tx.QueryRow(ctx, `SELECT customer_number FROM customers
			WHERE customer_number = (SELECT customer_number FROM accounts WHERE account_id = $1)
			FOR UPDATE`, account.AccountID).Scan(&customerNumber)
		if errors.Is(err, pgx.ErrNoRows) {
			return app.ErrAccountNotFound
		}

		if err != nil {
			return err
		}

		tag, err := tx.Exec(ctx,
			"UPDATE accounts SET open_status = $1, closed_at = $2 WHERE account_id = $3 AND open_status = $4",
			domain.OpenStatusClosed, closedAt, account.AccountID, domain.OpenStatusOpen)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return domain.ErrAccountClosed
		}

		err = tx.QueryRow(
			ctx, "SELECT NOT EXISTS (SELECT FROM accounts WHERE customer_number = $1 AND open_status = $2)",
			customerNumber, domain.OpenStatusOpen,
		).Scan(&noOpenAccount)
		if err != nil {
			return err
		}

		if check != nil {
			if err := check(ctx, noOpenAccount); err != nil {
				return err
			}
		}

		if !noOpenAccount || s.onNoOpenAccount == nil {
			return nil
		}

		return s.onNoOpenAccount(ctx, tx, customerNumber, closedAt)
	})

	return noOpenAccount && err == nil, err
}
