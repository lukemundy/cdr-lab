package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func dataRecipientState(dr domain.DataRecipient) domain.DataRecipientState {
	return domain.ApplyDataRecipient(domain.DataRecipientState{}, domain.DataRecipientImported{DataRecipient: dr})
}

func handleAndApplyDataRecipient(
	t *testing.T,
	state domain.DataRecipientState,
	command domain.DataRecipientCommand,
) (domain.DataRecipientState, []domain.DataRecipientEvent) {
	t.Helper()

	events, err := domain.HandleDataRecipient(state, command)
	require.NoError(t, err)

	for _, e := range events {
		state = domain.ApplyDataRecipient(state, e)
	}

	return state, events
}

func secondSoftwareProduct() domain.SoftwareProduct {
	sp := validDataRecipient().Brands[0].SoftwareProducts[0]
	sp.SoftwareProductID, sp.SoftwareProductName = "sp-2", "Recipient App Mobile"

	return sp
}

func secondDataRecipientBrand() domain.DataRecipientBrand {
	return domain.DataRecipientBrand{
		DataRecipientBrandID: "drb-2",
		BrandName:            "Recipient Money",
		LogoURI:              "https://recipient.example/money.png",
		Status:               domain.Active,
		SoftwareProducts:     []domain.SoftwareProduct{secondSoftwareProduct()},
	}
}

func TestCreateDataRecipientGivesACreatedEvent(t *testing.T) {
	dr := validDataRecipient()

	state, events := handleAndApplyDataRecipient(
		t,
		domain.DataRecipientState{},
		domain.CreateDataRecipient{DataRecipient: dr},
	)

	assert.Equal(t, []domain.DataRecipientEvent{domain.DataRecipientCreated{DataRecipient: dr}}, events)
	assert.Equal(t, domain.DataRecipientState{Imported: true, DataRecipient: dr}, state)
}

func TestCreateDataRecipientFailsWhenTheDataRecipientExists(t *testing.T) {
	_, err := domain.HandleDataRecipient(
		dataRecipientState(validDataRecipient()),
		domain.CreateDataRecipient{DataRecipient: validDataRecipient()},
	)

	require.ErrorIs(t, err, domain.ErrExists)
	assert.EqualError(t, err, `data recipient "dr-1" already exists`)
}

func TestCreateDataRecipientClassifiesTheProblems(t *testing.T) {
	notValid := validDataRecipient()
	notValid.AccreditationNumber = ""
	brokenRule := validDataRecipient()
	brokenRule.Status = domain.DataRecipientSuspended

	_, notValidErr := domain.HandleDataRecipient(
		domain.DataRecipientState{},
		domain.CreateDataRecipient{DataRecipient: notValid},
	)
	_, ruleErr := domain.HandleDataRecipient(
		domain.DataRecipientState{},
		domain.CreateDataRecipient{DataRecipient: brokenRule},
	)

	require.ErrorIs(t, notValidErr, domain.ErrNotValid)
	require.EqualError(t, notValidErr, `data recipient "dr-1": accreditationNumber is required`)

	var statusRule domain.StatusRuleError
	require.ErrorAs(t, ruleErr, &statusRule)
	assert.NotErrorIs(t, ruleErr, domain.ErrNotValid)
}

func TestEachDataRecipientChangeOfADataRecipientThatDoesNotExistIsNotFound(t *testing.T) {
	commands := map[string]domain.DataRecipientCommand{
		"change details":                 domain.ChangeDataRecipientDetails{},
		"add a brand":                    domain.AddDataRecipientBrand{Brand: secondDataRecipientBrand()},
		"change brand details":           domain.ChangeDataRecipientBrandDetails{BrandID: "drb-1"},
		"activate a brand":               domain.ActivateDataRecipientBrand{BrandID: "drb-1"},
		"add a software product":         domain.AddSoftwareProduct{BrandID: "drb-1"},
		"change software product detail": domain.ChangeSoftwareProductDetails{SoftwareProductID: "sp-1"},
	}
	for name, command := range commands {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataRecipient(domain.DataRecipientState{}, command)

			require.ErrorIs(t, err, domain.ErrNotFound)
			assert.EqualError(t, err, "data recipient does not exist")
		})
	}
}

