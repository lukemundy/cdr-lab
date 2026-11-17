package app

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var (
	ErrArrangementNotFound = errors.New("the data holder has no arrangement with this identifier")
	ErrArrangementEnded    = errors.New("the arrangement ended")
)

type ConsentData struct {
	Arrangements     eventsourcing.Store
	Customers        CustomerStore
	TransactionStore TransactionStore
	Now              func() time.Time
}

const (
	defaultTransactionPeriod = 90 * 24 * time.Hour
	// Older transactions are not required consumer data (CDR Rules, Schedule 3, clause 3.2(6)(a)).
	sharedTransactionYears = 2
	// allTransactions is a page size for each transaction of the period, because the period limits the number.
	allTransactions = math.MaxInt32
)

// noNewestLimit is after each effective time, because a transaction in the future also has detail.
var noNewestLimit = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)

// TransactionFilter gives an absent time its CDS default.
type TransactionFilter struct {
	Oldest, Newest *time.Time
	MinAmount      *domain.Amount
	MaxAmount      *domain.Amount
	Text           string
	Page           int
	PageSize       int
}

type ConsentTransaction struct {
	// ID is the transactionId that the software product of the arrangement gets.
	ID          string
	Transaction domain.Transaction
}

// AccountQuery matches each account for a filter with its zero value.
type AccountQuery struct {
	Category   domain.ProductCategory
	OpenStatus domain.OpenStatus
	IsOwned    *bool
	// IDs is a list of the accountIds of the software product when it is not nil.
	IDs []string
	// Page starts at 1.
	Page     int
	PageSize int
}

// matches finds each account owned, because each account has its customer as its only owner.
func (q AccountQuery) matches(a ConsentAccount) bool {
	return (q.Category == "" || a.Account.ProductCategory == q.Category) &&
		(q.OpenStatus == "" || a.Account.OpenStatus == q.OpenStatus) &&
		(q.IsOwned == nil || *q.IsOwned) &&
		(q.IDs == nil || slices.Contains(q.IDs, a.ID))
}

// page gives one page of the matches, and the count of all matches.
func (q AccountQuery) page(accounts []ConsentAccount) ([]ConsentAccount, int) {
	matching := slices.DeleteFunc(accounts, func(a ConsentAccount) bool { return !q.matches(a) })
	start := min((q.Page-1)*q.PageSize, len(matching))
	end := min(start+q.PageSize, len(matching))

	return matching[start:end], len(matching)
}

type ConsentAccount struct {
	// ID is the accountId that the software product of the arrangement gets.
	ID      string
	Account domain.Account
}

// RequireCurrent gives ErrArrangementNotFound, or ErrArrangementEnded for an arrangement that ended for any reason.
func (c ConsentData) RequireCurrent(ctx context.Context, arrangementID string) error {
	state, _, err := ArrangementType.Load(ctx, c.Arrangements, arrangementID)
	if err != nil {
		return fmt.Errorf("load the arrangement: %w", err)
	}

	if !state.Created {
		return ErrArrangementNotFound
	}

	if state.Ended(c.Now()) {
		return ErrArrangementEnded
	}

	return nil
}

// Customer gives ErrArrangementNotFound.
func (c ConsentData) Customer(ctx context.Context, arrangementID string) (domain.Customer, error) {
	arrangement, err := c.arrangement(ctx, arrangementID)
	if err != nil {
		return domain.Customer{}, err
	}

	record, err := c.Customers.Customer(ctx, arrangement.CustomerNumber)
	if err != nil {
		return domain.Customer{}, fmt.Errorf("read the customer: %w", err)
	}

	return record.Customer, nil
}

// Accounts gives one page in the sequence of creation, the count of all matches, or ErrArrangementNotFound.
func (c ConsentData) Accounts(
	ctx context.Context,
	arrangementID string,
	query AccountQuery,
) ([]ConsentAccount, int, error) {
	_, accounts, err := c.accounts(ctx, arrangementID)
	if err != nil {
		return nil, 0, err
	}

	found, total := query.page(accounts)

	return found, total, nil
}

// Account gives ErrArrangementNotFound, or ErrAccountNotFound when the arrangement does not give the ID.
func (c ConsentData) Account(ctx context.Context, arrangementID, id string) (ConsentAccount, error) {
	_, account, err := c.consentAccount(ctx, arrangementID, id)

	return account, err
}

// Transactions gives one page and the count of all matches, or the errors of Account.
func (c ConsentData) Transactions(
	ctx context.Context,
	arrangementID, accountID string,
	filter TransactionFilter,
) ([]ConsentTransaction, int, error) {
	arrangement, account, err := c.consentAccount(ctx, arrangementID, accountID)
	if err != nil {
		return nil, 0, err
	}

	found, total, err := sharedTransactions(ctx, c.TransactionStore, c.Now(), account.Account.AccountID, filter)
	if err != nil {
		return nil, 0, err
	}

	return c.withIDs(arrangement, found), total, nil
}

