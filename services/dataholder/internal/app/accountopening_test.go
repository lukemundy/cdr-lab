package app_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// currentProduct gives the products of the query only when it is the current-product query of the site, at now.
type currentProduct struct {
	products map[string]domain.Product
	query    app.ProductQuery
}

func (c *currentProduct) Products(_ context.Context, query app.ProductQuery) ([]domain.Product, int, error) {
	c.query = query

	want := app.CurrentProducts(now())
	want.ProductID = query.ProductID

	p, ok := c.products[query.ProductID]
	if !reflect.DeepEqual(query, want) || !ok {
		return nil, 0, nil
	}

	return []domain.Product{p}, 1, nil
}

func (c *currentProduct) Product(context.Context, string) (domain.Product, error) {
	return domain.Product{}, app.ErrProductNotFound
}

func productsOf(products ...domain.Product) *currentProduct {
	found := map[string]domain.Product{}
	for _, p := range products {
		found[p.ProductID] = p
	}

	return &currentProduct{products: found}
}

func saver() domain.Product {
	return domain.Product{
		ProductID: "saver", ProductCategory: domain.CategoryTransAndSavingsAccounts, Name: "Bonus Saver",
	}
}

// numbersInUse gives ErrAccountExists for each account number in used.
type numbersInUse struct {
	*fakeCustomerStore
	used map[string]bool
}

func (n numbersInUse) AddAccount(ctx context.Context, customerNumber string, account domain.Account) error {
	if n.used[account.AccountNumber] {
		return app.ErrAccountExists
	}

	return n.fakeCustomerStore.AddAccount(ctx, customerNumber, account)
}

func (n numbersInUse) AddAccountWithTransaction(
	ctx context.Context,
	customerNumber string,
	account domain.Account,
	transaction domain.Transaction,
) error {
	if n.used[account.AccountNumber] {
		return app.ErrAccountExists
	}

	return n.fakeCustomerStore.AddAccountWithTransaction(ctx, customerNumber, account, transaction)
}

func opening(store app.CustomerStore, products app.ProductReader, values ...uint64) app.Customers {
	return app.Customers{
		Store: store, Products: products, BSB: "372100", Random: random(values...), NewID: newID, Now: now,
	}
}

// Verifies: DH-2225
func TestAnOpenGivesAnAccountOfTheProductWithTheBSBOfTheBrandAndARandomAccountNumber(t *testing.T) {
	store := newFakeCustomerStore()
	store.customers["00000001"] = customer()

	opened, err := opening(
		store,
		productsOf(saver()),
		12_345_678,
	).OpenAccount(t.Context(), "00000001", "saver", domain.OpeningInput{})

	require.NoError(t, err)

	want := domain.Account{
		AccountID: "account-1", DisplayName: "Bonus Saver", ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName: "Bonus Saver", BSB: "372100", AccountNumber: "12345678", OpenStatus: domain.OpenStatusOpen,
		CreatedAt: time.Date(2026, 10, 15, 13, 30, 0, 0, time.UTC),
	}
	assert.Equal(t, want, opened)
	assert.Equal(t, []domain.Account{want}, store.accounts["00000001"])
}

// Verifies: DH-2210, DH-2223
func TestEachDepositCategoryOpens(t *testing.T) {
	for _, category := range domain.DepositCategories() {
		t.Run(string(category), func(t *testing.T) {
			store := newFakeCustomerStore()
			product := saver()
			product.ProductCategory = category

			opened, err := opening(
				store,
				productsOf(product),
				1,
			).OpenAccount(t.Context(), "00000001", "saver", domain.OpeningInput{})

			require.NoError(t, err)
			assert.Equal(t, category, opened.ProductCategory)
		})
	}
}

// Verifies: DH-2224
func TestAnOpenUsesTheCurrentProductQueryOfTheSite(t *testing.T) {
	products := productsOf(saver())

	_, err := opening(
		newFakeCustomerStore(),
		products,
		1,
	).OpenAccount(t.Context(), "00000001", "saver", domain.OpeningInput{})

	require.NoError(t, err)

	want := app.CurrentProducts(now())
	want.ProductID = "saver"
	assert.Equal(t, want, products.query)
}

func TestTheCurrentProductQueryGivesEachCurrentProductOnOnePage(t *testing.T) {
	assert.Equal(t, app.ProductQuery{Effective: app.EffectiveCurrent, At: now(), Page: 1}, app.CurrentProducts(now()))
}

// Verifies: DH-2224
func TestAnOpenOfAProductThatIsNotCurrentOpensNoAccount(t *testing.T) {
	store := newFakeCustomerStore()

	_, err := opening(store, productsOf(saver()), 1).OpenAccount(t.Context(), "00000001", "gone", domain.OpeningInput{})

	require.ErrorIs(t, err, app.ErrProductNotFound)
	assert.Empty(t, store.accounts)
}

func card() domain.Product {
	return domain.Product{ProductID: "card", ProductCategory: domain.CategoryCredAndChrgCards, Name: "Glide Card"}
}

func loan() domain.Product {
	return domain.Product{ProductID: "loan", ProductCategory: domain.CategoryPersLoans, Name: "Bloom Personal Loan"}
}

