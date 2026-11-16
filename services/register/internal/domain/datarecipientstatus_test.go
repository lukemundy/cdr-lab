package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// statusFixture has products in each status under an active brand, and one product under an inactive brand.
func statusFixture(status domain.DataRecipientStatus) domain.DataRecipientState {
	dr := validDataRecipient()
	dr.Status = status

	active, inactive, removed := domain.Active, domain.Inactive, domain.Removed
	if status != domain.DataRecipientActive {
		active = domain.Inactive
	}

	if status == domain.DataRecipientRevoked || status == domain.DataRecipientSurrendered {
		active, inactive = domain.Removed, domain.Removed
	}

	sp := dr.Brands[0].SoftwareProducts[0]
	dr.Brands[0].SoftwareProducts = []domain.SoftwareProduct{
		withProduct(
			sp,
			"sp-active",
			active,
		),
		withProduct(sp, "sp-inactive", inactive),
		withProduct(sp, "sp-removed", removed),
	}
	dr.Brands = append(dr.Brands, domain.DataRecipientBrand{
		DataRecipientBrandID: "drb-inactive",
		BrandName:            "Inactive App",
		LogoURI:              "https://recipient.example/inactive.png",
		Status:               domain.Inactive,
		SoftwareProducts:     []domain.SoftwareProduct{withProduct(sp, "sp-under-inactive-brand", inactive)},
	})

	return domain.ApplyDataRecipient(domain.DataRecipientState{}, domain.DataRecipientImported{DataRecipient: dr})
}

func withProduct(sp domain.SoftwareProduct, id string, status domain.Status) domain.SoftwareProduct {
	sp.SoftwareProductID = id
	sp.Status = status

	return sp
}

func execute(
	t *testing.T,
	state domain.DataRecipientState,
	command domain.DataRecipientCommand,
) ([]domain.DataRecipientEvent, domain.DataRecipientState) {
	t.Helper()

	events, err := domain.HandleDataRecipient(state, command)
	require.NoError(t, err)

	for _, e := range events {
		state = domain.ApplyDataRecipient(state, e)
	}

	return events, state
}

func productStatuses(dr domain.DataRecipient) map[string]domain.Status {
	got := map[string]domain.Status{}

	for _, b := range dr.Brands {
		for _, sp := range b.SoftwareProducts {
			got[sp.SoftwareProductID] = sp.Status
		}
	}

	return got
}

// Verifies: REG-0227
func TestSuspendMakesActiveSoftwareProductsInactive(t *testing.T) {
	events, state := execute(t, statusFixture(domain.DataRecipientActive), domain.SuspendAccreditation{})

	assert.Equal(t, []domain.DataRecipientEvent{
		domain.AccreditationSuspended{},
		domain.SoftwareProductDeactivated{SoftwareProductID: "sp-active"},
	}, events)
	assert.Equal(t, domain.DataRecipientSuspended, state.DataRecipient.Status)
	assert.Equal(t, map[string]domain.Status{
		"sp-active": domain.Inactive, "sp-inactive": domain.Inactive, "sp-removed": domain.Removed,
		"sp-under-inactive-brand": domain.Inactive,
	}, productStatuses(state.DataRecipient))
}

// Verifies: REG-0228
func TestRevokeAndSurrenderRemoveAllSoftwareProducts(t *testing.T) {
	tests := map[string]struct {
		from    domain.DataRecipientStatus
		command domain.DataRecipientCommand
		event   domain.DataRecipientEvent
		want    domain.DataRecipientStatus
	}{
		"revoke active": {
			domain.DataRecipientActive,
			domain.RevokeAccreditation{},
			domain.AccreditationRevoked{},
			domain.DataRecipientRevoked,
		},
		"revoke suspended": {
			domain.DataRecipientSuspended,
			domain.RevokeAccreditation{},
			domain.AccreditationRevoked{},
			domain.DataRecipientRevoked,
		},
		"surrender active": {
			domain.DataRecipientActive,
			domain.SurrenderAccreditation{},
			domain.AccreditationSurrendered{},
			domain.DataRecipientSurrendered,
		},
		"surrender suspended": {
			domain.DataRecipientSuspended,
			domain.SurrenderAccreditation{},
			domain.AccreditationSurrendered{},
			domain.DataRecipientSurrendered,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events, state := execute(t, statusFixture(tt.from), tt.command)

			want := []domain.DataRecipientEvent{tt.event}
			for _, id := range []string{"sp-active", "sp-inactive", "sp-under-inactive-brand"} {
				want = append(want, domain.SoftwareProductRemoved{SoftwareProductID: id})
			}

			assert.ElementsMatch(t, want, events)
			assert.Equal(t, tt.event, events[0])
			assert.Equal(t, tt.want, state.DataRecipient.Status)

			for id, s := range productStatuses(state.DataRecipient) {
				assert.Equal(t, domain.Removed, s, id)
			}
		})
	}
}

