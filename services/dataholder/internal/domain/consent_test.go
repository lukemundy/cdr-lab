package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func clusterNames(clusters []domain.DataCluster) []string {
	names := make([]string, 0, len(clusters))
	for _, c := range clusters {
		names = append(names, c.Name)
	}

	return names
}

func TestEachScopeGivesTheDataClusterAndThePermissionsOfTheCDS(t *testing.T) {
	tests := map[string]struct {
		scopes []string
		want   []domain.DataCluster
	}{
		"common:customer.basic:read": {
			scopes: []string{"openid", "common:customer.basic:read"},
			want: []domain.DataCluster{{
				Scope: "common:customer.basic:read", Name: "Name and occupation",
				Permissions: []string{"Name", "Occupation"},
			}},
		},
		"bank:accounts.basic:read": {
			scopes: []string{"openid", "bank:accounts.basic:read"},
			want: []domain.DataCluster{{
				Scope:       "bank:accounts.basic:read",
				Name:        "Account name, type and balance",
				Permissions: []string{"Name of account", "Type of account", "Account balance"},
			}},
		},
		"bank:accounts.detail:read with bank:accounts.basic:read": {
			scopes: []string{"bank:accounts.detail:read", "openid", "bank:accounts.basic:read"},
			want: []domain.DataCluster{
				{
					Scope:       "bank:accounts.basic:read",
					Name:        "Account name, type and balance",
					Permissions: []string{"Name of account", "Type of account", "Account balance"},
				},
				{
					Scope: "bank:accounts.detail:read",
					Name:  "Account numbers and features",
					Permissions: []string{
						"Account number",
						"Interest rates",
						"Fees",
						"Discounts",
						"Account terms",
						"Account mail address",
					},
				},
			},
		},
		"bank:accounts.detail:read without bank:accounts.basic:read": {
			scopes: []string{"openid", "bank:accounts.detail:read"},
			want: []domain.DataCluster{{
				Scope: "bank:accounts.detail:read",
				Name:  "Account balance and details",
				Permissions: []string{
					"Name of account", "Type of account", "Account balance", "Account number", "Interest rates", "Fees",
					"Discounts", "Account terms", "Account mail address",
				},
			}},
		},
		"bank:transactions:read": {
			scopes: []string{"openid", "bank:transactions:read"},
			want: []domain.DataCluster{{
				Scope: "bank:transactions:read",
				Name:  "Transaction details",
				Permissions: []string{
					"Incoming and outgoing transactions", "Amounts", "Dates", "Descriptions of transactions",
					"Who you have sent money to and received money from",
				},
			}},
		},
		"profile": {
			scopes: []string{"openid", "profile"},
			want: []domain.DataCluster{
				{Scope: "profile", Name: "Name", Permissions: []string{"Full name and title(s)"}},
			},
		},
		"openid only": {scopes: []string{"openid"}, want: []domain.DataCluster{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.DataClusters(tt.scopes))
		})
	}
}

func TestTheDataClustersHaveOneSequenceForEachSequenceOfScopes(t *testing.T) {
	want := []string{
		"Name and occupation", "Account name, type and balance", "Account numbers and features", "Transaction details",
		"Name",
	}

	got := domain.DataClusters([]string{
		"profile", "bank:transactions:read", "bank:accounts.detail:read", "openid", "bank:accounts.basic:read",
		"common:customer.basic:read",
	})

	assert.Equal(t, want, clusterNames(got))
}

func TestOnlyARequestWithABankScopeNeedsAccounts(t *testing.T) {
	assert.True(t, domain.NeedsAccounts([]string{"openid", "bank:transactions:read"}))
	assert.False(t, domain.NeedsAccounts([]string{"openid", "profile", "common:customer.basic:read"}))
}

func TestASharingDurationOf24HoursOrLessIsOneOccasion(t *testing.T) {
	assert.True(t, domain.IsOneOccasion(0))
	assert.True(t, domain.IsOneOccasion(24*time.Hour))
	assert.False(t, domain.IsOneOccasion(24*time.Hour+time.Second))
}
