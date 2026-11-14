package domain_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func importDataRecipient(dr domain.DataRecipient) ([]domain.DataRecipientEvent, error) {
	return domain.HandleDataRecipient(domain.DataRecipientState{}, domain.ImportDataRecipient{DataRecipient: dr})
}

func TestImportDataRecipientGivesAnImportedEvent(t *testing.T) {
	dr := validDataRecipient()

	events, err := importDataRecipient(dr)

	require.NoError(t, err)
	assert.Equal(t, []domain.DataRecipientEvent{domain.DataRecipientImported{DataRecipient: dr}}, events)
}

func TestApplyDataRecipientImportedGivesTheImportedState(t *testing.T) {
	dr := validDataRecipient()

	state := domain.ApplyDataRecipient(domain.DataRecipientState{}, domain.DataRecipientImported{DataRecipient: dr})

	assert.Equal(t, domain.DataRecipientState{Imported: true, DataRecipient: dr}, state)
}

func TestImportDataRecipientFailsWhenTheDataRecipientExists(t *testing.T) {
	dr := validDataRecipient()
	state := domain.ApplyDataRecipient(domain.DataRecipientState{}, domain.DataRecipientImported{DataRecipient: dr})

	_, err := domain.HandleDataRecipient(state, domain.ImportDataRecipient{DataRecipient: dr})

	assert.EqualError(t, err, `data recipient "dr-1" already exists`)
}

