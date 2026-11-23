package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// CustomerWrites runs write with stores that write in one database transaction. When write gives an error, nothing
// that it wrote stays.
type CustomerWrites func(ctx context.Context, write func(CustomerStore, TransactionStore) error) error

type LoadResult struct {
	CustomerNumber string `json:"customerNumber"`
	Created        bool   `json:"created"`
}

// LoadRecordError names the record that a write of a load refused by its position in the file, which is not a value
// of the record.
type LoadRecordError struct {
	Position string
	Err      error
}

func (e LoadRecordError) Error() string { return e.Position + ": " + e.Err.Error() }

func (e LoadRecordError) Unwrap() error { return e.Err }

type loadCustomer struct {
	customer domain.Customer
	accounts []loadAccount
}

type loadAccount struct {
	account      domain.Account
	closed       bool
	transactions []domain.Transaction
}

// Load adds all or nothing. It gives one domain.ValidationError with the errors of all records, also of a customer
// that exists, and a LoadRecordError for a record that a create refuses.
func (c Customers) Load(ctx context.Context, load domain.CustomerLoad) ([]LoadResult, error) {
	customers, err := c.checkLoad(ctx, load, c.Now().UTC())
	if err != nil {
		return nil, err
	}

	var results []LoadResult

	err = c.Writes(ctx, func(store CustomerStore, transactions TransactionStore) error {
		inTransaction := c
		inTransaction.Store, inTransaction.Transactions = store, transactions
		results = make([]LoadResult, 0, len(customers))

		for i, customer := range customers {
			result, err := inTransaction.addLoadCustomer(ctx, fmt.Sprintf("customers[%d]", i), customer)
			if err != nil {
				return err
			}

			results = append(results, result)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (c Customers) addLoadCustomer(ctx context.Context, position string, load loadCustomer) (LoadResult, error) {
	number, err := c.Store.CustomerNumberOf(ctx, load.customer.EmailAddress)
	if err == nil {
		return LoadResult{CustomerNumber: number}, nil
	}

	if !errors.Is(err, ErrCustomerNotFound) {
		return LoadResult{}, fmt.Errorf("find the customer: %w", err)
	}

	created, err := c.Create(ctx, load.customer)
	if err != nil {
		return LoadResult{}, LoadRecordError{Position: position, Err: err}
	}

	for j, account := range load.accounts {
		accountPosition := fmt.Sprintf("%s.accounts[%d]", position, j)

		added, err := c.AddAccount(ctx, created.CustomerNumber, account.account)
		if err != nil {
			return LoadResult{}, LoadRecordError{Position: accountPosition, Err: err}
		}

		for k, transaction := range account.transactions {
			_, err := c.AddTransaction(ctx, created.CustomerNumber, added.AccountID, transaction)
			if err != nil {
				return LoadResult{}, LoadRecordError{
					Position: fmt.Sprintf("%s.transactions[%d]", accountPosition, k), Err: err,
				}
			}
		}

		if account.closed {
			if _, err := c.CloseAccount(ctx, created.CustomerNumber, added.AccountID); err != nil {
				return LoadResult{}, LoadRecordError{Position: accountPosition, Err: err}
			}
		}
	}

	return LoadResult{CustomerNumber: created.CustomerNumber, Created: true}, nil
}

// checkLoad gives the records of the file, or one domain.ValidationError with the errors of all records.
func (c Customers) checkLoad(
	ctx context.Context,
	load domain.CustomerLoad,
	loadedAt time.Time,
) ([]loadCustomer, error) {
	if load.Customers == nil {
		return nil, domain.ValidationError{Errors: []domain.FieldError{
			{Field: "customers", Detail: "customers must be present"},
		}}
	}

	check := loadCheck{customers: c, loadedAt: loadedAt, products: map[string]*domain.Product{}}
	firstOfEmail := map[string]int{}
	customers := make([]loadCustomer, 0, len(*load.Customers))

	for i, file := range *load.Customers {
		position := fmt.Sprintf("customers[%d]", i)

		customer, err := file.Customer()
		if err := check.collect(position, err); err != nil {
			return nil, err
		}

		email := strings.ToLower(file.EmailAddress)
		if first, ok := firstOfEmail[email]; ok {
			check.faults = append(check.faults, domain.FieldError{
				Field:  position + ".emailAddress",
				Detail: fmt.Sprintf("%s: emailAddress is also the emailAddress of customers[%d]", position, first),
			})
		} else {
			firstOfEmail[email] = i
		}

		accounts, err := check.accounts(ctx, position, file.Accounts)
		if err != nil {
			return nil, err
		}

		customers = append(customers, loadCustomer{customer: customer, accounts: accounts})
	}

	if len(check.faults) > 0 {
		return nil, domain.ValidationError{Errors: check.faults}
	}

	return customers, nil
}

type loadCheck struct {
	customers Customers
	loadedAt  time.Time
	// products has nil for a productId that the brand does not have.
	products map[string]*domain.Product
	faults   []domain.FieldError
}

// collect keeps a domain.ValidationError at its position, and gives each other error.
func (l *loadCheck) collect(position string, err error) error {
	if invalid, ok := errors.AsType[domain.ValidationError](err); ok {
		l.faults = append(l.faults, domain.AtPosition(position, invalid)...)
		return nil
	}

	return err
}

func (l *loadCheck) accounts(ctx context.Context, position string, files []domain.LoadAccount) ([]loadAccount, error) {
	accounts := make([]loadAccount, 0, len(files))

	for j, file := range files {
		accountPosition := fmt.Sprintf("%s.accounts[%d]", position, j)

		product, err := l.product(ctx, file.ProductID)
		if err != nil {
			return nil, err
		}

		var account domain.Account

		if product == nil {
			l.faults = append(l.faults, domain.FieldError{
				Field:  accountPosition + ".productId",
				Detail: accountPosition + ": the data holder has no product with this productId",
			})
		} else {
			account, err = file.Account(*product, l.customers.BSB)
			if err := l.collect(accountPosition, err); err != nil {
				return nil, err
			}
		}

		transactions := make([]domain.Transaction, 0, len(file.Transactions))

		for k, t := range file.Transactions {
			transaction, err := t.Transaction(l.loadedAt)
			if err := l.collect(fmt.Sprintf("%s.transactions[%d]", accountPosition, k), err); err != nil {
				return nil, err
			}

			transactions = append(transactions, transaction)
		}

		accounts = append(accounts, loadAccount{account: account, closed: file.Closed(), transactions: transactions})
	}

	return accounts, nil
}

func (l *loadCheck) product(ctx context.Context, productID string) (*domain.Product, error) {
	if product, ok := l.products[productID]; ok {
		return product, nil
	}

	product, err := l.customers.Products.Product(ctx, productID)

	switch {
	case errors.Is(err, ErrProductNotFound):
		l.products[productID] = nil
	case err != nil:
		return nil, fmt.Errorf("read the product: %w", err)
	default:
		l.products[productID] = &product
	}

	return l.products[productID], nil
}