func TestReinstateDoesNotChangeSoftwareProducts(t *testing.T) {
	for _, from := range []domain.DataRecipientStatus{domain.DataRecipientSuspended, domain.DataRecipientRevoked} {
		t.Run(string(from), func(t *testing.T) {
			before := statusFixture(from)

			events, state := execute(t, before, domain.ReinstateAccreditation{})

			assert.Equal(t, []domain.DataRecipientEvent{domain.AccreditationReinstated{}}, events)
			assert.Equal(t, domain.DataRecipientActive, state.DataRecipient.Status)
			assert.Equal(t, productStatuses(before.DataRecipient), productStatuses(state.DataRecipient))
		})
	}
}

// Verifies: REG-0222
func TestDataRecipientTransitionsThatTheStatusRulesDoNotAllowAreRefused(t *testing.T) {
	tests := map[string]struct {
		from    domain.DataRecipientStatus
		command domain.DataRecipientCommand
		wantErr string
	}{
		"suspend suspended": {
			domain.DataRecipientSuspended,
			domain.SuspendAccreditation{},
			`data recipient "dr-1" cannot change from SUSPENDED to SUSPENDED: ` +
				`from SUSPENDED, the allowed statuses are ACTIVE, REVOKED and SURRENDERED`,
		},
		"suspend revoked": {
			domain.DataRecipientRevoked,
			domain.SuspendAccreditation{},
			`data recipient "dr-1" cannot change from REVOKED to SUSPENDED: ` +
				`from REVOKED, the allowed statuses are ACTIVE`,
		},
		"revoke revoked": {
			domain.DataRecipientRevoked,
			domain.RevokeAccreditation{},
			`data recipient "dr-1" cannot change from REVOKED to REVOKED: ` +
				`from REVOKED, the allowed statuses are ACTIVE`,
		},
		"surrender revoked": {
			domain.DataRecipientRevoked,
			domain.SurrenderAccreditation{},
			`data recipient "dr-1" cannot change from REVOKED to SURRENDERED: ` +
				`from REVOKED, the allowed statuses are ACTIVE`,
		},
		"reinstate active": {
			domain.DataRecipientActive,
			domain.ReinstateAccreditation{},
			`data recipient "dr-1" cannot change from ACTIVE to ACTIVE: ` +
				`from ACTIVE, the allowed statuses are SUSPENDED, REVOKED and SURRENDERED`,
		},
		"reinstate surrendered": {
			domain.DataRecipientSurrendered,
			domain.ReinstateAccreditation{},
			`data recipient "dr-1" cannot change from SURRENDERED to ACTIVE: SURRENDERED is final`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataRecipient(statusFixture(tt.from), tt.command)

			require.EqualError(t, err, tt.wantErr)

			var ruleErr domain.StatusRuleError
			assert.ErrorAs(t, err, &ruleErr)
		})
	}
}

