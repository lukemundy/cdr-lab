package app

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// CustomerData gives the online bank site the account data of one customer with the functions of ConsentData, so
// that the site and the APIs of a consent give the same values.
type CustomerData struct {
	Customers        CustomerStore
	TransactionStore TransactionStore
	Arrangements     ArrangementReader
	Disclosures      DisclosureReader
	Now              func() time.Time
}

type CustomerAccounts struct {
	// Accounts are in the sequence of Get Accounts.
	Accounts []AccountDetail
	// Arrangements are in the sequence of their confirmation.
	Arrangements []domain.ArrangementState
}

// Accounts gives ErrCustomerNotFound.
func (d CustomerData) Accounts(ctx context.Context, customerNumber string) (CustomerAccounts, error) {
	record, err := d.Customers.Customer(ctx, customerNumber)
	if err != nil {
		return CustomerAccounts{}, fmt.Errorf("read the customer: %w", err)
	}

	balances, err := balancesOf(ctx, d.TransactionStore, record.Accounts)
	if err != nil {
		return CustomerAccounts{}, err
	}

	arrangements, err := d.Arrangements.ArrangementsOf(ctx, customerNumber)
	if err != nil {
		return CustomerAccounts{}, fmt.Errorf("read the arrangements: %w", err)
	}

	accounts := make([]AccountDetail, 0, len(record.Accounts))
	for i, a := range record.Accounts {
		accounts = append(accounts, AccountDetail{Account: a, Balance: balances[i]})
	}

	return CustomerAccounts{Accounts: accounts, Arrangements: arrangements}, nil
}

// Balance gives ErrCustomerNotFound, or ErrAccountNotFound when the account is not of the customer.
func (d CustomerData) Balance(ctx context.Context, customerNumber, accountID string) (domain.Balance, error) {
	account, err := d.account(ctx, customerNumber, accountID)
	if err != nil {
		return domain.Balance{}, err
	}

	balances, err := balancesOf(ctx, d.TransactionStore, []domain.Account{account})
	if err != nil {
		return domain.Balance{}, err
	}

	return balances[0], nil
}

// Transactions gives one page of Get Transactions For Account with no filter, and the count of all matches, or the
// errors of Balance.
func (d CustomerData) Transactions(
	ctx context.Context,
	customerNumber, accountID string,
	page, pageSize int,
) ([]domain.Transaction, int, error) {
	account, err := d.account(ctx, customerNumber, accountID)
	if err != nil {
		return nil, 0, err
	}

	return sharedTransactions(ctx, d.TransactionStore, d.Now(), account.AccountID, TransactionFilter{
		Page: page, PageSize: pageSize,
	})
}

func (d CustomerData) account(ctx context.Context, customerNumber, accountID string) (domain.Account, error) {
	record, err := d.Customers.Customer(ctx, customerNumber)
	if err != nil {
		return domain.Account{}, fmt.Errorf("read the customer: %w", err)
	}

	return accountOf(record, accountID)
}

type Consent struct {
	domain.ArrangementState
	// Accounts are in the sequence of the arrangement.
	Accounts []domain.Account
	Clusters []ClusterDisclosures
}

type ClusterDisclosures struct {
	domain.DataCluster
	// Disclosed is nil when the data holder disclosed no data of the cluster.
	Disclosed *DisclosurePeriod
}

// Consents gives each arrangement of the customer, newest confirmation first, or ErrCustomerNotFound.
func (d CustomerData) Consents(ctx context.Context, customerNumber string) ([]Consent, error) {
	record, err := d.Customers.Customer(ctx, customerNumber)
	if err != nil {
		return nil, fmt.Errorf("read the customer: %w", err)
	}

	arrangements, err := d.Arrangements.ArrangementsOf(ctx, customerNumber)
	if err != nil {
		return nil, fmt.Errorf("read the arrangements: %w", err)
	}

	ids := make([]string, 0, len(arrangements))
	for _, state := range arrangements {
		ids = append(ids, state.Arrangement.ArrangementID)
	}

	disclosed, err := d.Disclosures.FirstAndLastDisclosures(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("read the disclosures: %w", err)
	}

	consents := make([]Consent, 0, len(arrangements))
	for _, state := range slices.Backward(arrangements) {
		consents = append(consents, consentOf(state, record.Accounts, disclosed))
	}

	return consents, nil
}

func consentOf(
	state domain.ArrangementState,
	accounts []domain.Account,
	disclosed map[DisclosureKey]DisclosurePeriod,
) Consent {
	a := state.Arrangement
	consent := Consent{ArrangementState: state, Accounts: []domain.Account{}, Clusters: []ClusterDisclosures{}}

	for _, id := range a.AccountIDs {
		if i := slices.IndexFunc(
			accounts,
			func(account domain.Account) bool { return account.AccountID == id },
		); i >= 0 {
			consent.Accounts = append(consent.Accounts, accounts[i])
		}
	}

	for _, cluster := range domain.DataClusters(a.Scopes) {
		c := ClusterDisclosures{DataCluster: cluster}
		if period, ok := disclosed[DisclosureKey{ArrangementID: a.ArrangementID, Scope: cluster.Scope}]; ok {
			c.Disclosed = &period
		}

		consent.Clusters = append(consent.Clusters, c)
	}

	return consent
}
