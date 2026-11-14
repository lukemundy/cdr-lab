package app

import (
	"context"
	"errors"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var ErrTransactionNotFound = errors.New("the account has no transaction with this identifier")

// TransactionQuery has inclusive time limits, and an absent amount limit matches each amount.
type TransactionQuery struct {
	AccountID      string
	Oldest, Newest time.Time
	MinAmount      *domain.Amount
	MaxAmount      *domain.Amount
	// Text matches the description or the reference, with no difference between upper case and lower case.
	Text string
	// Page starts at 1.
	Page     int
	PageSize int
}

type TransactionStore interface {
	AddTransaction(ctx context.Context, accountID string, transaction domain.Transaction) error
	// Transactions are in the sequence of their creation.
	Transactions(ctx context.Context, accountID string) ([]domain.Transaction, error)
	// FindTransactions gives one page in descending order of the effective time, then of the creation, and the count
	// of all matches.
	FindTransactions(ctx context.Context, query TransactionQuery) ([]domain.Transaction, int, error)
	// Totals has no entry for an account with no transactions.
	Totals(ctx context.Context, accountIDs []string) (map[string]domain.TransactionTotals, error)
}