// Verifies: DH-2225
func TestAnOpenOfRevolvingCreditGivesTheAccountTheCreditLimitAndNoTransaction(t *testing.T) {
	for _, category := range domain.RevolvingCategories() {
		t.Run(string(category), func(t *testing.T) {
			store := newFakeCustomerStore()
			product := card()
			product.ProductCategory = category

			opened, err := opening(store, productsOf(product), 1).OpenAccount(
				t.Context(), "00000001", "card", domain.OpeningInput{CreditLimit: new("5000.00")},
			)

			require.NoError(t, err)
			assert.Equal(t, category, opened.ProductCategory)
			assert.Equal(t, domain.Amount(500_000), opened.CreditLimit)
			assert.Equal(t, []domain.Account{opened}, store.accounts["00000001"])
			assert.Empty(t, store.transactions)
		})
	}
}

// Verifies: DH-2227, DH-2228
func TestAnOpenOfALoanAddsTheAccountWithADrawdownOfTheLoanAmountAndNoCreditLimit(t *testing.T) {
	for _, category := range domain.LoanCategories() {
		t.Run(string(category), func(t *testing.T) {
			store := newFakeCustomerStore()
			product := loan()
			product.ProductCategory = category

			opened, err := opening(store, productsOf(product), 1).OpenAccount(
				t.Context(), "00000001", "loan", domain.OpeningInput{LoanAmount: new("20000.00")},
			)

			require.NoError(t, err)
			assert.Equal(t, domain.Amount(0), opened.CreditLimit)
			assert.Equal(t, []domain.Account{opened}, store.accounts["00000001"])

			drawdown := domain.Drawdown(2_000_000, now().UTC())
			drawdown.TransactionID = "account-1"
			assert.Equal(t, map[string][]domain.Transaction{opened.AccountID: {drawdown}}, store.transactions)
		})
	}
}

// Verifies: DH-2222
func TestAnOpenWithAValueThatIsNotValidOpensNoAccount(t *testing.T) {
	tests := map[string]struct {
		productID string
		input     domain.OpeningInput
		want      any
	}{
		"credit limit of a deposit": {
			productID: "saver", input: domain.OpeningInput{CreditLimit: new("1.00")},
			want: &domain.ValueNotForProductError{},
		},
		"zero credit limit": {
			productID: "card", input: domain.OpeningInput{CreditLimit: new("0.00")}, want: &domain.ValidationError{},
		},
		"absent loan amount": {productID: "loan", want: &domain.ValidationError{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			store := newFakeCustomerStore()

			_, err := opening(store, productsOf(saver(), card(), loan()), 1).OpenAccount(
				t.Context(), "00000001", tt.productID, tt.input,
			)

			require.ErrorAs(t, err, tt.want)
			assert.Empty(t, store.accounts)
			assert.Empty(t, store.transactions)
		})
	}
}

// Verifies: DH-2226
func TestAnAccountNumberThatIsInUseIsReplacedAlsoForALoan(t *testing.T) {
	store := numbersInUse{fakeCustomerStore: newFakeCustomerStore(), used: map[string]bool{"00000001": true}}

	opened, err := opening(store, productsOf(loan()), 1, 2).OpenAccount(
		t.Context(), "00000001", "loan", domain.OpeningInput{LoanAmount: new("1.00")},
	)

	require.NoError(t, err)
	assert.Equal(t, "00000002", opened.AccountNumber)
	assert.Len(t, store.transactions[opened.AccountID], 1)
}

// Verifies: DH-2226
func TestAnAccountNumberThatIsInUseIsReplaced(t *testing.T) {
	store := numbersInUse{fakeCustomerStore: newFakeCustomerStore(), used: map[string]bool{"00000001": true}}

	opened, err := opening(
		store,
		productsOf(saver()),
		1,
		2,
	).OpenAccount(t.Context(), "00000001", "saver", domain.OpeningInput{})

	require.NoError(t, err)
	assert.Equal(t, "00000002", opened.AccountNumber)
}

// Verifies: DH-2226
func TestAnOpenFailsWhenEachOf3AccountNumbersIsInUse(t *testing.T) {
	store := numbersInUse{
		fakeCustomerStore: newFakeCustomerStore(),
		used:              map[string]bool{"00000001": true, "00000002": true, "00000003": true},
	}

	_, err := opening(
		store,
		productsOf(saver()),
		1,
		2,
		3,
		4,
	).OpenAccount(t.Context(), "00000001", "saver", domain.OpeningInput{})

	require.ErrorIs(t, err, app.ErrNoFreeAccountNumber)
	require.NotErrorIs(t, err, app.ErrAccountExists)
	assert.Empty(t, store.accounts)
}

func TestAnOpenGivesTheErrorOfTheStore(t *testing.T) {
	store := newFakeCustomerStore()
	store.err = app.ErrCustomerNotFound

	_, err := opening(
		store,
		productsOf(saver()),
		1,
	).OpenAccount(t.Context(), "00000001", "saver", domain.OpeningInput{})

	require.ErrorIs(t, err, app.ErrCustomerNotFound)
}
