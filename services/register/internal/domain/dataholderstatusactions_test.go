package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type (
	holderAction = domain.StatusAction[
		domain.DataHolderAction, domain.DataHolderStatus, domain.DataHolderBrandChange,
	]
	holderBrandAction = domain.StatusAction[domain.RecordAction, domain.Status, domain.DataHolderBrandChange]
)

func dataHolderStatusFixture() domain.DataHolderState {
	dh := validDataHolder()
	inactive, removed := secondDataHolderBrand(), secondDataHolderBrand()
	inactive.InterimID, inactive.BrandName, inactive.Status = "dhb-inactive", "Holder Inactive", domain.Inactive
	removed.InterimID, removed.BrandName, removed.Status = "dhb-removed", "Holder Removed", domain.Removed
	dh.Brands = append(dh.Brands, secondDataHolderBrand(), inactive, removed)

	return dataHolderState(dh)
}

// Verifies: REG-0231
func TestTheRemovalOfADataHolderGivesEachBrandThatItRemoves(t *testing.T) {
	actions := domain.PermittedDataHolderStatusActions(dataHolderStatusFixture())

	assert.Equal(t, []holderAction{{
		Action:       domain.RemoveLegalEntity,
		Status:       domain.DataHolderRemoved,
		Irreversible: true,
		Changes: []domain.DataHolderBrandChange{
			{DataHolderBrandID: "dhb-1", BrandName: "Holder Bank", From: domain.Active, To: domain.Removed},
			{InterimID: "dhb-2", BrandName: "Holder Energy", From: domain.Active, To: domain.Removed},
			{InterimID: "dhb-inactive", BrandName: "Holder Inactive", From: domain.Inactive, To: domain.Removed},
		},
	}}, actions.DataHolder)
}

// Verifies: REG-0225
func TestTheActionsOfEachDataHolderBrandFollowTheStatusRules(t *testing.T) {
	actions := domain.PermittedDataHolderStatusActions(dataHolderStatusFixture())

	none := []domain.DataHolderBrandChange{}
	deactivate := holderBrandAction{Action: domain.Deactivate, Status: domain.Inactive, Changes: none}
	remove := holderBrandAction{Action: domain.Remove, Status: domain.Removed, Irreversible: true, Changes: none}

	assert.Equal(t, map[string][]holderBrandAction{
		"dhb-1":        {deactivate, remove},
		"dhb-2":        {deactivate, remove},
		"dhb-inactive": {{Action: domain.Activate, Status: domain.Active, Changes: none}, remove},
		"dhb-removed":  {},
	}, actions.Brands)
}

// Verifies: REG-0224
func TestARemovedDataHolderHasNoAction(t *testing.T) {
	state, _ := handleAndApply(t, dataHolderStatusFixture(), domain.RemoveDataHolder{})

	actions := domain.PermittedDataHolderStatusActions(state)

	assert.Empty(t, actions.DataHolder)

	for id, brand := range actions.Brands {
		assert.Empty(t, brand, "brand %s", id)
	}
}

func TestEachDataHolderActionGivesTheCommandWithTheSameEffectAsItsCascade(t *testing.T) {
	state := dataHolderStatusFixture()
	actions := domain.PermittedDataHolderStatusActions(state)

	for _, a := range actions.DataHolder {
		after, _ := handleAndApply(t, state, a.Action.Command())

		assert.Equal(t, a.Status, after.DataHolder.Status, "%s", a.Action)

		for _, c := range a.Changes {
			assert.Equal(
				t,
				c.To,
				holderBrandStatuses(after.DataHolder)[c.DataHolderBrandID+c.InterimID],
				"%s",
				a.Action,
			)
		}
	}

	for id, brandActions := range actions.Brands {
		for _, a := range brandActions {
			after, _ := handleAndApply(t, state, a.Action.DataHolderBrandCommand(id))

			assert.Equal(t, a.Status, holderBrandStatuses(after.DataHolder)[id], "%s %s", id, a.Action)
		}
	}
}

func TestTheDataHolderActionsAreTheOnlyDataHolderStatusActionsThatTheRegisterKnows(t *testing.T) {
	assert.Equal(t, []domain.DataHolderAction{domain.RemoveLegalEntity}, domain.DataHolderActions)
}

func holderBrandStatuses(dh domain.DataHolder) map[string]domain.Status {
	got := map[string]domain.Status{}
	for _, b := range dh.Brands {
		got[b.ID()] = b.Status
	}

	return got
}
