package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var (
	ErrCustomerNotFound     = errors.New("the data holder has no customer with this customer number")
	ErrCustomerNumberUsed   = errors.New("a customer has this customer number")
	ErrEmailAddressUsed     = errors.New("a customer has this email address")
	ErrNoFreeCustomerNumber = errors.New("the data holder found no customer number that is not in use")
	ErrAccountExists        = errors.New("an account has this BSB and account number")
	ErrAccountNotFound      = errors.New("the customer has no account with this identifier")
	ErrEndsChanged          = errors.New("the close ends different arrangements from the arrangements of the request")
)

// CloseCheck runs in the close, after the change and before the commit. An error undoes the close.
type CloseCheck func(ctx context.Context, noOpenAccount bool) error

type CustomerRecord struct {
	Customer domain.Customer
	// Accounts are in the sequence of their creation.
	Accounts []domain.Account
}

type CustomerStore interface {
	// AddCustomer gives ErrCustomerNumberUsed when a customer has the customer number, and ErrEmailAddressUsed when a
	// customer has the email address with no difference for upper and lower case.
	AddCustomer(ctx context.Context, customer domain.Customer) error
	// CustomerNumberOf finds the email address with no difference for upper and lower case. It gives
	// ErrCustomerNotFound when no customer has the email address.
	CustomerNumberOf(ctx context.Context, emailAddress string) (string, error)
	// AddAccount gives ErrCustomerNotFound when no customer has the customer number, and ErrAccountExists when an
	// account of the brand has the BSB and the account number.
	AddAccount(ctx context.Context, customerNumber string, account domain.Account) error
	// AddAccountWithTransaction adds the account and its first transaction together, or neither, with the errors of
	// AddAccount.
	AddAccountWithTransaction(
		ctx context.Context,
		customerNumber string,
		account domain.Account,
		transaction domain.Transaction,
	) error
	// Customer gives ErrCustomerNotFound when no customer has the customer number.
	Customer(ctx context.Context, customerNumber string) (CustomerRecord, error)
	// CloseAccount gives domain.ErrAccountClosed for an account that is closed, and true when the customer of the
	// account has no open account after the close. A nil check accepts each close.
	CloseAccount(ctx context.Context, account domain.Account, closedAt time.Time, check CloseCheck) (bool, error)
}

type ArrangementEnder interface {
	EndForIneligibleCustomer(ctx context.Context, customerNumber string, closedAt time.Time) error
}

type IneligibleCustomerArrangements interface {
	ArrangementEnder
	IneligibilityEnds(ctx context.Context, customerNumber string, at time.Time) ([]domain.ArrangementState, error)
}

// maxCustomerNumbers is the number of customer numbers that Create tries. A second number is necessary only when a
// random customer number is in use.
const maxCustomerNumbers = 3

type Customers struct {
	Store        CustomerStore
	Arrangements IneligibleCustomerArrangements
	Transactions TransactionStore
	// Random gives a value that a person cannot predict.
	Random func() uint64
	// NewID gives a random UUID, which becomes the identifier of an account.
	NewID func() string
	Now   func() time.Time
	// Products gives the product of an open on the online bank site.
	Products ProductReader
	// BSB is the bsb of the brand directory, for the accounts that a consumer opens.
	BSB string
}

// Create gives a domain.ValidationError when the customer is not valid, ErrEmailAddressUsed and
// ErrNoFreeCustomerNumber.
func (c Customers) Create(ctx context.Context, customer domain.Customer) (domain.Customer, error) {
	if err := customer.Validate(); err != nil {
		return domain.Customer{}, err
	}

	customer.LastUpdated = c.Now().UTC()

	for range maxCustomerNumbers {
		customer.CustomerNumber = domain.NewCustomerNumber(c.Random())

		err := c.Store.AddCustomer(ctx, customer)
		if err == nil {
			return customer, nil
		}

		if !errors.Is(err, ErrCustomerNumberUsed) {
			return domain.Customer{}, fmt.Errorf("add the customer: %w", err)
		}
	}

	return domain.Customer{}, fmt.Errorf("%w after %d attempts", ErrNoFreeCustomerNumber, maxCustomerNumbers)
}

// Identify gives ErrCustomerNotFound when no customer has the email address or the customer number.
func (c Customers) Identify(ctx context.Context, identifier string) (CustomerRecord, error) {
	return identify(ctx, c.Store, identifier)
}

// identify takes a value with "@" as an email address and a different value as a customer number, because a customer
// number has only digits.
func identify(ctx context.Context, store CustomerStore, identifier string) (CustomerRecord, error) {
	identifier = strings.TrimSpace(identifier)
	if !strings.Contains(identifier, "@") {
		return store.Customer(ctx, identifier)
	}

	customerNumber, err := store.CustomerNumberOf(ctx, identifier)
	if err != nil {
		return CustomerRecord{}, err
	}

	return store.Customer(ctx, customerNumber)
}

// AddAccount gives the account with its new identifier, or a domain.ValidationError and the errors of the store.
func (c Customers) AddAccount(
	ctx context.Context,
	customerNumber string,
	account domain.Account,
) (domain.Account, error) {
	account, err := c.newAccount(account)
	if err != nil {
		return domain.Account{}, err
	}

	if err := c.Store.AddAccount(ctx, customerNumber, account); err != nil {
		return domain.Account{}, fmt.Errorf("add the account: %w", err)
	}

	return account, nil
}

