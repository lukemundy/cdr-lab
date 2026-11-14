package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func dataHolderState(dh domain.DataHolder) domain.DataHolderState {
	return domain.ApplyDataHolder(domain.DataHolderState{}, domain.DataHolderImported{DataHolder: dh})
}

func secondDataHolderBrand() domain.DataHolderBrand {
	return domain.DataHolderBrand{
		InterimID:      "dhb-2",
		BrandName:      "Holder Energy",
		Industries:     []domain.Industry{domain.Energy},
		LogoURI:        "https://holder.example/energy.png",
		Status:         domain.Active,
		EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://energy.holder.example"},
	}
}

// handleAndApply gives the state after the events of the command.
func handleAndApply(
	t *testing.T,
	state domain.DataHolderState,
	command domain.DataHolderCommand,
) (domain.DataHolderState, []domain.DataHolderEvent) {
	t.Helper()

	events, err := domain.HandleDataHolder(state, command)
	require.NoError(t, err)

	for _, e := range events {
		state = domain.ApplyDataHolder(state, e)
	}

	return state, events
}

func TestCreateDataHolderGivesACreatedEvent(t *testing.T) {
	dh := validDataHolder()

	state, events := handleAndApply(t, domain.DataHolderState{}, domain.CreateDataHolder{DataHolder: dh})

	assert.Equal(t, []domain.DataHolderEvent{domain.DataHolderCreated{DataHolder: dh}}, events)
	assert.Equal(t, domain.DataHolderState{Imported: true, DataHolder: dh}, state)
}

func TestCreateDataHolderFailsWhenTheDataHolderExists(t *testing.T) {
	_, err := domain.HandleDataHolder(
		dataHolderState(validDataHolder()),
		domain.CreateDataHolder{DataHolder: validDataHolder()},
	)

	require.ErrorIs(t, err, domain.ErrExists)
	assert.EqualError(t, err, `data holder "dh-1" already exists`)
}

func TestCreateDataHolderRefusesAnInvalidDataHolderAsNotValid(t *testing.T) {
	dh := validDataHolder()
	dh.LegalEntityName = ""

	_, err := domain.HandleDataHolder(domain.DataHolderState{}, domain.CreateDataHolder{DataHolder: dh})

	require.ErrorIs(t, err, domain.ErrNotValid)
	assert.EqualError(t, err, `data holder "dh-1": legalEntityName is required`)
}

func TestCreateDataHolderRefusesABrokenStatusRuleAsAStatusRuleError(t *testing.T) {
	dh := validDataHolder()
	dh.Status = domain.DataHolderRemoved

	_, err := domain.HandleDataHolder(domain.DataHolderState{}, domain.CreateDataHolder{DataHolder: dh})

	var ruleErr domain.StatusRuleError
	require.ErrorAs(t, err, &ruleErr)
	assert.NotErrorIs(t, err, domain.ErrNotValid)
}

func TestAnInvalidFieldAndABrokenStatusRuleTogetherAreNotValid(t *testing.T) {
	dh := validDataHolder()
	dh.Status = domain.DataHolderRemoved
	dh.LogoURI = ""

	_, err := domain.HandleDataHolder(domain.DataHolderState{}, domain.CreateDataHolder{DataHolder: dh})

	require.ErrorIs(t, err, domain.ErrNotValid)
	require.ErrorContains(t, err, "logoUri is required")
	assert.ErrorContains(t, err, "a brand of a REMOVED data holder must be REMOVED")
}

func TestEachDataHolderChangeOfADataHolderThatDoesNotExistIsNotFound(t *testing.T) {
	commands := map[string]domain.DataHolderCommand{
		"change details":       domain.ChangeDataHolderDetails{},
		"remove":               domain.RemoveDataHolder{},
		"add a brand":          domain.AddDataHolderBrand{Brand: secondDataHolderBrand()},
		"change brand details": domain.ChangeDataHolderBrandDetails{BrandID: "dhb-1"},
		"activate a brand":     domain.ActivateDataHolderBrand{BrandID: "dhb-1"},
		"deactivate a brand":   domain.DeactivateDataHolderBrand{BrandID: "dhb-1"},
		"remove a brand":       domain.RemoveDataHolderBrand{BrandID: "dhb-1"},
	}
	for name, command := range commands {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataHolder(domain.DataHolderState{}, command)

			require.ErrorIs(t, err, domain.ErrNotFound)
			assert.EqualError(t, err, "data holder does not exist")
		})
	}
}