// Verifies: REG-0223, REG-0230
func TestSoftwareProductStatusChanges(t *testing.T) {
	tests := map[string]struct {
		command domain.DataRecipientCommand
		event   domain.DataRecipientEvent
		id      string
		want    domain.Status
	}{
		"deactivate active": {
			domain.DeactivateSoftwareProduct{SoftwareProductID: "sp-active"},
			domain.SoftwareProductDeactivated{SoftwareProductID: "sp-active"},
			"sp-active",
			domain.Inactive,
		},
		"remove active": {
			domain.RemoveSoftwareProduct{SoftwareProductID: "sp-active"},
			domain.SoftwareProductRemoved{SoftwareProductID: "sp-active"},
			"sp-active",
			domain.Removed,
		},
		"activate inactive": {
			domain.ActivateSoftwareProduct{SoftwareProductID: "sp-inactive"},
			domain.SoftwareProductActivated{SoftwareProductID: "sp-inactive"},
			"sp-inactive",
			domain.Active,
		},
		"remove inactive": {
			domain.RemoveSoftwareProduct{SoftwareProductID: "sp-inactive"},
			domain.SoftwareProductRemoved{SoftwareProductID: "sp-inactive"},
			"sp-inactive",
			domain.Removed,
		},
		"remove under brand": {
			domain.RemoveSoftwareProduct{SoftwareProductID: "sp-under-inactive-brand"},
			domain.SoftwareProductRemoved{SoftwareProductID: "sp-under-inactive-brand"},
			"sp-under-inactive-brand",
			domain.Removed,
		},
		"deactivate is local": {
			domain.DeactivateSoftwareProduct{SoftwareProductID: "sp-active"},
			domain.SoftwareProductDeactivated{SoftwareProductID: "sp-active"},
			"sp-inactive",
			domain.Inactive,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events, state := execute(t, statusFixture(domain.DataRecipientActive), tt.command)

			assert.Equal(t, []domain.DataRecipientEvent{tt.event}, events)
			assert.Equal(t, tt.want, productStatuses(state.DataRecipient)[tt.id])
			assert.Equal(t, domain.DataRecipientActive, state.DataRecipient.Status)
		})
	}
}

// Verifies: REG-0223
func TestSoftwareProductChangesThatTheStatusRulesDoNotAllowAreRefused(t *testing.T) {
	tests := map[string]struct {
		from    domain.DataRecipientStatus
		command domain.DataRecipientCommand
		wantErr string
	}{
		"activate active": {
			domain.DataRecipientActive,
			domain.ActivateSoftwareProduct{SoftwareProductID: "sp-active"},
			`software product "sp-active" cannot change from ACTIVE to ACTIVE: ` +
				`from ACTIVE, the allowed statuses are INACTIVE and REMOVED`,
		},
		"deactivate inactive": {
			domain.DataRecipientActive,
			domain.DeactivateSoftwareProduct{SoftwareProductID: "sp-inactive"},
			`software product "sp-inactive" cannot change from INACTIVE to INACTIVE: ` +
				`from INACTIVE, the allowed statuses are ACTIVE and REMOVED`,
		},
		"activate removed": {
			domain.DataRecipientActive,
			domain.ActivateSoftwareProduct{SoftwareProductID: "sp-removed"},
			`software product "sp-removed" cannot change from REMOVED to ACTIVE: REMOVED is final`,
		},
		"activate under an inactive brand": {
			domain.DataRecipientActive,
			domain.ActivateSoftwareProduct{SoftwareProductID: "sp-under-inactive-brand"},
			`software product "sp-under-inactive-brand" is ACTIVE, ` +
				`but data recipient brand "drb-inactive" is INACTIVE: ` +
				`a software product of an INACTIVE data recipient brand must be INACTIVE or REMOVED`,
		},
		"activate under a suspended data recipient": {
			domain.DataRecipientSuspended,
			domain.ActivateSoftwareProduct{SoftwareProductID: "sp-active"},
			`software product "sp-active" is ACTIVE, but data recipient "dr-1" is SUSPENDED: ` +
				`a software product of a SUSPENDED data recipient must be INACTIVE or REMOVED`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataRecipient(statusFixture(tt.from), tt.command)

			require.EqualError(t, err, tt.wantErr)

			var ruleErr domain.StatusRuleError
			assert.ErrorAs(t, err, &ruleErr)
		})
	}
}

func TestStatusCommandsForARecordThatDoesNotExistAreNotFound(t *testing.T) {
	tests := map[string]struct {
		state   domain.DataRecipientState
		command domain.DataRecipientCommand
		wantErr string
	}{
		"data recipient": {
			domain.DataRecipientState{}, domain.SuspendAccreditation{}, `data recipient does not exist`,
		},
		"software product": {
			statusFixture(domain.DataRecipientActive),
			domain.DeactivateSoftwareProduct{SoftwareProductID: "sp-unknown"},
			`software product "sp-unknown" does not exist`,
		},
		"software product of a data recipient that does not exist": {
			domain.DataRecipientState{},
			domain.DeactivateSoftwareProduct{SoftwareProductID: "sp-active"},
			`data recipient does not exist`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.HandleDataRecipient(tt.state, tt.command)

			require.EqualError(t, err, tt.wantErr)
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}
