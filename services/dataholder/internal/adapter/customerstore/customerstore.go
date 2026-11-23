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

// db is a pool, or a transaction in which each write of the store has a savepoint.
type db interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Store struct {
	db                db
	onNoOpenAccount   NoOpenAccountHook
	insertTransaction TransactionInserter
}

var _ app.CustomerStore = (*Store)(nil)

// New takes a nil onNoOpenAccount when nothing must occur at the close of the last open account, and a nil
// insertTransaction when no account is added with a transaction.
func New(pool *pgxpool.Pool, onNoOpenAccount NoOpenAccountHook, insertTransaction TransactionInserter) *Store {
	return &Store{db: pool, onNoOpenAccount: onNoOpenAccount, insertTransaction: insertTransaction}
}

// InTransaction gives write a store that writes in tx. When write gives an error, nothing that it wrote stays.
func (s *Store) InTransaction(ctx context.Context, write func(store *Store, tx pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		return write(&Store{db: tx, onNoOpenAccount: s.onNoOpenAccount, insertTransaction: s.insertTransaction}, tx)
	})
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

// AddCustomer has its own transaction, so that in the transaction of a load a customer number that is in use undoes
// only this insert, and Create can try a different number.
func (s *Store) AddCustomer(ctx context.Context, customer domain.Customer) error {
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO customers (customer_number, first_name, last_name, email_address, last_updated, created_at)
			VALUES ($1, $2, $3, $4, $5, $5)`,
			customer.CustomerNumber, customer.FirstName, customer.LastName, customer.EmailAddress,
			customer.LastUpdated,
		)

		return err
	})
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

	err := s.db.QueryRow(
		ctx, "SELECT customer_number FROM customers WHERE lower(email_address) = lower($1)", emailAddress,
	).Scan(&customerNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", app.ErrCustomerNotFound
	}

	return customerNumber, err
}

func (s *Store) AddAccount(ctx context.Context, customerNumber string, account domain.Account) error {
	return addAccount(ctx, s.db, customerNumber, account)
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

	return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
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

	err := s.db.QueryRow(
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

	rows, err := s.db.Query(
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

	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
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
