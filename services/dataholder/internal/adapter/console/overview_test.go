package console_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type fixedOverview struct {
	overview app.Overview
	err      error
	at       *time.Time
	newest   *int
}

func (f fixedOverview) Overview(_ context.Context, at time.Time, newest int) (app.Overview, error) {
	if f.at != nil {
		*f.at, *f.newest = at, newest
	}

	return f.overview, f.err
}

// Verifies: DH-1661
func TestTheOverviewGivesTheCountsAndTheChanges(t *testing.T) {
	at := time.Date(2026, 11, 3, 6, 41, 7, 0, time.UTC)

	resp := get(t, console.Ports{Overview: fixedOverview{overview: app.Overview{
		Customers: 128, OpenAccounts: 214, ClosedAccounts: 19, CurrentArrangements: 37, EndedArrangements: 112,
		Registrations: 4, SoftwareProductsNotActive: 1,
		Changes: []app.Change{
			{
				Time: at, Kind: app.ChangeArrangementConfirmed, CustomerNumber: "00000001", ClientID: "client-1",
				ClientName: "Recipient Web", ClientRegistered: true, SharingEndsAt: at.AddDate(1, 0, 0),
			},
			{
				Time: at.Add(-time.Minute), Kind: app.ChangeArrangementEnded, CustomerNumber: "00000002",
				ClientID: "client-1", ClientName: "Recipient Web", ClientRegistered: true,
				EndReason:        domain.EndReasonWithdrawnByConsumer,
				WithdrawalMethod: domain.WithdrawalMethodConsumerDashboard,
			},
			{
				Time: at.Add(-2 * time.Minute), Kind: app.ChangeAccountCreated, CustomerNumber: "00000002",
				ProductName: "Wattle Pocket",
			},
			{Time: at.Add(-3 * time.Minute), Kind: app.ChangeClientDeleted, ClientID: "client-2", ClientName: "Beta"},
			{
				Time: at.Add(-4 * time.Minute), Kind: app.ChangeArrangementEnded, CustomerNumber: "00000003",
				ClientID: "client-2", ClientName: "Beta",
			},
		},
	}}}, "/api/overview")

	require.Equal(t, http.StatusOK, resp.status, resp.body)
	assert.JSONEq(t, `{
		"customerCount": 128, "openAccountCount": 214, "closedAccountCount": 19,
		"currentArrangementCount": 37, "endedArrangementCount": 112,
		"registrationCount": 4, "notActiveSoftwareProductCount": 1,
		"changes": [
			{"time": "2026-11-03T06:41:07Z", "kind": "ARRANGEMENT_CONFIRMED", "customerNumber": "00000001",
				"clientId": "client-1", "clientName": "Recipient Web", "clientRegistered": true,
				"sharingEndsAt": "2027-11-03T06:41:07Z"},
			{"time": "2026-11-03T06:40:07Z", "kind": "ARRANGEMENT_ENDED", "customerNumber": "00000002",
				"clientId": "client-1", "clientName": "Recipient Web", "clientRegistered": true,
				"endReason": "WITHDRAWN_BY_CONSUMER",
				"withdrawalMethod": "CONSUMER_DASHBOARD"},
			{"time": "2026-11-03T06:39:07Z", "kind": "ACCOUNT_CREATED", "customerNumber": "00000002",
				"productName": "Wattle Pocket"},
			{"time": "2026-11-03T06:38:07Z", "kind": "CLIENT_DELETED", "clientId": "client-2", "clientName": "Beta",
				"clientRegistered": false},
			{"time": "2026-11-03T06:37:07Z", "kind": "ARRANGEMENT_ENDED", "customerNumber": "00000003",
				"clientId": "client-2", "clientName": "Beta", "clientRegistered": false}
		]
	}`, resp.body)
}

func TestTheOverviewReadsTheNewestChangesAtTheTimeOfTheRequest(t *testing.T) {
	var (
		at     time.Time
		newest int
	)

	resp := get(t, console.Ports{Overview: fixedOverview{at: &at, newest: &newest}}, "/api/overview")

	require.Equal(t, http.StatusOK, resp.status)
	assert.JSONEq(t, `{
		"customerCount": 0, "openAccountCount": 0, "closedAccountCount": 0, "currentArrangementCount": 0,
		"endedArrangementCount": 0, "registrationCount": 0, "notActiveSoftwareProductCount": 0, "changes": []
	}`, resp.body)
	assert.Equal(t, now, at)
	assert.Equal(t, app.NewestChanges, newest)
}

func TestAnOverviewThatFailsGivesAProblem(t *testing.T) {
	resp := get(t, console.Ports{Overview: fixedOverview{err: errors.New("database down")}}, "/api/overview")

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "database down")
}
