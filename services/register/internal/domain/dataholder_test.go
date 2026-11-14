package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func importDataHolder(dh domain.DataHolder) ([]domain.DataHolderEvent, error) {
	return domain.HandleDataHolder(domain.DataHolderState{}, domain.ImportDataHolder{DataHolder: dh})
}

func TestImportDataHolderGivesAnImportedEvent(t *testing.T) {
	dh := validDataHolder()

	events, err := importDataHolder(dh)

	require.NoError(t, err)
	assert.Equal(t, []domain.DataHolderEvent{domain.DataHolderImported{DataHolder: dh}}, events)
}

func TestApplyDataHolderImportedGivesTheImportedState(t *testing.T) {
	dh := validDataHolder()

	state := domain.ApplyDataHolder(domain.DataHolderState{}, domain.DataHolderImported{DataHolder: dh})

	assert.Equal(t, domain.DataHolderState{Imported: true, DataHolder: dh}, state)
}

func TestImportDataHolderFailsWhenTheDataHolderExists(t *testing.T) {
	dh := validDataHolder()
	state := domain.ApplyDataHolder(domain.DataHolderState{}, domain.DataHolderImported{DataHolder: dh})

	_, err := domain.HandleDataHolder(state, domain.ImportDataHolder{DataHolder: dh})

	assert.EqualError(t, err, `data holder "dh-1" already exists`)
}