func TestChangeDataHolderDetailsReplacesTheDetails(t *testing.T) {
	details := domain.DataHolderDetails{LegalEntityName: "Holder Group Ltd", LogoURI: "https://holder.example/new.png"}

	state, events := handleAndApply(
		t,
		dataHolderState(validDataHolder()),
		domain.ChangeDataHolderDetails{Details: details},
	)

	assert.Equal(t, []domain.DataHolderEvent{domain.DataHolderDetailsChanged{DataHolderDetails: details}}, events)

	want := validDataHolder()
	want.LegalEntityName, want.LogoURI = details.LegalEntityName, details.LogoURI
	assert.Equal(t, want, state.DataHolder)
}

func TestChangeDataHolderDetailsRefusesDetailsThatAreNotValid(t *testing.T) {
	_, err := domain.HandleDataHolder(
		dataHolderState(validDataHolder()),
		domain.ChangeDataHolderDetails{Details: domain.DataHolderDetails{
			LegalEntityName: "A name that is longer than thirty-six characters",
		}},
	)

	require.ErrorIs(t, err, domain.ErrNotValid)
	require.ErrorContains(t, err, `data holder "dh-1": legalEntityName is longer than 36 characters`)
	assert.ErrorContains(t, err, `data holder "dh-1": logoUri is required`)
}

func TestRemoveDataHolderRemovesEachBrandThatIsNotRemoved(t *testing.T) {
	dh := validDataHolder()
	inactive, removed := secondDataHolderBrand(), secondDataHolderBrand()
	inactive.InterimID, inactive.Status = "dhb-2", domain.Inactive
	removed.InterimID, removed.Status = "dhb-3", domain.Removed
	dh.Brands = append(dh.Brands, inactive, removed)

	state, events := handleAndApply(t, dataHolderState(dh), domain.RemoveDataHolder{})

	assert.Equal(t, []domain.DataHolderEvent{
		domain.DataHolderLegalEntityRemoved{},
		domain.DataHolderBrandRemoved{BrandID: "dhb-1"},
		domain.DataHolderBrandRemoved{BrandID: "dhb-2"},
	}, events)
	assert.Equal(t, domain.DataHolderRemoved, state.DataHolder.Status)

	for _, b := range state.DataHolder.Brands {
		assert.Equal(t, domain.Removed, b.Status, b.ID())
	}
}

func TestRemoveDataHolderIsFinal(t *testing.T) {
	state, _ := handleAndApply(t, dataHolderState(validDataHolder()), domain.RemoveDataHolder{})

	_, err := domain.HandleDataHolder(state, domain.RemoveDataHolder{})

	var ruleErr domain.StatusRuleError
	require.ErrorAs(t, err, &ruleErr)
	assert.EqualError(t, err, `data holder "dh-1" cannot change from REMOVED to REMOVED: REMOVED is final`)
}

func TestAddDataHolderBrandAddsTheBrand(t *testing.T) {
	brand := secondDataHolderBrand()

	state, events := handleAndApply(
		t,
		dataHolderState(validDataHolder()),
		domain.AddDataHolderBrand{Brand: brand},
	)

	assert.Equal(t, []domain.DataHolderEvent{domain.DataHolderBrandAdded{Brand: brand}}, events)
	assert.Equal(t, append(validDataHolder().Brands, brand), state.DataHolder.Brands)
}

func TestAddDataHolderBrandRefusesAnIDThatTheDataHolderHas(t *testing.T) {
	for _, id := range []string{"dh-1", "dhb-1"} {
		t.Run(id, func(t *testing.T) {
			brand := secondDataHolderBrand()
			brand.InterimID = id

			_, err := domain.HandleDataHolder(
				dataHolderState(validDataHolder()),
				domain.AddDataHolderBrand{Brand: brand},
			)

			var notUnique domain.IDNotUniqueError
			require.ErrorAs(t, err, &notUnique)
			assert.EqualError(t, err, `the ID "`+id+`" is not unique: a different record has it`)
		})
	}
}

