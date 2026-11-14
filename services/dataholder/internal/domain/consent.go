package domain

import (
	"slices"
	"strings"
	"time"
)

const (
	scopeCustomerBasic  = "common:customer.basic:read"
	scopeAccountsBasic  = "bank:accounts.basic:read"
	scopeAccountsDetail = "bank:accounts.detail:read"
	scopeTransactions   = "bank:transactions:read"
	scopeProfile        = "profile"
	// A request with a scope that has this prefix asks for data of accounts.
	bankScopePrefix = "bank:"
)

// oneOccasionLimit is the longest sharing duration that the CDS gives as a disclosure on one occasion.
const oneOccasionLimit = 24 * time.Hour

// DataCluster is a group of data in the language that the CDS requires a consumer to see, word for word.
type DataCluster struct {
	// In one arrangement, each scope selects at most one cluster.
	Scope       string
	Name        string
	Permissions []string
}

var (
	accountsBasicPermissions  = []string{"Name of account", "Type of account", "Account balance"}
	accountsDetailPermissions = []string{
		"Account number", "Interest rates", "Fees", "Discounts", "Account terms", "Account mail address",
	}
)

// DataClusters gives the clusters in the same sequence for each sequence of the scopes.
func DataClusters(scopes []string) []DataCluster {
	has := func(scope string) bool { return slices.Contains(scopes, scope) }
	clusters := []DataCluster{}

	if has(scopeCustomerBasic) {
		clusters = append(
			clusters,
			DataCluster{
				Scope: scopeCustomerBasic, Name: "Name and occupation", Permissions: []string{"Name", "Occupation"},
			},
		)
	}

	if has(scopeAccountsBasic) {
		clusters = append(
			clusters,
			DataCluster{
				Scope:       scopeAccountsBasic,
				Name:        "Account name, type and balance",
				Permissions: accountsBasicPermissions,
			},
		)
	}

	switch {
	case has(scopeAccountsDetail) && has(scopeAccountsBasic):
		clusters = append(
			clusters,
			DataCluster{
				Scope:       scopeAccountsDetail,
				Name:        "Account numbers and features",
				Permissions: accountsDetailPermissions,
			},
		)
	case has(scopeAccountsDetail):
		// The CDS gives the detail scope without the basic scope the permissions of the two scopes.
		clusters = append(clusters, DataCluster{
			Scope:       scopeAccountsDetail,
			Name:        "Account balance and details",
			Permissions: slices.Concat(accountsBasicPermissions, accountsDetailPermissions),
		})
	}

	if has(scopeTransactions) {
		clusters = append(clusters, DataCluster{
			Scope: scopeTransactions,
			Name:  "Transaction details",
			Permissions: []string{
				"Incoming and outgoing transactions", "Amounts", "Dates", "Descriptions of transactions",
				"Who you have sent money to and received money from",
			},
		})
	}

	if has(scopeProfile) {
		clusters = append(
			clusters,
			DataCluster{Scope: scopeProfile, Name: "Name", Permissions: []string{"Full name and title(s)"}},
		)
	}

	return clusters
}

func NeedsAccounts(scopes []string) bool {
	return slices.ContainsFunc(scopes, func(scope string) bool { return strings.HasPrefix(scope, bankScopePrefix) })
}

func IsOneOccasion(sharingDuration time.Duration) bool {
	return sharingDuration <= oneOccasionLimit
}