func TestImportDataHolderAppliesTheStatusRules(t *testing.T) {
	tests := map[string]struct {
		legalEntity domain.DataHolderStatus
		brand       domain.Status
		wantErr     string
	}{
		"removed with an active brand": {
			legalEntity: domain.DataHolderRemoved,
			brand:       domain.Active,
			wantErr: `data holder brand "dhb-1" is ACTIVE, but data holder "dh-1" is REMOVED: ` +
				`a brand of a REMOVED data holder must be REMOVED`,
		},
		"removed with an inactive brand": {
			legalEntity: domain.DataHolderRemoved,
			brand:       domain.Inactive,
			wantErr: `data holder brand "dhb-1" is INACTIVE, but data holder "dh-1" is REMOVED: ` +
				`a brand of a REMOVED data holder must be REMOVED`,
		},
		"removed with a removed brand":  {legalEntity: domain.DataHolderRemoved, brand: domain.Removed},
		"active with an inactive brand": {legalEntity: domain.DataHolderActive, brand: domain.Inactive},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dh := validDataHolder()
			dh.Status = tt.legalEntity
			dh.Brands[0].Status = tt.brand

			_, err := importDataHolder(dh)

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestImportDataHolderRejectsInvalidFields(t *testing.T) {
	tests := map[string]struct {
		change  func(*domain.DataHolder)
		wantErr string
	}{
		"unknown status": {
			func(dh *domain.DataHolder) { dh.Status = "INACTIVE" },
			`data holder "dh-1": status "INACTIVE" is not ACTIVE or REMOVED`,
		},
		"logo URI missing": {
			func(dh *domain.DataHolder) { dh.LogoURI = "" },
			`data holder "dh-1": logoUri is required`,
		},
		"no industries": {
			func(dh *domain.DataHolder) { dh.Brands[0].Industries = nil },
			`data holder brand "dhb-1": industries is required`,
		},
		"unknown industry": {
			func(dh *domain.DataHolder) { dh.Brands[0].Industries = []domain.Industry{"all"} },
			`data holder brand "dhb-1": industry "all" is not banking, energy, non-bank-lending or telco`,
		},
		"duplicate industry": {
			func(dh *domain.DataHolder) { dh.Brands[0].Industries = []domain.Industry{domain.Energy, domain.Energy} },
			`data holder brand "dhb-1": industry "energy" is not unique`,
		},
		"endpoint version missing": {
			func(dh *domain.DataHolder) { dh.Brands[0].EndpointDetail.Version = "" },
			`data holder brand "dhb-1": endpointDetail.version is required`,
		},
		"JWKS endpoint missing": {
			func(dh *domain.DataHolder) { dh.Brands[0].JWKSEndpoint = "" },
			`data holder brand "dhb-1": jwksEndpoint is required`,
		},
		"duplicate ID in the aggregate": {
			func(dh *domain.DataHolder) { dh.Brands[0].DataHolderBrandID = "dh-1" },
			`data holder brand "dh-1": the ID is not unique`,
		},
		"no brand ID and no interim ID": {
			func(dh *domain.DataHolder) { dh.Brands[0].DataHolderBrandID = "" },
			`data holder brand "": dataHolderBrandId or interimId is required`,
		},
		"legal entity name too long": {
			func(dh *domain.DataHolder) { dh.LegalEntityName = strings.Repeat("n", 37) },
			`data holder "dh-1": legalEntityName is longer than 36 characters`,
		},
		"interim ID too long": {
			func(dh *domain.DataHolder) { dh.Brands[0].InterimID = strings.Repeat("i", 37) },
			`data holder brand "dhb-1": interimId is longer than 36 characters`,
		},
		"interim ID is the brand ID": {
			func(dh *domain.DataHolder) { dh.Brands[0].InterimID = "dhb-1" },
			`data holder brand "dhb-1": the ID is not unique`,
		},
		"interim ID is another ID in the aggregate": {
			func(dh *domain.DataHolder) { dh.Brands[0].InterimID = "dh-1" },
			`data holder brand "dhb-1": the ID is not unique`,
		},
		"interim-only brand with no public base URI": {
			func(dh *domain.DataHolder) {
				dh.Brands[0] = interimOnlyBrand()
				dh.Brands[0].EndpointDetail.PublicBaseURI = ""
			},
			`data holder brand "dhi-1": endpointDetail.publicBaseUri is required`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dh := validDataHolder()
			tt.change(&dh)

			_, err := importDataHolder(dh)

			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestTheIndustriesOfADataHolderAreTheIndustriesOfItsBrandsInAllStatuses(t *testing.T) {
	dh := validDataHolder()
	dh.Brands = append(dh.Brands, dh.Brands[0], dh.Brands[0])
	dh.Brands[1].Industries = []domain.Industry{domain.Telco, domain.Energy}
	dh.Brands[2].Industries = []domain.Industry{domain.NonBankLending, domain.Energy}
	dh.Brands[2].Status = domain.Removed

	assert.Equal(
		t,
		[]domain.Industry{domain.Banking, domain.Energy, domain.NonBankLending, domain.Telco},
		dh.Industries(),
	)
}

func TestADataHolderWithNoBrandsHasNoIndustries(t *testing.T) {
	dh := validDataHolder()
	dh.Brands = nil

	assert.Empty(t, dh.Industries())
}

func TestParseIndustryAcceptsOnlyTheIndustriesOfTheCDS(t *testing.T) {
	for _, s := range []string{"banking", "energy", "non-bank-lending", "telco"} {
		industry, ok := domain.ParseIndustry(s)

		assert.True(t, ok, s)
		assert.Equal(t, domain.Industry(s), industry)
	}

	for _, s := range []string{"all", "", "Banking", "foo"} {
		_, ok := domain.ParseIndustry(s)

		assert.False(t, ok, s)
	}
}

func interimOnlyBrand() domain.DataHolderBrand {
	return domain.DataHolderBrand{
		InterimID:      "dhi-1",
		BrandName:      "Holder Energy",
		Industries:     []domain.Industry{domain.Energy},
		LogoURI:        "https://holder.example/energy.png",
		Status:         domain.Active,
		EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://prd.holder.example"},
	}
}

func TestABrandWithOnlyAnInterimIDNeedsOnlyItsPublicDetails(t *testing.T) {
	dh := validDataHolder()
	dh.Brands = append(dh.Brands, interimOnlyBrand())

	_, err := importDataHolder(dh)

	assert.NoError(t, err)
}

func TestABrandCanHaveABrandIDAndAnInterimID(t *testing.T) {
	dh := validDataHolder()
	dh.Brands[0].InterimID = "dhi-1"

	_, err := importDataHolder(dh)

	assert.NoError(t, err)
}

func TestTheIDOfABrandIsItsBrandIDWhenItHasOne(t *testing.T) {
	both := validDataHolder().Brands[0]
	both.InterimID = "dhi-1"

	assert.Equal(t, "dhb-1", both.ID())
	assert.Equal(t, "dhi-1", interimOnlyBrand().ID())
}