func TestChangeDataRecipientDetailsReplacesTheDetails(t *testing.T) {
	details := domain.DataRecipientDetails{
		LegalEntityName:     "Recipient Group Pty Ltd",
		AccreditationNumber: "ADR000002",
		AccreditationLevel:  domain.Sponsored,
		LogoURI:             "https://recipient.example/new.png",
	}

	state, events := handleAndApplyDataRecipient(
		t,
		dataRecipientState(validDataRecipient()),
		domain.ChangeDataRecipientDetails{Details: details},
	)

	assert.Equal(
		t,
		[]domain.DataRecipientEvent{domain.DataRecipientDetailsChanged{DataRecipientDetails: details}},
		events,
	)

	want := validDataRecipient()
	want.LegalEntityName, want.AccreditationNumber = details.LegalEntityName, details.AccreditationNumber
	want.AccreditationLevel, want.LogoURI = details.AccreditationLevel, details.LogoURI
	assert.Equal(t, want, state.DataRecipient)
}

func TestChangeDataRecipientDetailsRefusesDetailsThatAreNotValid(t *testing.T) {
	_, err := domain.HandleDataRecipient(
		dataRecipientState(validDataRecipient()),
		domain.ChangeDataRecipientDetails{Details: domain.DataRecipientDetails{LegalEntityName: "Recipient"}},
	)

	require.ErrorIs(t, err, domain.ErrNotValid)
	require.ErrorContains(t, err, `data recipient "dr-1": accreditationNumber is required`)
	assert.ErrorContains(t, err, `accreditationLevel "" is not UNRESTRICTED or SPONSORED`)
}

func TestAddDataRecipientBrandAddsTheBrandWithItsSoftwareProducts(t *testing.T) {
	brand := secondDataRecipientBrand()

	state, events := handleAndApplyDataRecipient(
		t,
		dataRecipientState(validDataRecipient()),
		domain.AddDataRecipientBrand{Brand: brand},
	)

	assert.Equal(t, []domain.DataRecipientEvent{domain.DataRecipientBrandAdded{Brand: brand}}, events)
	assert.Equal(t, append(validDataRecipient().Brands, brand), state.DataRecipient.Brands)
}

func TestAnAddRefusesAnIDThatTheDataRecipientHas(t *testing.T) {
	brandWithProductID := secondDataRecipientBrand()
	brandWithProductID.DataRecipientBrandID = "sp-1"
	productWithBrandID := secondSoftwareProduct()
	productWithBrandID.SoftwareProductID = "drb-1"
	brandWithUsedProduct := secondDataRecipientBrand()
	brandWithUsedProduct.SoftwareProducts[0].SoftwareProductID = "dr-1"

	tests := map[string]struct {
		command domain.DataRecipientCommand
		wantID  string
	}{
		"a brand with a software product ID": {domain.AddDataRecipientBrand{Brand: brandWithProductID}, "sp-1"},
		"a software product of a new brand with the legal entity ID": {
			domain.AddDataRecipientBrand{Brand: brandWithUsedProduct}, "dr-1",
		},
		"a software product with a brand ID": {
			domain.AddSoftwareProduct{BrandID: "drb-1", SoftwareProduct: productWithBrandID}, "drb-1",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataRecipient(dataRecipientState(validDataRecipient()), tt.command)

			var notUnique domain.IDNotUniqueError
			require.ErrorAs(t, err, &notUnique)
			assert.Equal(t, tt.wantID, notUnique.ID)
		})
	}
}

func TestChangeDataRecipientBrandDetailsReplacesTheDetailsOfTheBrand(t *testing.T) {
	details := domain.DataRecipientBrandDetails{
		BrandName: "Recipient Pay",
		LogoURI:   "https://recipient.example/pay.png",
	}

	state, events := handleAndApplyDataRecipient(
		t,
		dataRecipientState(validDataRecipient()),
		domain.ChangeDataRecipientBrandDetails{BrandID: "drb-1", Details: details},
	)

	assert.Equal(t, []domain.DataRecipientEvent{
		domain.DataRecipientBrandDetailsChanged{BrandID: "drb-1", DataRecipientBrandDetails: details},
	}, events)

	want := validDataRecipient().Brands[0]
	want.BrandName, want.LogoURI = details.BrandName, details.LogoURI
	assert.Equal(t, []domain.DataRecipientBrand{want}, state.DataRecipient.Brands)
}

