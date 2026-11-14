package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type AccountDetail struct {
	Account domain.Account
	Balance domain.Balance
}

type ArrangementDetail struct {
	domain.ArrangementState
	// ClientRegistered is false after the registration of the client is deleted.
	ClientRegistered bool
}

type CustomerDetail struct {
	Customer domain.Customer
	// Accounts are in the sequence of their creation, and Arrangements in the sequence of their confirmation.
	Accounts     []AccountDetail
	Arrangements []ArrangementDetail
}

type CustomerDetails struct {
	Customers     CustomerStore
	Transactions  TransactionStore
	Arrangements  ArrangementReader
	Registrations RegistrationReader
}

// Detail gives ErrCustomerNotFound.
func (d CustomerDetails) Detail(ctx context.Context, customerNumber string) (CustomerDetail, error) {
	record, err := d.Customers.Customer(ctx, customerNumber)
	if err != nil {
		return CustomerDetail{}, fmt.Errorf("read the customer: %w", err)
	}

	balances, err := balancesOf(ctx, d.Transactions, record.Accounts)
	if err != nil {
		return CustomerDetail{}, err
	}

	arrangements, err := d.Arrangements.ArrangementsOf(ctx, customerNumber)
	if err != nil {
		return CustomerDetail{}, fmt.Errorf("read the arrangements: %w", err)
	}

	detail := CustomerDetail{
		Customer:     record.Customer,
		Accounts:     make([]AccountDetail, 0, len(record.Accounts)),
		Arrangements: make([]ArrangementDetail, 0, len(arrangements)),
	}
	for i, a := range record.Accounts {
		detail.Accounts = append(detail.Accounts, AccountDetail{Account: a, Balance: balances[i]})
	}

	for _, a := range arrangements {
		registered, err := d.registered(ctx, a.Arrangement.ClientID)
		if err != nil {
			return CustomerDetail{}, err
		}

		detail.Arrangements = append(
			detail.Arrangements,
			ArrangementDetail{ArrangementState: a, ClientRegistered: registered},
		)
	}

	return detail, nil
}

func (d CustomerDetails) registered(ctx context.Context, clientID string) (bool, error) {
	_, err := d.Registrations.Registration(ctx, clientID)
	if errors.Is(err, ErrRegistrationNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("read the registration of the client: %w", err)
	}

	return true, nil
}

// TransactionPage gives each transaction of the account, with no time limit, in the sequence of
// TransactionStore.FindTransactions. It gives ErrCustomerNotFound and ErrAccountNotFound.
func (d CustomerDetails) TransactionPage(
	ctx context.Context,
	customerNumber, accountID string,
	page, pageSize int,
) ([]domain.Transaction, int, error) {
	record, err := d.Customers.Customer(ctx, customerNumber)
	if err != nil {
		return nil, 0, fmt.Errorf("read the customer: %w", err)
	}

	if _, err := accountOf(record, accountID); err != nil {
		return nil, 0, err
	}

	found, total, err := d.Transactions.FindTransactions(ctx, TransactionQuery{
		AccountID: accountID,
		Newest:    noNewestLimit,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("find the transactions: %w", err)
	}

	return found, total, nil
}
