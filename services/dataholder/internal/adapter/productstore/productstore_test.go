package productstore_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/productstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

func productIDs(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()

	rows, err := pool.Query(t.Context(), "SELECT product_id FROM products ORDER BY product_id")
	require.NoError(t, err)

	var ids []string

	for rows.Next() {
		var id string
		require.NoError(t, rows.Scan(&id))

		ids = append(ids, id)
	}

	require.NoError(t, rows.Err())

	return ids
}

func product(id string) domain.Product {
	return domain.Product{
		ProductID: id, ProductCategory: domain.CategoryTransAndSavingsAccounts, Name: "Bonus Saver",
		Description: "A savings account.", Brand: "Mock Bank",
		LastUpdated: time.Date(2026, 8, 3, 1, 0, 0, 0, time.UTC),
	}
}

func TestImportProductsKeepsEachFieldOfAProduct(t *testing.T) {
	pool := newPool(t)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	saver := product("saver")
	saver.EffectiveFrom = &from
	saver.Features = []domain.Feature{{FeatureType: domain.FeatureNppPayid}}
	saver.DepositRates = []domain.DepositRate{{
		DepositRateType: domain.DepositRateVariable, Rate: "0.0155",
		ApplicationType: domain.ApplicationPeriodic, ApplicationFrequency: "P1M",
	}}

	require.NoError(t, productstore.New(pool).ImportProducts(t.Context(), []domain.Product{saver}))

	var (
		id, category, brand        string
		lastUpdated                time.Time
		effectiveFrom, effectiveTo *time.Time
		encoded                    []byte
		stored                     domain.Product
	)

	require.NoError(t, pool.QueryRow(t.Context(), `
		SELECT product_id, product_category, brand, last_updated, effective_from, effective_to, product
		FROM products`).Scan(&id, &category, &brand, &lastUpdated, &effectiveFrom, &effectiveTo, &encoded))
	assert.Equal(t, "saver", id)
	assert.Equal(t, "TRANS_AND_SAVINGS_ACCOUNTS", category)
	assert.Equal(t, "Mock Bank", brand)
	assert.True(t, lastUpdated.Equal(saver.LastUpdated))
	require.NotNil(t, effectiveFrom)
	assert.True(t, effectiveFrom.Equal(from))
	assert.Nil(t, effectiveTo)

	require.NoError(t, json.Unmarshal(encoded, &stored))
	assert.Equal(t, saver.Features, stored.Features)
	assert.Equal(t, saver.DepositRates, stored.DepositRates)
	assert.Equal(t, saver.Name, stored.Name)
}

func TestImportProductsRefusesAStoreThatHasProducts(t *testing.T) {
	pool := newPool(t)
	store := productstore.New(pool)

	require.NoError(t, store.ImportProducts(t.Context(), []domain.Product{product("saver")}))

	err := store.ImportProducts(t.Context(), []domain.Product{product("loan")})

	require.ErrorIs(t, err, app.ErrNotEmpty)
	assert.Equal(t, []string{"saver"}, productIDs(t, pool))
}

func TestImportProductsAddsNoProductWhenOneFails(t *testing.T) {
	pool := newPool(t)

	err := productstore.New(pool).ImportProducts(t.Context(), []domain.Product{product("saver"), product("saver")})

	require.Error(t, err)
	assert.Empty(t, productIDs(t, pool))
}
