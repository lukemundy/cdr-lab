package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type fixedDisclosures map[app.DisclosureKey]app.DisclosurePeriod

func (fixedDisclosures) AddDisclosure(context.Context, string, string, time.Time) error { return nil }

func (f fixedDisclosures) FirstAndLastDisclosures(
	_ context.Context,
	arrangementIDs []string,
) (map[app.DisclosureKey]app.DisclosurePeriod, error) {
	found := map[app.DisclosureKey]app.DisclosurePeriod{}

	for key, period := range f {
		for _, id := range arrangementIDs {
			if key.ArrangementID == id {
				found[key] = period
			}
		}
	}

	return found, nil
}

func arrangementWith(id string, confirmedAt time.Time, scopes []string, accountIDs ...string) domain.ArrangementState {
	return domain.ArrangementState{Created: true, Arrangement: domain.Arrangement{
		ArrangementID: id, CustomerNumber: "00000001", Scopes: scopes, AccountIDs: accountIDs,
		ConfirmedAt: confirmedAt,
	}}
}

func TestTheConsentsOfACustomerAreNewestFirstWithTheirAccountsAndTheDisclosuresOfEachDataCluster(t *testing.T) {
	data := customerData(t)
	older := arrangementWith("arrangement-1", transactionsNow.Add(-2*time.Hour),
		[]string{"openid", "bank:accounts.basic:read", "bank:transactions:read"}, "account-2", "account-1")
	newer := arrangementWith("arrangement-2", transactionsNow.Add(-time.Hour), []string{"openid", "profile"})
	data.Arrangements = fixedArrangements{"00000001": {older, newer}}
	first, last := transactionsNow.Add(-90*time.Minute), transactionsNow.Add(-30*time.Minute)
	data.Disclosures = fixedDisclosures{
		{ArrangementID: "arrangement-1", Scope: "bank:accounts.basic:read"}: {First: first, Last: last},
		{ArrangementID: "arrangement-3", Scope: "profile"}:                  {First: first, Last: last},
	}

	got, err := data.Consents(t.Context(), "00000001")

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, newer, got[0].ArrangementState)
	assert.Empty(t, got[0].Accounts)
	assert.Equal(t, []app.ClusterDisclosures{
		{DataCluster: domain.DataClusters([]string{"profile"})[0]},
	}, got[0].Clusters)

	assert.Equal(t, older, got[1].ArrangementState)
	require.Len(t, got[1].Accounts, 2)
	assert.Equal(t, "account-2", got[1].Accounts[0].AccountID, "in the sequence of the arrangement")
	assert.Equal(t, domain.OpenStatusClosed, got[1].Accounts[0].OpenStatus)
	assert.Equal(t, "account-1", got[1].Accounts[1].AccountID)

	clusters := domain.DataClusters(older.Arrangement.Scopes)
	assert.Equal(t, []app.ClusterDisclosures{
		{DataCluster: clusters[0], Disclosed: &app.DisclosurePeriod{First: first, Last: last}},
		{DataCluster: clusters[1]},
	}, got[1].Clusters)
}

func TestTheDisclosureOfTheDetailScopeIsInTheClusterOfTheScopesOfTheArrangement(t *testing.T) {
	data := customerData(t)
	data.Arrangements = fixedArrangements{"00000001": {
		arrangementWith("arrangement-1", transactionsNow, []string{"bank:accounts.detail:read"}, "account-1"),
	}}
	data.Disclosures = fixedDisclosures{
		{ArrangementID: "arrangement-1", Scope: "bank:accounts.detail:read"}: {
			First: transactionsNow,
			Last:  transactionsNow,
		},
	}

	got, err := data.Consents(t.Context(), "00000001")

	require.NoError(t, err)
	require.Len(t, got[0].Clusters, 1)
	assert.Equal(t, "Account balance and details", got[0].Clusters[0].Name)
	assert.NotNil(t, got[0].Clusters[0].Disclosed)
}

func TestACustomerWithNoConsentHasAnEmptyList(t *testing.T) {
	data := customerData(t)
	data.Arrangements = fixedArrangements{}
	data.Disclosures = fixedDisclosures{}

	got, err := data.Consents(t.Context(), "00000001")

	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestTheConsentsOfACustomerThatDoesNotExistAreNotFound(t *testing.T) {
	data := customerData(t)
	data.Disclosures = fixedDisclosures{}

	_, err := data.Consents(t.Context(), "00000009")

	require.ErrorIs(t, err, app.ErrCustomerNotFound)
}