func TestAddDataHolderBrandAppliesTheRulesOfTheParticipantModel(t *testing.T) {
	brand := secondDataHolderBrand()
	brand.BrandName = ""

	_, err := domain.HandleDataHolder(dataHolderState(validDataHolder()), domain.AddDataHolderBrand{Brand: brand})

	require.ErrorIs(t, err, domain.ErrNotValid)
	assert.EqualError(t, err, `data holder brand "dhb-2": brandName is required`)
}

func TestAddDataHolderBrandAppliesTheStatusRules(t *testing.T) {
	removed := validDataHolder()
	removed.Status = domain.DataHolderRemoved
	removed.Brands[0].Status = domain.Removed

	_, err := domain.HandleDataHolder(
		dataHolderState(removed),
		domain.AddDataHolderBrand{Brand: secondDataHolderBrand()},
	)

	var ruleErr domain.StatusRuleError
	require.ErrorAs(t, err, &ruleErr)
	assert.EqualError(t, err, `data holder brand "dhb-2" is ACTIVE, but data holder "dh-1" is REMOVED: `+
		`a brand of a REMOVED data holder must be REMOVED`)
}

func TestChangeDataHolderBrandDetailsReplacesTheDetailsOfTheBrand(t *testing.T) {
	dh := validDataHolder()
	dh.Brands = append(dh.Brands, secondDataHolderBrand())
	details := domain.DataHolderBrandDetails{
		BrandName:      "Holder Power",
		BrandGroup:     "Holder Group",
		Industries:     []domain.Industry{domain.Energy, domain.Telco},
		LogoURI:        "https://holder.example/power.png",
		EndpointDetail: domain.EndpointDetail{PublicBaseURI: "https://power.holder.example"},
	}

	state, events := handleAndApply(
		t,
		dataHolderState(dh),
		domain.ChangeDataHolderBrandDetails{BrandID: "dhb-2", Details: details},
	)

	assert.Equal(t, []domain.DataHolderEvent{
		domain.DataHolderBrandDetailsChanged{BrandID: "dhb-2", DataHolderBrandDetails: details},
	}, events)

	want := secondDataHolderBrand()
	want.BrandName, want.BrandGroup, want.Industries = details.BrandName, details.BrandGroup, details.Industries
	want.LogoURI, want.EndpointDetail = details.LogoURI, details.EndpointDetail
	assert.Equal(t, []domain.DataHolderBrand{validDataHolder().Brands[0], want}, state.DataHolder.Brands)
}

func TestADataHolderBrandCommandFindsTheBrandByItsBrandIDOrItsInterimID(t *testing.T) {
	dh := validDataHolder()
	dh.Brands[0].InterimID = "interim-1"

	for _, id := range []string{"dhb-1", "interim-1"} {
		t.Run(id, func(t *testing.T) {
			events, err := domain.HandleDataHolder(dataHolderState(dh), domain.DeactivateDataHolderBrand{BrandID: id})

			require.NoError(t, err)
			assert.Equal(t, []domain.DataHolderEvent{domain.DataHolderBrandDeactivated{BrandID: "dhb-1"}}, events)
		})
	}
}

func TestChangeDataHolderBrandDetailsRefusesDetailsThatAreNotValid(t *testing.T) {
	_, err := domain.HandleDataHolder(
		dataHolderState(validDataHolder()),
		domain.ChangeDataHolderBrandDetails{BrandID: "dhb-1", Details: domain.DataHolderBrandDetails{
			BrandName:  "Holder Bank",
			Industries: []domain.Industry{domain.Banking},
			LogoURI:    "https://holder.example/brand.png",
		}},
	)

	require.ErrorIs(t, err, domain.ErrNotValid)
	require.ErrorContains(t, err, `data holder brand "dhb-1": endpointDetail.publicBaseUri is required`)
	assert.ErrorContains(t, err, `data holder brand "dhb-1": jwksEndpoint is required`)
}

