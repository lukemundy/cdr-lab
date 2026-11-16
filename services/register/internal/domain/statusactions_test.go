package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type recipientAction = domain.StatusAction[
	domain.DataRecipientAction, domain.DataRecipientStatus, domain.SoftwareProductChange,
]

type recordAction = domain.StatusAction[domain.RecordAction, domain.Status, domain.SoftwareProductChange]

func change(id string, from, to domain.Status) domain.SoftwareProductChange {
	return domain.SoftwareProductChange{
		SoftwareProductID:   id,
		SoftwareProductName: "Recipient App Web",
		From:                from,
		To:                  to,
	}
}

func TestTheActionsOfAnActiveDataRecipientGiveTheCascadeOfEachAction(t *testing.T) {
	actions := domain.PermittedStatusActions(statusFixture(domain.DataRecipientActive))

	assert.Equal(t, []recipientAction{
		{
			Action:  domain.Suspend,
			Status:  domain.DataRecipientSuspended,
			Changes: []domain.SoftwareProductChange{change("sp-active", domain.Active, domain.Inactive)},
		},
		{
			Action:       domain.Revoke,
			Status:       domain.DataRecipientRevoked,
			Irreversible: true,
			Changes: []domain.SoftwareProductChange{
				change("sp-active", domain.Active, domain.Removed),
				change("sp-inactive", domain.Inactive, domain.Removed),
				change("sp-under-inactive-brand", domain.Inactive, domain.Removed),
			},
		},
		{
			Action:       domain.Surrender,
			Status:       domain.DataRecipientSurrendered,
			Irreversible: true,
			Changes: []domain.SoftwareProductChange{
				change("sp-active", domain.Active, domain.Removed),
				change("sp-inactive", domain.Inactive, domain.Removed),
				change("sp-under-inactive-brand", domain.Inactive, domain.Removed),
			},
		},
	}, actions.DataRecipient)
}

func TestTheActionsOfASuspendedDataRecipientIncludeAReinstatementThatChangesNoSoftwareProduct(t *testing.T) {
	actions := domain.PermittedStatusActions(statusFixture(domain.DataRecipientSuspended))

	require.Len(t, actions.DataRecipient, 3)
	assert.Equal(t, domain.Revoke, actions.DataRecipient[0].Action)
	assert.Equal(t, domain.Surrender, actions.DataRecipient[1].Action)
	assert.Equal(t, recipientAction{
		Action:  domain.Reinstate,
		Status:  domain.DataRecipientActive,
		Changes: []domain.SoftwareProductChange{},
	}, actions.DataRecipient[2])
}

func TestARevocationWithNoSoftwareProductToRemoveIsReversible(t *testing.T) {
	state := statusFixture(domain.DataRecipientSuspended)
	for _, id := range []string{"sp-active", "sp-inactive", "sp-under-inactive-brand"} {
		_, state = execute(t, state, domain.RemoveSoftwareProduct{SoftwareProductID: id})
	}

	actions := domain.PermittedStatusActions(state)

	assert.Contains(t, actions.DataRecipient, recipientAction{
		Action: domain.Revoke, Status: domain.DataRecipientRevoked, Changes: []domain.SoftwareProductChange{},
	})
}

// Verifies: REG-0222
func TestASurrenderedDataRecipientHasNoAction(t *testing.T) {
	actions := domain.PermittedStatusActions(statusFixture(domain.DataRecipientSurrendered))

	assert.Empty(t, actions.DataRecipient)

	for id, product := range actions.SoftwareProducts {
		assert.Empty(t, product, "software product %s", id)
	}
}