// Transaction gives the errors of Account, or ErrTransactionNotFound when no transaction with detail has the ID.
func (c ConsentData) Transaction(
	ctx context.Context,
	arrangementID, accountID, id string,
) (ConsentTransaction, error) {
	arrangement, account, err := c.consentAccount(ctx, arrangementID, accountID)
	if err != nil {
		return ConsentTransaction{}, err
	}

	now := c.Now().UTC()

	// An ID cannot give its transaction, so the data holder calculates the ID of each transaction of the period.
	found, _, err := c.TransactionStore.FindTransactions(ctx, TransactionQuery{
		AccountID: account.Account.AccountID,
		Oldest:    now.AddDate(-sharedTransactionYears, 0, 0),
		Newest:    noNewestLimit,
		Page:      1,
		PageSize:  allTransactions,
	})
	if err != nil {
		return ConsentTransaction{}, fmt.Errorf("find the transactions: %w", err)
	}

	for _, transaction := range c.withIDs(arrangement, found) {
		if transaction.ID == id && transaction.Transaction.IsDetailAvailable() {
			return transaction, nil
		}
	}

	return ConsentTransaction{}, ErrTransactionNotFound
}

func (c ConsentData) withIDs(arrangement domain.Arrangement, found []domain.Transaction) []ConsentTransaction {
	transactions := make([]ConsentTransaction, 0, len(found))

	for _, transaction := range found {
		transactions = append(transactions, ConsentTransaction{
			ID: domain.ResourceID(
				arrangement.SoftwareID,
				arrangement.CustomerNumber,
				transaction.TransactionID,
			),
			Transaction: transaction,
		})
	}

	return transactions
}

// sharedTransactions is the query of Get Transactions For Account: the CDS defaults of the period and the 2-year
// limit. The online bank site also uses it, so that the site gives the same transactions as a consent.
func sharedTransactions(
	ctx context.Context,
	store TransactionStore,
	at time.Time,
	accountID string,
	filter TransactionFilter,
) ([]domain.Transaction, int, error) {
	now := at.UTC()
	newest := now

	if filter.Newest != nil {
		newest = *filter.Newest
	}

	oldest := newest.Add(-defaultTransactionPeriod)
	if filter.Oldest != nil {
		oldest = *filter.Oldest
	}

	found, total, err := store.FindTransactions(ctx, TransactionQuery{
		AccountID: accountID,
		Oldest:    latest(oldest, now.AddDate(-sharedTransactionYears, 0, 0)),
		Newest:    newest,
		MinAmount: filter.MinAmount,
		MaxAmount: filter.MaxAmount,
		Text:      filter.Text,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("find the transactions: %w", err)
	}

	return found, total, nil
}

func latest(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}

func (c ConsentData) consentAccount(
	ctx context.Context,
	arrangementID, id string,
) (domain.Arrangement, ConsentAccount, error) {
	arrangement, accounts, err := c.accounts(ctx, arrangementID)
	if err != nil {
		return domain.Arrangement{}, ConsentAccount{}, err
	}

	index := slices.IndexFunc(accounts, func(a ConsentAccount) bool { return a.ID == id })
	if index < 0 {
		return domain.Arrangement{}, ConsentAccount{}, ErrAccountNotFound
	}

	return arrangement, accounts[index], nil
}

func (c ConsentData) arrangement(ctx context.Context, arrangementID string) (domain.Arrangement, error) {
	state, _, err := ArrangementType.Load(ctx, c.Arrangements, arrangementID)
	if err != nil {
		return domain.Arrangement{}, fmt.Errorf("load the arrangement: %w", err)
	}

	if !state.Created {
		return domain.Arrangement{}, ErrArrangementNotFound
	}

	return state.Arrangement, nil
}

// accounts gives the accounts of the arrangement in the sequence of their creation.
func (c ConsentData) accounts(
	ctx context.Context,
	arrangementID string,
) (domain.Arrangement, []ConsentAccount, error) {
	arrangement, err := c.arrangement(ctx, arrangementID)
	if err != nil {
		return domain.Arrangement{}, nil, err
	}

	record, err := c.Customers.Customer(ctx, arrangement.CustomerNumber)
	if err != nil {
		return domain.Arrangement{}, nil, fmt.Errorf("read the customer: %w", err)
	}

	accounts := []ConsentAccount{}

	for _, account := range record.Accounts {
		if !slices.Contains(arrangement.AccountIDs, account.AccountID) {
			continue
		}

		accounts = append(accounts, ConsentAccount{
			ID:      domain.ResourceID(arrangement.SoftwareID, arrangement.CustomerNumber, account.AccountID),
			Account: account,
		})
	}

	return arrangement, accounts, nil
}