func TestEachDataHolderBrandCommandOfABrandThatDoesNotExistIsNotFound(t *testing.T) {
	commands := map[string]domain.DataHolderCommand{
		"change details": domain.ChangeDataHolderBrandDetails{BrandID: "dhb-9"},
		"activate":       domain.ActivateDataHolderBrand{BrandID: "dhb-9"},
		"deactivate":     domain.DeactivateDataHolderBrand{BrandID: "dhb-9"},
		"remove":         domain.RemoveDataHolderBrand{BrandID: "dhb-9"},
	}
	for name, command := range commands {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataHolder(dataHolderState(validDataHolder()), command)

			require.ErrorIs(t, err, domain.ErrNotFound)
			assert.EqualError(t, err, `data holder brand "dhb-9" does not exist`)
		})
	}
}

func TestDataHolderBrandStatusTransitions(t *testing.T) {
	tests := map[string]struct {
		from    domain.Status
		command domain.DataHolderCommand
		want    domain.DataHolderEvent
		wantErr string
	}{
		"deactivate an active brand": {
			from:    domain.Active,
			command: domain.DeactivateDataHolderBrand{BrandID: "dhb-1"},
			want:    domain.DataHolderBrandDeactivated{BrandID: "dhb-1"},
		},
		"remove an active brand": {
			from:    domain.Active,
			command: domain.RemoveDataHolderBrand{BrandID: "dhb-1"},
			want:    domain.DataHolderBrandRemoved{BrandID: "dhb-1"},
		},
		"activate an inactive brand": {
			from:    domain.Inactive,
			command: domain.ActivateDataHolderBrand{BrandID: "dhb-1"},
			want:    domain.DataHolderBrandActivated{BrandID: "dhb-1"},
		},
		"remove an inactive brand": {
			from:    domain.Inactive,
			command: domain.RemoveDataHolderBrand{BrandID: "dhb-1"},
			want:    domain.DataHolderBrandRemoved{BrandID: "dhb-1"},
		},
		"activate an active brand": {
			from:    domain.Active,
			command: domain.ActivateDataHolderBrand{BrandID: "dhb-1"},
			wantErr: `data holder brand "dhb-1" cannot change from ACTIVE to ACTIVE: ` +
				`from ACTIVE, the allowed statuses are INACTIVE and REMOVED`,
		},
		"activate a removed brand": {
			from:    domain.Removed,
			command: domain.ActivateDataHolderBrand{BrandID: "dhb-1"},
			wantErr: `data holder brand "dhb-1" cannot change from REMOVED to ACTIVE: REMOVED is final`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dh := validDataHolder()
			dh.Brands[0].Status = tt.from

			events, err := domain.HandleDataHolder(dataHolderState(dh), tt.command)

			if tt.wantErr != "" {
				var ruleErr domain.StatusRuleError
				require.ErrorAs(t, err, &ruleErr)
				assert.EqualError(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, []domain.DataHolderEvent{tt.want}, events)
		})
	}
}

func TestApplyDataHolderBrandStatusEventsChangeOnlyThatBrand(t *testing.T) {
	dh := validDataHolder()
	dh.Brands = append(dh.Brands, secondDataHolderBrand())
	state := dataHolderState(dh)

	tests := map[string]struct {
		event domain.DataHolderEvent
		want  domain.Status
	}{
		"activated":   {domain.DataHolderBrandActivated{BrandID: "dhb-2"}, domain.Active},
		"deactivated": {domain.DataHolderBrandDeactivated{BrandID: "dhb-2"}, domain.Inactive},
		"removed":     {domain.DataHolderBrandRemoved{BrandID: "dhb-2"}, domain.Removed},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			after := domain.ApplyDataHolder(state, tt.event)

			assert.Equal(t, domain.Active, after.DataHolder.Brands[0].Status)
			assert.Equal(t, tt.want, after.DataHolder.Brands[1].Status)
			assert.Equal(t, domain.Active, state.DataHolder.Brands[1].Status, "apply does not change its input")
		})
	}
}
