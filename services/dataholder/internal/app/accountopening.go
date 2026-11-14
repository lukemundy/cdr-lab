package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var ErrNoFreeAccountNumber = errors.New("the data holder found no account number that is not in use")

// maxAccountNumbers is the number of account numbers that OpenAccount tries, as for a customer number.
const maxAccountNumbers = 3

// CurrentProducts is the query of Get Products with effective=CURRENT at the time of the request, with all products
// on one page. The online bank site and an open use it, so that they have the rule and the products of Get Products.
func CurrentProducts(at time.Time) ProductQuery {
	return ProductQuery{Effective: EffectiveCurrent, At: at, Page: 1}
}

// OpenAccount gives ErrProductNotFound for a product that is not current, domain.ValueNotForProductError and
// domain.ValidationError for the input, ErrNoFreeAccountNumber and the errors of AddAccount.
func (c Customers) OpenAccount(
	ctx context.Context,
	customerNumber, productID string,
	input domain.OpeningInput,
) (domain.Account, error) {
	query := CurrentProducts(c.Now())
	query.ProductID = productID

	found, _, err := c.Products.Products(ctx, query)
	if err != nil {
		return domain.Account{}, fmt.Errorf("read the product: %w", err)
	}

	if len(found) != 1 {
		return domain.Account{}, ErrProductNotFound
	}

	product := found[0]

	opening, err := input.Opening(product.ProductCategory)
	if err != nil {
		return domain.Account{}, err
	}

	for range maxAccountNumbers {
		account, err := c.addOpenedAccount(ctx, customerNumber, opening.LoanAmount, domain.Account{
			DisplayName:     product.Name,
			ProductCategory: product.ProductCategory,
			ProductName:     product.Name,
			BSB:             c.BSB,
			AccountNumber:   domain.NewAccountNumber(c.Random()),
			CreditLimit:     opening.CreditLimit,
		})
		if err == nil {
			return account, nil
		}

		if !errors.Is(err, ErrAccountExists) {
			return domain.Account{}, err
		}
	}

	return domain.Account{}, fmt.Errorf("%w after %d attempts", ErrNoFreeAccountNumber, maxAccountNumbers)
}

// addOpenedAccount adds a drawdown only for a loan amount, which only a loan has.
func (c Customers) addOpenedAccount(
	ctx context.Context,
	customerNumber string,
	loanAmount domain.Amount,
	account domain.Account,
) (domain.Account, error) {
	if loanAmount == 0 {
		return c.AddAccount(ctx, customerNumber, account)
	}

	account, err := c.newAccount(account)
	if err != nil {
		return domain.Account{}, err
	}

	drawdown := domain.Drawdown(loanAmount, account.CreatedAt)
	drawdown.TransactionID = c.NewID()

	if err := c.Store.AddAccountWithTransaction(ctx, customerNumber, account, drawdown); err != nil {
		return domain.Account{}, fmt.Errorf("add the account and its drawdown: %w", err)
	}

	return account, nil
}
