package productstore_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/productstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var (
	requestTime = time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	hourBefore  = requestTime.Add(-time.Hour)
	hourAfter   = requestTime.Add(time.Hour)
)

func allProducts() app.ProductQuery {
	return app.ProductQuery{Effective: app.EffectiveAll, At: requestTime, Page: 1, PageSize: 25}
}

func storeWith(t *testing.T, products ...domain.Product) *productstore.Store {
	t.Helper()

	store := productstore.New(newPool(t))
	require.NoError(t, store.ImportProducts(t.Context(), products))

	return store
}

func ids(products []domain.Product) []string {
	list := make([]string, 0, len(products))
	for _, p := range products {
		list = append(list, p.ProductID)
	}

	return list
}

func updated(id string, lastUpdated time.Time) domain.Product {
	p := product(id)
	p.LastUpdated = lastUpdated

	return p
}

func effective(id string, from, to *time.Time) domain.Product {
	p := product(id)
	p.EffectiveFrom, p.EffectiveTo = from, to

	return p
}

// Verifies: DH-0241, DH-0242
func TestProductsGivesTheNewestLastUpdatedFirstAndThenTheOrderOfProductID(t *testing.T) {
	store := storeWith(
		t,
		updated("b", hourBefore),
		updated("d", requestTime),
		updated("a", hourBefore),
		updated("c", hourAfter),
	)

	page, total, err := store.Products(t.Context(), allProducts())

	require.NoError(t, err)
	assert.Equal(t, []string{"c", "d", "a", "b"}, ids(page))
	assert.Equal(t, 4, total)
}

func TestProductsGivesEachFieldOfAProduct(t *testing.T) {
	saver := effective("saver", &hourBefore, &hourAfter)
	saver.IsTailored = true
	saver.Features = []domain.Feature{{FeatureType: domain.FeatureNppPayid}}
	store := storeWith(t, saver)

	page, _, err := store.Products(t.Context(), allProducts())

	require.NoError(t, err)
	assert.Equal(t, []domain.Product{saver}, page)
}

// Verifies: DH-0219, DH-0220, DH-0238
func TestProductsGivesTheProductsThatAreEffectiveAtTheTimeOfTheQuery(t *testing.T) {
	store := storeWith(
		t,
		effective("always", nil, nil),
		effective("ended", &hourBefore, &hourBefore),
		effective("ends-now", nil, &requestTime),
		effective("ends-later", nil, &hourAfter),
		effective("started", &hourBefore, nil),
		effective("starts-now", &requestTime, &hourAfter),
		effective("starts-later", &hourAfter, nil),
	)

	tests := map[app.Effective][]string{
		app.EffectiveCurrent: {"always", "ends-later", "started", "starts-now"},
		app.EffectiveFuture:  {"starts-later"},
		app.EffectiveAll: {
			"always", "ended", "ends-later", "ends-now", "started", "starts-later", "starts-now",
		},
	}
	for filter, want := range tests {
		t.Run(string(filter), func(t *testing.T) {
			query := allProducts()
			query.Effective = filter

			page, total, err := store.Products(t.Context(), query)

			require.NoError(t, err)
			assert.Equal(t, want, ids(page))
			assert.Equal(t, len(want), total)
		})
	}
}

// Verifies: DH-0237
func TestProductsGivesOnlyTheProductsThatMatchEachFilter(t *testing.T) {
	loan := updated("loan", hourAfter)
	loan.ProductCategory = domain.CategoryPersLoans
	otherBrand := updated("other-brand", hourAfter)
	otherBrand.Brand = "Mock Bank Private"
	store := storeWith(t, loan, otherBrand, updated("new", hourAfter), updated("old", requestTime))

	tests := map[string]struct {
		filter func(*app.ProductQuery)
		want   []string
	}{
		"updated-since gives a later lastUpdated only": {
			func(q *app.ProductQuery) { q.UpdatedSince = &requestTime },
			[]string{"loan", "new", "other-brand"},
		},
		"brand": {
			func(q *app.ProductQuery) { q.Brand = "Mock Bank" },
			[]string{"loan", "new", "old"},
		},
		"brand is case sensitive": {
			func(q *app.ProductQuery) { q.Brand = "mock bank" },
			[]string{},
		},
		"category": {
			func(q *app.ProductQuery) { q.Category = domain.CategoryTransAndSavingsAccounts },
			[]string{"new", "other-brand", "old"},
		},
		"product ID": {
			func(q *app.ProductQuery) { q.ProductID = "loan" },
			[]string{"loan"},
		},
		"product ID is case sensitive": {
			func(q *app.ProductQuery) { q.ProductID = "Loan" },
			[]string{},
		},
		"all filters": {
			func(q *app.ProductQuery) {
				q.UpdatedSince = &requestTime
				q.Brand = "Mock Bank"
				q.Category = domain.CategoryTransAndSavingsAccounts
			},
			[]string{"new"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			query := allProducts()
			tt.filter(&query)

			page, total, err := store.Products(t.Context(), query)

			require.NoError(t, err)
			assert.Equal(t, tt.want, ids(page))
			assert.Equal(t, len(tt.want), total)
		})
	}
}

func TestProductsGivesOnePageAndTheTotalOfAllPages(t *testing.T) {
	store := storeWith(t, product("a"), product("b"), product("c"), product("d"), product("e"))

	tests := map[int][]string{1: {"a", "b"}, 2: {"c", "d"}, 3: {"e"}, 4: {}}
	for number, want := range tests {
		query := allProducts()
		query.Page, query.PageSize = number, 2

		page, total, err := store.Products(t.Context(), query)

		require.NoError(t, err)
		assert.Equal(t, want, ids(page), "page %d", number)
		assert.Equal(t, 5, total)
	}
}

func TestProductsGivesAllProductsWithAPageSizeOfZero(t *testing.T) {
	store := storeWith(t, product("a"), product("b"), product("c"))
	query := allProducts()
	query.PageSize = 0

	page, total, err := store.Products(t.Context(), query)

	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, ids(page))
	assert.Equal(t, 3, total)
}

func TestProductGivesTheProductWithTheID(t *testing.T) {
	loan := effective("loan", nil, &hourBefore)
	loan.LendingRates = []domain.LendingRate{{
		LendingRateType: domain.LendingRateVariable, Rate: "0.0899", ApplicationType: domain.ApplicationPeriodic,
		ApplicationFrequency: "P1M", RepaymentType: domain.RepaymentPrincipalAndInterest,
		LoanPurpose: domain.PurposeOwnerOccupied,
	}}
	store := storeWith(t, product("saver"), loan)

	got, err := store.Product(t.Context(), "loan")

	require.NoError(t, err)
	assert.Equal(t, loan, got)
}

func TestProductReturnsNotFoundForAnIDThatNoProductHas(t *testing.T) {
	store := storeWith(t, product("saver"))

	_, err := store.Product(t.Context(), "Saver")

	require.ErrorIs(t, err, app.ErrProductNotFound)
}
