package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type fakeStore struct {
	products []domain.Product
	err      error
	calls    int
}

func (s *fakeStore) ImportProducts(_ context.Context, products []domain.Product) error {
	s.calls++
	s.products = products

	return s.err
}

const brandName = "Mock Bank"

func product(id, extra string) []byte {
	return []byte(`{"productId": "` + id + `", "lastUpdated": "2026-08-03T01:00:00Z",
		"productCategory": "TRANS_AND_SAVINGS_ACCOUNTS", "name": "n", "description": "d", "brand": "Mock Bank"` +
		extra + `}`)
}

func TestSeedImportsAllTheProductsOfAValidSet(t *testing.T) {
	store := &fakeStore{}

	err := app.Seed(t.Context(), store, brandName, [][]byte{product("saver", ""), product("loan", "")})

	require.NoError(t, err)
	require.Len(t, store.products, 2)
	assert.Equal(t, "saver", store.products[0].ProductID)
	assert.Equal(t, "loan", store.products[1].ProductID)
}

// Verifies: DH-0229, DH-0230
func TestSeedImportsNothingAndNamesEachProductAndRuleThatFails(t *testing.T) {
	store := &fakeStore{}

	err := app.Seed(t.Context(), store, brandName, [][]byte{
		product("saver", ""),
		product("card", `, "features": [{"featureType": "OTHER"}, {"featureType": "FREE_COFFEE"}]`),
		product("", `, "interestRate": "0.05"`),
		product("loan", ""),
		product("saver", ""),
	})

	require.Error(t, err)
	assert.Zero(t, store.calls, "the store gets no products")

	message := err.Error()
	assert.Contains(t, message, `product 2 ("card")`)
	assert.Contains(t, message, "features[0]: additionalInfo is mandatory when the type is OTHER")
	assert.Contains(t, message, `features[1].featureType "FREE_COFFEE" is not a value of the CDS enumeration`)
	assert.Contains(t, message, "product 3:")
	assert.Contains(t, message, `unknown field "interestRate"`)
	assert.Contains(t, message, `productId "saver" is not unique`)
	assert.NotContains(t, message, `"loan"`)
}

// Verifies: DH-0212
func TestSeedFindsADuplicateOfAProductThatIsNotValid(t *testing.T) {
	err := app.Seed(t.Context(), &fakeStore{}, brandName, [][]byte{
		product("card", `, "features": [{"featureType": "OTHER"}]`),
		product("card", ""),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), `productId "card" is not unique`)
}

func TestSeedGivesTheErrorOfTheStore(t *testing.T) {
	store := &fakeStore{err: app.ErrNotEmpty}

	err := app.Seed(t.Context(), store, brandName, [][]byte{product("saver", "")})

	assert.ErrorIs(t, err, app.ErrNotEmpty)
}

// Verifies: DH-0217, DH-0232
func TestSeedRefusesAProductWhoseBrandIsNotTheBrandName(t *testing.T) {
	store := &fakeStore{}

	err := app.Seed(t.Context(), store, "Coastal Bank", [][]byte{product("saver", "")})

	require.Error(t, err)
	assert.Zero(t, store.calls, "the store gets no products")
	assert.Contains(t, err.Error(), `product 1 ("saver"): brand "Mock Bank" is not the brand name "Coastal Bank"`)
}

func TestCheckProductsGivesTheProductsOfAValidSet(t *testing.T) {
	products, err := app.CheckProducts(brandName, [][]byte{product("saver", ""), product("loan", "")})

	require.NoError(t, err)
	require.Len(t, products, 2)
	assert.Equal(t, "loan", products[1].ProductID)
}