func (c Customers) newAccount(account domain.Account) (domain.Account, error) {
	if err := account.Validate(); err != nil {
		return domain.Account{}, err
	}

	account.AccountID = c.NewID()
	account.OpenStatus = domain.OpenStatusOpen
	account.CreatedAt = c.Now().UTC()

	return account, nil
}

// CloseAccount gives ErrCustomerNotFound, ErrAccountNotFound and domain.ErrAccountClosed.
func (c Customers) CloseAccount(ctx context.Context, customerNumber, accountID string) (domain.Account, error) {
	return c.closeAccount(ctx, customerNumber, accountID, func(time.Time) CloseCheck { return nil })
}

// CloseAccountThatEnds closes the account only when the close ends the arrangements with the IDs in ends, and no
// other arrangement. It gives ErrEndsChanged and the errors of CloseAccount.
func (c Customers) CloseAccountThatEnds(
	ctx context.Context,
	customerNumber, accountID string,
	ends []string,
) (domain.Account, error) {
	return c.closeAccount(ctx, customerNumber, accountID, func(closedAt time.Time) CloseCheck {
		return func(ctx context.Context, noOpenAccount bool) error {
			endsNow := []string{}

			if noOpenAccount {
				states, err := c.Arrangements.IneligibilityEnds(ctx, customerNumber, closedAt)
				if err != nil {
					return err
				}

				for _, state := range states {
					endsNow = append(endsNow, state.Arrangement.ArrangementID)
				}
			}

			if !sameSet(ends, endsNow) {
				return ErrEndsChanged
			}

			return nil
		}
	})
}

func sameSet(a, b []string) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)

	return slices.Equal(slices.Compact(a), slices.Compact(b))
}

func (c Customers) closeAccount(
	ctx context.Context,
	customerNumber, accountID string,
	checkAt func(closedAt time.Time) CloseCheck,
) (domain.Account, error) {
	account, err := c.account(ctx, customerNumber, accountID)
	if err != nil {
		return domain.Account{}, err
	}

	closed, err := account.Close()
	if err != nil {
		return domain.Account{}, err
	}

	closedAt := c.Now().UTC()

	noOpenAccount, err := c.Store.CloseAccount(ctx, closed, closedAt, checkAt(closedAt))
	if err != nil {
		return domain.Account{}, fmt.Errorf("close the account: %w", err)
	}

	if !noOpenAccount {
		return closed, nil
	}

	// The close stays: the job that the store inserted with the close ends the arrangements later.
	if err := c.Arrangements.EndForIneligibleCustomer(ctx, customerNumber, closedAt); err != nil {
		slog.ErrorContext(ctx, "the arrangements of a customer who is not eligible did not end",
			"error.type", fmt.Sprintf("%T", err))
	}

	return closed, nil
}

// AddTransaction gives a domain.ValidationError, ErrCustomerNotFound, ErrAccountNotFound and
// domain.ErrAccountClosed.
func (c Customers) AddTransaction(
	ctx context.Context,
	customerNumber, accountID string,
	transaction domain.Transaction,
) (domain.Transaction, error) {
	if err := transaction.Validate(); err != nil {
		return domain.Transaction{}, err
	}

	account, err := c.account(ctx, customerNumber, accountID)
	if err != nil {
		return domain.Transaction{}, err
	}

	if err := account.AcceptTransaction(); err != nil {
		return domain.Transaction{}, err
	}

	transaction = transaction.InUTC()
	transaction.TransactionID = c.NewID()

	if err := c.Transactions.AddTransaction(ctx, accountID, transaction); err != nil {
		return domain.Transaction{}, fmt.Errorf("add the transaction: %w", err)
	}

	return transaction, nil
}

// AccountTransactions gives ErrCustomerNotFound and ErrAccountNotFound.
func (c Customers) AccountTransactions(
	ctx context.Context,
	customerNumber, accountID string,
) ([]domain.Transaction, error) {
	if _, err := c.account(ctx, customerNumber, accountID); err != nil {
		return nil, err
	}

	transactions, err := c.Transactions.Transactions(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("read the transactions: %w", err)
	}

	return transactions, nil
}

func (c Customers) account(ctx context.Context, customerNumber, accountID string) (domain.Account, error) {
	record, err := c.Store.Customer(ctx, customerNumber)
	if err != nil {
		return domain.Account{}, fmt.Errorf("read the customer: %w", err)
	}

	return accountOf(record, accountID)
}

func accountOf(record CustomerRecord, accountID string) (domain.Account, error) {
	index := slices.IndexFunc(record.Accounts, func(a domain.Account) bool { return a.AccountID == accountID })
	if index < 0 {
		return domain.Account{}, ErrAccountNotFound
	}

	return record.Accounts[index], nil
}

// Customer gives ErrCustomerNotFound when no customer has the customer number.
func (c Customers) Customer(ctx context.Context, customerNumber string) (CustomerRecord, error) {
	return c.Store.Customer(ctx, customerNumber)
}