func TestImportDataRecipientAppliesTheStatusRules(t *testing.T) {
	tests := map[string]struct {
		legalEntity domain.DataRecipientStatus
		brand       domain.Status
		product     domain.Status
		wantErr     string
	}{
		"suspended with an active product": {
			legalEntity: domain.DataRecipientSuspended,
			brand:       domain.Active,
			product:     domain.Active,
			wantErr: `software product "sp-1" is ACTIVE, but data recipient "dr-1" is SUSPENDED: ` +
				`a software product of a SUSPENDED data recipient must be INACTIVE or REMOVED`,
		},
		"suspended with an inactive product": {
			legalEntity: domain.DataRecipientSuspended, brand: domain.Active, product: domain.Inactive,
		},
		"revoked with an inactive product": {
			legalEntity: domain.DataRecipientRevoked,
			brand:       domain.Active,
			product:     domain.Inactive,
			wantErr: `software product "sp-1" is INACTIVE, but data recipient "dr-1" is REVOKED: ` +
				`a software product of a REVOKED data recipient must be REMOVED`,
		},
		"surrendered with an active product": {
			legalEntity: domain.DataRecipientSurrendered,
			brand:       domain.Active,
			product:     domain.Active,
			wantErr: `software product "sp-1" is ACTIVE, but data recipient "dr-1" is SURRENDERED: ` +
				`a software product of a SURRENDERED data recipient must be REMOVED`,
		},
		"revoked with a removed product": {
			legalEntity: domain.DataRecipientRevoked, brand: domain.Active, product: domain.Removed,
		},
		"inactive brand with an active product": {
			legalEntity: domain.DataRecipientActive,
			brand:       domain.Inactive,
			product:     domain.Active,
			wantErr: `software product "sp-1" is ACTIVE, but data recipient brand "drb-1" is INACTIVE: ` +
				`a software product of an INACTIVE data recipient brand must be INACTIVE or REMOVED`,
		},
		"removed brand with an inactive product": {
			legalEntity: domain.DataRecipientActive,
			brand:       domain.Removed,
			product:     domain.Inactive,
			wantErr: `software product "sp-1" is INACTIVE, but data recipient brand "drb-1" is REMOVED: ` +
				`a software product of a REMOVED data recipient brand must be REMOVED`,
		},
		"active with a removed product": {
			legalEntity: domain.DataRecipientActive, brand: domain.Active, product: domain.Removed,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dr := validDataRecipient()
			dr.Status = tt.legalEntity
			dr.Brands[0].Status = tt.brand
			dr.Brands[0].SoftwareProducts[0].Status = tt.product

			_, err := importDataRecipient(dr)

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestImportDataRecipientRejectsInvalidFields(t *testing.T) {
	tests := map[string]struct {
		change  func(*domain.DataRecipient)
		wantErr string
	}{
		"legal entity ID missing": {
			func(dr *domain.DataRecipient) { dr.LegalEntityID = "" },
			`data recipient "": legalEntityId is required`,
		},
		"legal entity ID too long": {
			func(dr *domain.DataRecipient) { dr.LegalEntityID = strings.Repeat("a", domain.MaxIDLength+1) },
			fmt.Sprintf(
				`data recipient %q: legalEntityId is longer than %d characters`,
				strings.Repeat("a", domain.MaxIDLength+1),
				domain.MaxIDLength,
			),
		},
		"unknown status": {
			func(dr *domain.DataRecipient) { dr.Status = "PAUSED" },
			`data recipient "dr-1": status "PAUSED" is not ACTIVE, SUSPENDED, REVOKED or SURRENDERED`,
		},
		"unknown accreditation level": {
			func(dr *domain.DataRecipient) { dr.AccreditationLevel = "FULL" },
			`data recipient "dr-1": accreditationLevel "FULL" is not UNRESTRICTED or SPONSORED`,
		},
		"accreditation number missing": {
			func(dr *domain.DataRecipient) { dr.AccreditationNumber = "" },
			`data recipient "dr-1": accreditationNumber is required`,
		},
		"brand name missing": {
			func(dr *domain.DataRecipient) { dr.Brands[0].BrandName = "" },
			`data recipient brand "drb-1": brandName is required`,
		},
		"unknown brand status": {
			func(dr *domain.DataRecipient) { dr.Brands[0].Status = "GONE" },
			`data recipient brand "drb-1": status "GONE" is not ACTIVE, INACTIVE or REMOVED`,
		},
		"product with no redirect URIs": {
			func(dr *domain.DataRecipient) { dr.Brands[0].SoftwareProducts[0].RedirectURIs = nil },
			`software product "sp-1": redirectUris is required`,
		},
		"product with no scopes": {
			func(dr *domain.DataRecipient) { dr.Brands[0].SoftwareProducts[0].Scopes = nil },
			`software product "sp-1": scopes is required`,
		},
		"product JWKS URI missing": {
			func(dr *domain.DataRecipient) { dr.Brands[0].SoftwareProducts[0].JWKSURI = "" },
			`software product "sp-1": jwksUri is required`,
		},
		"legal entity name too long": {
			func(dr *domain.DataRecipient) { dr.LegalEntityName = strings.Repeat("é", domain.MaxNameLength+1) },
			fmt.Sprintf(`data recipient "dr-1": legalEntityName is longer than %d characters`, domain.MaxNameLength),
		},
		"accreditation number too long": {
			func(dr *domain.DataRecipient) {
				dr.AccreditationNumber = strings.Repeat("a", domain.MaxAccreditationNumberLength+1)
			},
			fmt.Sprintf(
				`data recipient "dr-1": accreditationNumber is longer than %d characters`,
				domain.MaxAccreditationNumberLength,
			),
		},
		"legal entity logo URI too long": {
			func(dr *domain.DataRecipient) { dr.LogoURI = strings.Repeat("a", domain.MaxURILength+1) },
			fmt.Sprintf(`data recipient "dr-1": logoUri is longer than %d characters`, domain.MaxURILength),
		},
		"brand name too long": {
			func(dr *domain.DataRecipient) { dr.Brands[0].BrandName = strings.Repeat("a", domain.MaxNameLength+1) },
			fmt.Sprintf(`data recipient brand "drb-1": brandName is longer than %d characters`, domain.MaxNameLength),
		},
		"brand logo URI too long": {
			func(dr *domain.DataRecipient) { dr.Brands[0].LogoURI = strings.Repeat("a", domain.MaxURILength+1) },
			fmt.Sprintf(`data recipient brand "drb-1": logoUri is longer than %d characters`, domain.MaxURILength),
		},
		"product name too long": {
			func(dr *domain.DataRecipient) {
				dr.Brands[0].SoftwareProducts[0].SoftwareProductName = strings.Repeat("a", domain.MaxNameLength+1)
			},
			fmt.Sprintf(
				`software product "sp-1": softwareProductName is longer than %d characters`,
				domain.MaxNameLength,
			),
		},
		"product description too long": {
			func(dr *domain.DataRecipient) {
				dr.Brands[0].SoftwareProducts[0].SoftwareProductDescription = strings.Repeat(
					"a",
					domain.MaxDescriptionLength+1,
				)
			},
			fmt.Sprintf(
				`software product "sp-1": softwareProductDescription is longer than %d characters`,
				domain.MaxDescriptionLength,
			),
		},
		"product logo URI too long": {
			func(dr *domain.DataRecipient) {
				dr.Brands[0].SoftwareProducts[0].LogoURI = strings.Repeat("a", domain.MaxURILength+1)
			},
			fmt.Sprintf(`software product "sp-1": logoUri is longer than %d characters`, domain.MaxURILength),
		},
		"duplicate ID in the aggregate": {
			func(dr *domain.DataRecipient) { dr.Brands[0].SoftwareProducts[0].SoftwareProductID = "drb-1" },
			`software product "drb-1": the ID is not unique`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dr := validDataRecipient()
			tt.change(&dr)

			_, err := importDataRecipient(dr)

			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestImportDataRecipientAcceptsTextAtTheMaximumLength(t *testing.T) {
	dr := validDataRecipient()
	dr.LegalEntityName = strings.Repeat("é", domain.MaxNameLength)
	dr.Brands[0].SoftwareProducts[0].SoftwareProductDescription = strings.Repeat("a", domain.MaxDescriptionLength)

	_, err := importDataRecipient(dr)

	require.NoError(t, err)
}

func TestImportDataRecipientGivesAllErrors(t *testing.T) {
	dr := validDataRecipient()
	dr.LegalEntityName = ""
	dr.Brands[0].SoftwareProducts[0].SoftwareProductName = ""

	_, err := importDataRecipient(dr)

	assert.EqualError(t, err, `data recipient "dr-1": legalEntityName is required
software product "sp-1": softwareProductName is required`)
}

func TestSoftwareProductCountCountsTheProductsOfEachBrand(t *testing.T) {
	dr := domain.DataRecipient{Brands: []domain.DataRecipientBrand{
		{SoftwareProducts: []domain.SoftwareProduct{{SoftwareProductID: "sp-1"}, {SoftwareProductID: "sp-2"}}},
		{},
		{SoftwareProducts: []domain.SoftwareProduct{{SoftwareProductID: "sp-3"}}},
	}}

	assert.Equal(t, 3, dr.SoftwareProductCount())
	assert.Zero(t, domain.DataRecipient{}.SoftwareProductCount())
}
