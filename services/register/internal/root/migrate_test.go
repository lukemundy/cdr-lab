package root

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

func migrationsBefore(t *testing.T, name string) fs.FS {
	t.Helper()

	files, err := fs.Glob(migrations.FS, "*.sql")
	require.NoError(t, err)

	before := fstest.MapFS{}

	for _, file := range files {
		if file >= name {
			continue
		}

		data, err := fs.ReadFile(migrations.FS, file)
		require.NoError(t, err)

		before[file] = &fstest.MapFile{Data: data}
	}

	return before
}

// The event store has no projections, because the projections of this build need the new schema.
func databaseBefore(t *testing.T, name string, participants domain.Participants) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrationsBefore(t, name)))
	require.NoError(t, app.Seed(t.Context(), eventstore.New(pool), participants))

	return pool
}

func TestMigrateFillsTheDataRecipientColumnsOfRowsThatExist(t *testing.T) {
	pool := databaseBefore(t, "00009", domain.Participants{DataRecipients: []domain.DataRecipient{{
		LegalEntityID: "dr-1", LegalEntityName: "Mock Budgeting Pty Ltd", AccreditationNumber: "ADR-1",
		AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
		Status: domain.DataRecipientSuspended,
	}}})
	// A row of the earlier schema, so that the migration changes a table that is not empty.
	_, err := pool.Exec(t.Context(),
		`INSERT INTO data_recipients (legal_entity_id, data_recipient, last_updated) VALUES ('dr-1', '{}', now())`)
	require.NoError(t, err)

	require.NoError(t, migrate(t.Context(), pool))

	var name, accreditationNumber, status string
	require.NoError(t, pool.QueryRow(t.Context(),
		"SELECT legal_entity_name, accreditation_number, status FROM data_recipients WHERE legal_entity_id = 'dr-1'").
		Scan(&name, &accreditationNumber, &status))
	assert.Equal(
		t,
		[]string{"Mock Budgeting Pty Ltd", "ADR-1", "SUSPENDED"},
		[]string{name, accreditationNumber, status},
	)

	suspended := domain.DataRecipientSuspended
	page, total, err := readmodel.NewQueries(pool).DataRecipientPage(t.Context(), app.DataRecipientPageQuery{
		Status: &suspended, NameContains: "budget", Sort: app.SortByAccreditationNumber, Page: 1, PageSize: 25,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, page, 1)
	assert.Equal(t, "Mock Budgeting Pty Ltd", page[0].DataRecipient.LegalEntityName)
	assert.Equal(t, int64(1), page[0].Version)
}

func TestMigrateMakesTheDataHolderListOfTheStreamsThatExist(t *testing.T) {
	pool := databaseBefore(t, "00010", domain.Participants{DataHolders: []domain.DataHolder{{
		LegalEntityID: "dh-1", LegalEntityName: "Mock Holdings Ltd", LogoURI: "https://holder.example/logo.png",
		Status: domain.DataHolderActive,
	}}})

	require.NoError(t, migrate(t.Context(), pool))

	page, total, err := readmodel.NewQueries(pool).DataHolderPage(t.Context(), app.DataHolderPageQuery{
		NameContains: "holdings", Sort: app.SortByName, Page: 1, PageSize: 25,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, page, 1)
	assert.Equal(t, "dh-1", page[0].DataHolder.LegalEntityID)
}

func TestMigrateFillsTheCountColumnsOfRowsThatExist(t *testing.T) {
	brand := func(id string, products ...domain.SoftwareProduct) domain.DataRecipientBrand {
		return domain.DataRecipientBrand{
			DataRecipientBrandID: id, BrandName: "Brand " + id, LogoURI: "https://recipient.example/brand.png",
			Status: domain.Active, SoftwareProducts: products,
		}
	}
	product := func(id string) domain.SoftwareProduct {
		return domain.SoftwareProduct{
			SoftwareProductID: id, SoftwareProductName: "Product " + id, SoftwareProductDescription: "A product",
			LogoURI: "https://recipient.example/product.png", Status: domain.Active,
			ClientURI: "https://recipient.example", RedirectURIs: []string{"https://recipient.example/callback"},
			JWKSURI: "https://recipient.example/jwks", RevocationURI: "https://recipient.example/revoke",
			RecipientBaseURI: "https://cdr.recipient.example", Scopes: []string{"openid", "profile"},
		}
	}
	pool := databaseBefore(t, "00014", domain.Participants{
		DataRecipients: []domain.DataRecipient{{
			LegalEntityID: "dr-1", LegalEntityName: "Mock Budgeting Pty Ltd", AccreditationNumber: "ADR-1",
			AccreditationLevel: domain.Unrestricted, LogoURI: "https://recipient.example/logo.png",
			Status: domain.DataRecipientActive,
			Brands: []domain.DataRecipientBrand{brand("b-1", product("sp-1"), product("sp-2")), brand("b-2")},
		}},
		DataHolders: []domain.DataHolder{{
			LegalEntityID: "dh-1", LegalEntityName: "Mock Holdings Ltd", LogoURI: "https://holder.example/logo.png",
			Status: domain.DataHolderActive,
			Brands: []domain.DataHolderBrand{{
				DataHolderBrandID: "dh-1-b1", BrandName: "Mock Bank", Industries: []domain.Industry{domain.Banking},
				LogoURI: "https://holder.example/brand.png", Status: domain.Active,
				EndpointDetail: domain.EndpointDetail{
					Version: "1", PublicBaseURI: "https://api.holder.example",
					ResourceBaseURI: "https://mtls.holder.example", InfosecBaseURI: "https://id.holder.example",
					WebsiteURI: "https://holder.example",
				},
				JWKSEndpoint: "https://holder.example/jwks",
			}},
		}},
	})
	// Rows of the earlier schema, so that the migration changes tables that are not empty.
	_, err := pool.Exec(t.Context(),
		`INSERT INTO data_recipients
			(legal_entity_id, data_recipient, last_updated, legal_entity_name, accreditation_number, status, version)
			VALUES ('dr-1', '{}', now(), 'Mock Budgeting Pty Ltd', 'ADR-1', 'ACTIVE', 1)`)
	require.NoError(t, err)
	_, err = pool.Exec(t.Context(),
		`INSERT INTO data_holders (legal_entity_id, legal_entity_name, status, brand_names, data_holder, last_updated)
			VALUES ('dh-1', 'Mock Holdings Ltd', 'ACTIVE', '{}', '{}', now())`)
	require.NoError(t, err)

	require.NoError(t, migrate(t.Context(), pool))

	var brands, products int
	require.NoError(t, pool.QueryRow(t.Context(),
		"SELECT brand_count, software_product_count FROM data_recipients WHERE legal_entity_id = 'dr-1'").
		Scan(&brands, &products))
	assert.Equal(t, []int{2, 2}, []int{brands, products})

	var holderBrands int
	require.NoError(t, pool.QueryRow(t.Context(),
		"SELECT brand_count FROM data_holders WHERE legal_entity_id = 'dh-1'").Scan(&holderBrands))
	assert.Equal(t, 1, holderBrands)
}

func TestMigrateFillsTheDataHolderBrandIDOfRowsThatExist(t *testing.T) {
	brand := func(registeredID, interimID string) domain.DataHolderBrand {
		return domain.DataHolderBrand{
			DataHolderBrandID: registeredID, InterimID: interimID, BrandName: "Mock Bank",
			Industries: []domain.Industry{domain.Banking}, LogoURI: "https://holder.example/brand.png",
			Status: domain.Active,
			EndpointDetail: domain.EndpointDetail{
				Version: "1", PublicBaseURI: "https://api.holder.example",
				ResourceBaseURI: "https://mtls.holder.example", InfosecBaseURI: "https://id.holder.example",
				WebsiteURI: "https://holder.example",
			},
			JWKSEndpoint: "https://holder.example/jwks",
		}
	}
	pool := databaseBefore(t, "00015", domain.Participants{DataHolders: []domain.DataHolder{{
		LegalEntityID: "dh-1", LegalEntityName: "Mock Holdings Ltd", LogoURI: "https://holder.example/logo.png",
		Status: domain.DataHolderActive, Brands: []domain.DataHolderBrand{brand("dh-1-b1", ""), brand("", "dh-1-i1")},
	}}})
	// Rows of the earlier schema, so that the migration changes a table that is not empty.
	_, err := pool.Exec(t.Context(),
		`INSERT INTO data_holder_brands
			(brand_id, legal_entity_id, status, industries, brand, last_updated, legal_entity)
			VALUES ('dh-1-b1', 'dh-1', 'ACTIVE', '{BANKING}', '{"dataHolderBrandId": "dh-1-b1"}', now(), '{}'),
				('dh-1-i1', 'dh-1', 'ACTIVE', '{BANKING}', '{"interimId": "dh-1-i1"}', now(), '{}')`)
	require.NoError(t, err)

	require.NoError(t, migrate(t.Context(), pool))

	rows, err := pool.Query(t.Context(),
		"SELECT brand_id, data_holder_brand_id FROM data_holder_brands ORDER BY brand_id")
	require.NoError(t, err)
	got, err := pgx.CollectRows(rows, pgx.RowToMap)
	require.NoError(t, err)
	assert.Equal(t, []map[string]any{
		{"brand_id": "dh-1-b1", "data_holder_brand_id": "dh-1-b1"},
		{"brand_id": "dh-1-i1", "data_holder_brand_id": nil},
	}, got)

	page, total, err := readmodel.NewQueries(pool).RegisteredDataHolderBrands(
		t.Context(), app.DataHolderBrandQuery{Page: 1, PageSize: 25},
	)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, page, 1)
	assert.Equal(t, "dh-1-b1", page[0].Brand.DataHolderBrandID)
}
