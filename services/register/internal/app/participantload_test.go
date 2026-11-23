package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// loadFile gives the participants of seededWithProduct, so that each record of the file exists.
func loadFile(t *testing.T, store *memoryStore) domain.Participants {
	t.Helper()

	dh, _, err := app.DataHolderType.Load(t.Context(), store, "dh-1")
	require.NoError(t, err)
	dr, _, err := app.DataRecipientType.Load(t.Context(), store, "dr-1")
	require.NoError(t, err)

	return domain.Participants{
		DataHolders:    []domain.DataHolder{dh.DataHolder},
		DataRecipients: []domain.DataRecipient{dr.DataRecipient},
	}
}

func product(t *testing.T, store *memoryStore, id string) domain.SoftwareProduct {
	t.Helper()

	sp := loadFile(t, store).DataRecipients[0].Brands[0].SoftwareProducts[0]
	sp.SoftwareProductID = id

	return sp
}

func holderBrand(id string) domain.DataHolderBrand {
	return domain.DataHolderBrand{
		DataHolderBrandID: id,
		BrandName:         "Bank",
		Industries:        []domain.Industry{domain.Banking},
		LogoURI:           "https://holder.example/brand.png",
		Status:            domain.Active,
		EndpointDetail: domain.EndpointDetail{
			Version:         "1",
			PublicBaseURI:   "https://public.holder.example",
			ResourceBaseURI: "https://resource.holder.example",
			InfosecBaseURI:  "https://infosec.holder.example",
			WebsiteURI:      "https://holder.example",
		},
		JWKSEndpoint: "https://holder.example/jwks",
	}
}

// Verifies: REG-1533, REG-1540
func TestASecondLoadOfTheSameFileChangesNothing(t *testing.T) {
	store := seededWithProduct(t)
	file := loadFile(t, store)
	file.DataRecipients[0].Brands[0].SoftwareProducts[0].SoftwareProductName = "Changed"
	file.DataRecipients[0].Brands[0].Status = domain.Inactive
	file.DataRecipients[0].Brands[0].SoftwareProducts[0].Status = domain.Inactive

	load, err := app.NewOperator(store).LoadParticipants(t.Context(), file)

	require.NoError(t, err)
	assert.Empty(t, load.Added)
	assert.Equal(t, []string{"dh-1", "dr-1", "drb-1", "sp-1"}, load.Existing)
	assert.Len(t, store.appends, 2)
}

// Verifies: REG-1534, REG-1537
func TestALoadAddsEachRecordThatDoesNotExistWithTheCreateAndAddCommands(t *testing.T) {
	store := seededWithProduct(t)
	file := loadFile(t, store)
	file.DataHolders[0].Brands = []domain.DataHolderBrand{holderBrand("dhb-1")}
	brand := &file.DataRecipients[0].Brands[0]
	brand.SoftwareProducts = append(brand.SoftwareProducts, product(t, store, "sp-2"))
	newRecipient := file.DataRecipients[0]
	newRecipient.LegalEntityID, newRecipient.Brands = "dr-2", []domain.DataRecipientBrand{{
		DataRecipientBrandID: "drb-2",
		BrandName:            "Second",
		LogoURI:              "https://recipient.example/brand.png",
		Status:               domain.Active,
		SoftwareProducts:     []domain.SoftwareProduct{product(t, store, "sp-3")},
	}}
	file.DataRecipients = append(file.DataRecipients, newRecipient)

	load, err := app.NewOperator(store).LoadParticipants(t.Context(), file)

	require.NoError(t, err)
	assert.Equal(t, []string{"dhb-1", "sp-2", "dr-2", "drb-2", "sp-3"}, load.Added)
	assert.Equal(t, []string{"dh-1", "dr-1", "drb-1", "sp-1"}, load.Existing)
	require.Len(t, store.appends, 5)

	events := map[string][]string{}

	for _, a := range store.appends[2:] {
		for _, e := range a.Events {
			events[a.StreamID] = append(events[a.StreamID], e.EventType)
		}
	}

	assert.Equal(t, map[string][]string{
		"dh-1": {"DataHolderBrandAdded"},
		"dr-1": {"SoftwareProductAdded"},
		"dr-2": {"DataRecipientCreated"},
	}, events)
	assert.Equal(t, int64(1), store.appends[2].ExpectedVersion)
	assert.Equal(t, int64(0), store.appends[4].ExpectedVersion)

	dr, _, err := app.DataRecipientType.Load(t.Context(), store, "dr-1")
	require.NoError(t, err)
	assert.Equal(t, "sp-2", dr.DataRecipient.Brands[0].SoftwareProducts[1].SoftwareProductID)
}