// Verifies: REG-0223
func TestTheActionsOfEachSoftwareProductFollowTheStatusRules(t *testing.T) {
	tests := map[string]struct {
		recipient domain.DataRecipientStatus
		want      map[string][]recordAction
	}{
		"active data recipient": {
			recipient: domain.DataRecipientActive,
			want: map[string][]recordAction{
				"sp-active": {
					{Action: domain.Deactivate, Status: domain.Inactive, Changes: []domain.SoftwareProductChange{}},
					{
						Action: domain.Remove, Status: domain.Removed, Irreversible: true,
						Changes: []domain.SoftwareProductChange{},
					},
				},
				"sp-inactive": {
					{Action: domain.Activate, Status: domain.Active, Changes: []domain.SoftwareProductChange{}},
					{
						Action: domain.Remove, Status: domain.Removed, Irreversible: true,
						Changes: []domain.SoftwareProductChange{},
					},
				},
				"sp-removed": {},
				// The brand is INACTIVE, so the software product cannot be ACTIVE.
				"sp-under-inactive-brand": {
					{
						Action: domain.Remove, Status: domain.Removed, Irreversible: true,
						Changes: []domain.SoftwareProductChange{},
					},
				},
			},
		},
		"suspended data recipient": {
			recipient: domain.DataRecipientSuspended,
			want: map[string][]recordAction{
				"sp-active": {
					{
						Action: domain.Remove, Status: domain.Removed, Irreversible: true,
						Changes: []domain.SoftwareProductChange{},
					},
				},
				"sp-inactive": {
					{
						Action: domain.Remove, Status: domain.Removed, Irreversible: true,
						Changes: []domain.SoftwareProductChange{},
					},
				},
				"sp-removed": {},
				"sp-under-inactive-brand": {
					{
						Action: domain.Remove, Status: domain.Removed, Irreversible: true,
						Changes: []domain.SoftwareProductChange{},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actions := domain.PermittedStatusActions(statusFixture(tt.recipient))

			assert.Equal(t, tt.want, actions.SoftwareProducts)
		})
	}
}

// Verifies: REG-0232, REG-0233, REG-0234
func TestTheActionsOfEachBrandGiveTheCascadeToItsSoftwareProducts(t *testing.T) {
	actions := domain.PermittedStatusActions(statusFixture(domain.DataRecipientActive))

	assert.Equal(t, map[string][]recordAction{
		"drb-1": {
			{
				Action:  domain.Deactivate,
				Status:  domain.Inactive,
				Changes: []domain.SoftwareProductChange{change("sp-active", domain.Active, domain.Inactive)},
			},
			{
				Action:       domain.Remove,
				Status:       domain.Removed,
				Irreversible: true,
				Changes: []domain.SoftwareProductChange{
					change("sp-active", domain.Active, domain.Removed),
					change("sp-inactive", domain.Inactive, domain.Removed),
				},
			},
		},
		// An activation does not change the software products.
		"drb-inactive": {
			{Action: domain.Activate, Status: domain.Active, Changes: []domain.SoftwareProductChange{}},
			{
				Action:       domain.Remove,
				Status:       domain.Removed,
				Irreversible: true,
				Changes: []domain.SoftwareProductChange{
					change("sp-under-inactive-brand", domain.Inactive, domain.Removed),
				},
			},
		},
	}, actions.Brands)
}

// Verifies: REG-0225
func TestARemovedBrandHasNoAction(t *testing.T) {
	_, state := execute(t, statusFixture(domain.DataRecipientActive), domain.RemoveDataRecipientBrand{BrandID: "drb-1"})

	actions := domain.PermittedStatusActions(state)

	assert.Equal(t, []recordAction{}, actions.Brands["drb-1"])
}

func TestEachBrandActionGivesTheCommandWithTheSameEffectAsItsCascade(t *testing.T) {
	state := statusFixture(domain.DataRecipientActive)

	for id, brandActions := range domain.PermittedStatusActions(state).Brands {
		for _, a := range brandActions {
			_, after := execute(t, state, a.Action.DataRecipientBrandCommand(id))

			assert.Equal(t, a.Status, brandStatuses(after.DataRecipient)[id], "%s %s", id, a.Action)

			for _, c := range a.Changes {
				assert.Equal(t, c.To, productStatuses(after.DataRecipient)[c.SoftwareProductID], "%s %s", id, a.Action)
			}
		}
	}
}

func brandStatuses(dr domain.DataRecipient) map[string]domain.Status {
	got := map[string]domain.Status{}
	for _, b := range dr.Brands {
		got[b.DataRecipientBrandID] = b.Status
	}

	return got
}

func TestEachActionGivesTheCommandWithTheSameEffectAsItsCascade(t *testing.T) {
	state := statusFixture(domain.DataRecipientActive)

	for _, a := range domain.PermittedStatusActions(state).DataRecipient {
		_, after := execute(t, state, a.Action.Command())

		assert.Equal(t, a.Status, after.DataRecipient.Status, "%s", a.Action)

		for _, c := range a.Changes {
			assert.Equal(t, c.To, productStatuses(after.DataRecipient)[c.SoftwareProductID], "%s", a.Action)
		}
	}
}

func TestTheActionsAreTheOnlyStatusActionsThatTheRegisterKnows(t *testing.T) {
	assert.Equal(t,
		[]domain.DataRecipientAction{domain.Suspend, domain.Revoke, domain.Surrender, domain.Reinstate},
		domain.DataRecipientActions,
	)
	assert.Equal(t,
		[]domain.RecordAction{domain.Activate, domain.Deactivate, domain.Remove},
		domain.RecordActions,
	)
}
