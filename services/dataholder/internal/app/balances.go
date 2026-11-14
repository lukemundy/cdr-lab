package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type ConsentBalance struct {
	// ID is the accountId that the software product of the arrangement gets.
	ID      string
	Balance domain.Balance
}

// UnknownAccountsError has each requested accountId that the arrangement does not give, one time each, in the
// sequence of the request.
type UnknownAccountsError struct {
	IDs []string
}

func (e UnknownAccountsError) Error() string {
	return fmt.Sprintf("the arrangement does not give %d of the requested accounts", len(e.IDs))
}

// Balance gives the errors of Account.
func (c ConsentData) Balance(ctx context.Context, arrangementID, id string) (ConsentBalance, error) {
	account, err := c.Account(ctx, arrangementID, id)
	if err != nil {
		return ConsentBalance{}, err
	}

	balances, err := c.withBalances(ctx, []ConsentAccount{account})
	if err != nil {
		return ConsentBalance{}, err
	}

	return balances[0], nil
}

// Balances gives the balances of the accounts that Accounts gives, or an UnknownAccountsError when query.IDs has an
// accountId that the arrangement does not give.
func (c ConsentData) Balances(
	ctx context.Context,
	arrangementID string,
	query AccountQuery,
) ([]ConsentBalance, int, error) {
	_, accounts, err := c.accounts(ctx, arrangementID)
	if err != nil {
		return nil, 0, err
	}

	if unknown := unknownIDs(accounts, query.IDs); len(unknown) > 0 {
		return nil, 0, UnknownAccountsError{IDs: unknown}
	}

	found, total := query.page(accounts)

	balances, err := c.withBalances(ctx, found)

	return balances, total, err
}

func unknownIDs(accounts []ConsentAccount, ids []string) []string {
	var unknown []string

	for _, id := range ids {
		known := slices.ContainsFunc(accounts, func(a ConsentAccount) bool { return a.ID == id })
		if !known && !slices.Contains(unknown, id) {
			unknown = append(unknown, id)
		}
	}

	return unknown
}

func (c ConsentData) withBalances(ctx context.Context, accounts []ConsentAccount) ([]ConsentBalance, error) {
	internal := make([]domain.Account, 0, len(accounts))
	for _, a := range accounts {
		internal = append(internal, a.Account)
	}

	found, err := balancesOf(ctx, c.TransactionStore, internal)
	if err != nil {
		return nil, err
	}

	balances := make([]ConsentBalance, 0, len(accounts))
	for i, a := range accounts {
		balances = append(balances, ConsentBalance{ID: a.ID, Balance: found[i]})
	}

	return balances, nil
}

// balancesOf gives the balance of each account, in the sequence of accounts. The online bank site also uses it, so
// that the site gives the same balances as a consent.
func balancesOf(ctx context.Context, store TransactionStore, accounts []domain.Account) ([]domain.Balance, error) {
	ids := make([]string, 0, len(accounts))
	for _, a := range accounts {
		ids = append(ids, a.AccountID)
	}

	totals, err := store.Totals(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("sum the transactions: %w", err)
	}

	balances := make([]domain.Balance, 0, len(accounts))
	for _, a := range accounts {
		balances = append(balances, a.Balance(totals[a.AccountID]))
	}

	return balances, nil
}