// Verifies: REG-1533
func TestASoftwareProductInADifferentBrandOfTheLegalEntityExists(t *testing.T) {
	store := seededWithProduct(t)
	file := loadFile(t, store)
	moved := file.DataRecipients[0].Brands[0]
	moved.DataRecipientBrandID = "drb-2"
	file.DataRecipients[0].Brands[0].SoftwareProducts = nil
	file.DataRecipients[0].Brands = append(file.DataRecipients[0].Brands, moved)

	load, err := app.NewOperator(store).LoadParticipants(t.Context(), file)

	require.NoError(t, err)
	assert.Equal(t, []string{"drb-2"}, load.Added)
	assert.Equal(t, []string{"dh-1", "dr-1", "drb-1", "sp-1"}, load.Existing)

	dr, _, err := app.DataRecipientType.Load(t.Context(), store, "dr-1")
	require.NoError(t, err)
	require.Len(t, dr.DataRecipient.Brands, 2)
	assert.Empty(t, dr.DataRecipient.Brands[1].SoftwareProducts)
	assert.Len(t, dr.DataRecipient.Brands[0].SoftwareProducts, 1)
}

// Verifies: REG-1535, REG-1539
func TestALoadWithOneRecordThatIsNotValidAddsNothing(t *testing.T) {
	store := seededWithProduct(t)
	file := loadFile(t, store)
	file.DataHolders[0].Brands = []domain.DataHolderBrand{holderBrand("dhb-1")}
	bad := product(t, store, "sp-2")
	bad.JWKSURI = "http://recipient.example/jwks"
	file.DataRecipients[0].Brands[0].SoftwareProducts = append(file.DataRecipients[0].Brands[0].SoftwareProducts, bad)

	_, err := app.NewOperator(store).LoadParticipants(t.Context(), file)

	require.ErrorIs(t, err, domain.ErrNotValid)
	require.ErrorContains(t, err, `software product "sp-2": jwksUri is not an https URI with a host`)
	assert.Len(t, store.appends, 2)
}

// Verifies: REG-1535
func TestALoadGivesAStatusRuleErrorOnlyWhenEachErrorIsAStatusRule(t *testing.T) {
	store := seededWithProduct(t)
	file := loadFile(t, store)
	file.DataRecipients[0].Brands[0].SoftwareProducts = nil
	inactive := file.DataRecipients[0].Brands[0]
	inactive.DataRecipientBrandID, inactive.Status = "drb-2", domain.Inactive
	inactive.SoftwareProducts = []domain.SoftwareProduct{product(t, store, "sp-2")}
	file.DataRecipients[0].Brands = append(file.DataRecipients[0].Brands, inactive)

	_, err := app.NewOperator(store).LoadParticipants(t.Context(), file)

	require.ErrorAs(t, err, new(domain.StatusRuleError))
	require.NotErrorIs(t, err, domain.ErrNotValid)

	file.DataHolders[0].Brands = []domain.DataHolderBrand{holderBrand("")}

	_, err = app.NewOperator(store).LoadParticipants(t.Context(), file)

	require.ErrorIs(t, err, domain.ErrNotValid)
	assert.Len(t, store.appends, 2)
}

// Verifies: REG-1536
func TestALoadRefusesAnIDOfADifferentKindInTheLegalEntity(t *testing.T) {
	store := seededWithProduct(t)
	file := loadFile(t, store)
	file.DataRecipients[0].Brands[0].SoftwareProducts[0].SoftwareProductID = "drb-2"
	file.DataRecipients[0].Brands = append(file.DataRecipients[0].Brands, domain.DataRecipientBrand{
		DataRecipientBrandID: "sp-1",
		BrandName:            "Second",
		LogoURI:              "https://recipient.example/brand.png",
		Status:               domain.Active,
	})

	_, err := app.NewOperator(store).LoadParticipants(t.Context(), file)

	require.ErrorAs(t, err, new(domain.IDNotUniqueError))
	assert.Len(t, store.appends, 2)
}

func TestALoadRefusesADuplicateIDInTheFileAsNotValid(t *testing.T) {
	store := seededWithProduct(t)
	file := loadFile(t, store)
	file.DataHolders[0].LegalEntityID = "dr-1"

	_, err := app.NewOperator(store).LoadParticipants(t.Context(), file)

	require.ErrorIs(t, err, domain.ErrNotValid)
	assert.Len(t, store.appends, 2)
}