func TestEachDataRecipientBrandCommandOfABrandThatDoesNotExistIsNotFound(t *testing.T) {
	commands := map[string]domain.DataRecipientCommand{
		"change details":         domain.ChangeDataRecipientBrandDetails{BrandID: "drb-9"},
		"activate":               domain.ActivateDataRecipientBrand{BrandID: "drb-9"},
		"deactivate":             domain.DeactivateDataRecipientBrand{BrandID: "drb-9"},
		"remove":                 domain.RemoveDataRecipientBrand{BrandID: "drb-9"},
		"add a software product": domain.AddSoftwareProduct{BrandID: "drb-9", SoftwareProduct: secondSoftwareProduct()},
	}
	for name, command := range commands {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataRecipient(dataRecipientState(validDataRecipient()), command)

			require.ErrorIs(t, err, domain.ErrNotFound)
			assert.EqualError(t, err, `data recipient brand "drb-9" does not exist`)
		})
	}
}

func TestADataRecipientBrandStatusChangeCascadesToItsSoftwareProducts(t *testing.T) {
	dr := validDataRecipient()
	inactive, removed := secondSoftwareProduct(), secondSoftwareProduct()
	inactive.Status = domain.Inactive
	removed.SoftwareProductID, removed.Status = "sp-3", domain.Removed
	dr.Brands[0].SoftwareProducts = append(dr.Brands[0].SoftwareProducts, inactive, removed)
	dr.Brands = append(dr.Brands, secondDataRecipientBrand())
	dr.Brands[1].SoftwareProducts[0].SoftwareProductID = "sp-4"

	tests := map[string]struct {
		command domain.DataRecipientCommand
		want    []domain.DataRecipientEvent
	}{
		"deactivate": {
			domain.DeactivateDataRecipientBrand{BrandID: "drb-1"},
			[]domain.DataRecipientEvent{
				domain.DataRecipientBrandDeactivated{BrandID: "drb-1"},
				domain.SoftwareProductDeactivated{SoftwareProductID: "sp-1"},
			},
		},
		"remove": {
			domain.RemoveDataRecipientBrand{BrandID: "drb-1"},
			[]domain.DataRecipientEvent{
				domain.DataRecipientBrandRemoved{BrandID: "drb-1"},
				domain.SoftwareProductRemoved{SoftwareProductID: "sp-1"},
				domain.SoftwareProductRemoved{SoftwareProductID: "sp-2"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			state, events := handleAndApplyDataRecipient(t, dataRecipientState(dr), tt.command)

			assert.Equal(t, tt.want, events)
			assert.Equal(t, domain.Active, state.DataRecipient.Brands[1].SoftwareProducts[0].Status,
				"a software product of a different brand does not change")
		})
	}
}

func TestActivateDataRecipientBrandChangesNoSoftwareProduct(t *testing.T) {
	dr := validDataRecipient()
	dr.Brands[0].Status = domain.Inactive
	dr.Brands[0].SoftwareProducts[0].Status = domain.Inactive

	state, events := handleAndApplyDataRecipient(
		t,
		dataRecipientState(dr),
		domain.ActivateDataRecipientBrand{BrandID: "drb-1"},
	)

	assert.Equal(t, []domain.DataRecipientEvent{domain.DataRecipientBrandActivated{BrandID: "drb-1"}}, events)
	assert.Equal(t, domain.Active, state.DataRecipient.Brands[0].Status)
	assert.Equal(t, domain.Inactive, state.DataRecipient.Brands[0].SoftwareProducts[0].Status)
}

func TestARemovedDataRecipientBrandIsFinal(t *testing.T) {
	dr := validDataRecipient()
	dr.Brands[0].Status = domain.Removed
	dr.Brands[0].SoftwareProducts[0].Status = domain.Removed

	_, err := domain.HandleDataRecipient(dataRecipientState(dr), domain.ActivateDataRecipientBrand{BrandID: "drb-1"})

	var ruleErr domain.StatusRuleError
	require.ErrorAs(t, err, &ruleErr)
	assert.EqualError(t, err, `data recipient brand "drb-1" cannot change from REMOVED to ACTIVE: REMOVED is final`)
}

func TestAddSoftwareProductAddsTheSoftwareProductToTheBrand(t *testing.T) {
	sp := secondSoftwareProduct()

	state, events := handleAndApplyDataRecipient(
		t,
		dataRecipientState(validDataRecipient()),
		domain.AddSoftwareProduct{BrandID: "drb-1", SoftwareProduct: sp},
	)

	assert.Equal(
		t,
		[]domain.DataRecipientEvent{domain.SoftwareProductAdded{BrandID: "drb-1", SoftwareProduct: sp}},
		events,
	)
	assert.Equal(
		t,
		append(validDataRecipient().Brands[0].SoftwareProducts, sp),
		state.DataRecipient.Brands[0].SoftwareProducts,
	)
}

func TestAddSoftwareProductAppliesTheStatusRules(t *testing.T) {
	dr := validDataRecipient()
	dr.Brands[0].Status = domain.Inactive
	dr.Brands[0].SoftwareProducts[0].Status = domain.Inactive

	_, err := domain.HandleDataRecipient(
		dataRecipientState(dr),
		domain.AddSoftwareProduct{BrandID: "drb-1", SoftwareProduct: secondSoftwareProduct()},
	)

	var ruleErr domain.StatusRuleError
	require.ErrorAs(t, err, &ruleErr)
	assert.EqualError(t, err, `software product "sp-2" is ACTIVE, but data recipient brand "drb-1" is INACTIVE: `+
		`a software product of an INACTIVE data recipient brand must be INACTIVE or REMOVED`)
}

func TestChangeSoftwareProductDetailsReplacesTheMetadataAndKeepsTheStatusAndCertificates(t *testing.T) {
	dr := validDataRecipient()
	dr.Brands[0].SoftwareProducts[0].ClientCertificates = []string{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}
	details := domain.SoftwareProductDetails{
		SoftwareProductName:        "Recipient App Web 2",
		SoftwareProductDescription: "A new web app",
		LogoURI:                    "https://recipient.example/product2.png",
		ClientURI:                  "https://recipient.example/2",
		SectorIdentifierURI:        "https://recipient.example/sector",
		RedirectURIs:               []string{"https://recipient.example/callback2"},
		TOSURI:                     "https://recipient.example/tos",
		PolicyURI:                  "https://recipient.example/policy",
		JWKSURI:                    "https://recipient.example/jwks2",
		RevocationURI:              "https://recipient.example/revoke2",
		RecipientBaseURI:           "https://cdr2.recipient.example",
		Scopes:                     []string{"openid", "profile"},
	}

	state, events := handleAndApplyDataRecipient(
		t,
		dataRecipientState(dr),
		domain.ChangeSoftwareProductDetails{SoftwareProductID: "sp-1", Details: details},
	)

	assert.Equal(t, []domain.DataRecipientEvent{
		domain.SoftwareProductDetailsChanged{SoftwareProductID: "sp-1", SoftwareProductDetails: details},
	}, events)
	assert.Equal(t, domain.SoftwareProduct{
		SoftwareProductID:          "sp-1",
		SoftwareProductName:        details.SoftwareProductName,
		SoftwareProductDescription: details.SoftwareProductDescription,
		LogoURI:                    details.LogoURI,
		Status:                     domain.Active,
		ClientURI:                  details.ClientURI,
		SectorIdentifierURI:        details.SectorIdentifierURI,
		RedirectURIs:               details.RedirectURIs,
		TOSURI:                     details.TOSURI,
		PolicyURI:                  details.PolicyURI,
		JWKSURI:                    details.JWKSURI,
		RevocationURI:              details.RevocationURI,
		RecipientBaseURI:           details.RecipientBaseURI,
		Scopes:                     details.Scopes,
		ClientCertificates:         dr.Brands[0].SoftwareProducts[0].ClientCertificates,
	}, state.DataRecipient.Brands[0].SoftwareProducts[0])
}

func TestChangeSoftwareProductDetailsRefusesDetailsThatAreNotValidOrAnUnknownSoftwareProduct(t *testing.T) {
	state := dataRecipientState(validDataRecipient())

	_, notValid := domain.HandleDataRecipient(state, domain.ChangeSoftwareProductDetails{SoftwareProductID: "sp-1"})
	_, notFound := domain.HandleDataRecipient(state, domain.ChangeSoftwareProductDetails{SoftwareProductID: "sp-9"})

	require.ErrorIs(t, notValid, domain.ErrNotValid)
	require.ErrorContains(t, notValid, `software product "sp-1": softwareProductName is required`)
	require.ErrorIs(t, notFound, domain.ErrNotFound)
	assert.EqualError(t, notFound, `software product "sp-9" does not exist`)
}
