package eligibilityjobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/eligibilityjobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/grantstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const waitForWork = 10 * time.Second

// Verifies: DH-0923
func TestTheJobEndsTheArrangementsAtTheTimeOfTheCloseAfterAStopBetweenTheCloseAndTheEnd(t *testing.T) {
	pool, inserter := newInserter(t)
	events := eventstore.New(pool, arrangementstore.Projections()...)
	customers := customerstore.New(pool, eligibilityjobs.Hook(inserter), nil)
	require.NoError(t, customers.AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: "00000001", LastName: "Example", EmailAddress: "alex@example.com", LastUpdated: closedAt,
	}))

	account := domain.Account{
		AccountID: "account-1", DisplayName: "Everyday", ProductCategory: domain.CategoryTransAndSavingsAccounts,
		ProductName: "Mock Everyday Account", BSB: "123456", AccountNumber: "12345678",
		OpenStatus: domain.OpenStatusOpen, CreatedAt: closedAt.Add(-time.Hour),
	}
	require.NoError(t, customers.AddAccount(t.Context(), "00000001", account))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "arrangement-1", domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: "arrangement-1", CustomerNumber: "00000001", ClientID: "client-1",
			Scopes: []string{"openid", "profile"}, AccountIDs: []string{},
			ConfirmedAt: closedAt.Add(-time.Minute), SharingDurationSeconds: int64(365 * 24 * time.Hour / time.Second),
		},
	}))

	// The close commits, and the data holder stops before the request ends the arrangements.
	noOpenAccount, err := customers.CloseAccount(t.Context(), account, closedAt, nil)
	require.NoError(t, err)
	require.True(t, noOpenAccount)

	// The test does not wait for the time of the job or for the River leader.
	tag, err := pool.Exec(t.Context(), "UPDATE river_job SET scheduled_at = now(), state = 'available' WHERE kind = $1",
		eligibilityjobs.Args{}.Kind())
	require.NoError(t, err)
	require.Equal(t, int64(1), tag.RowsAffected())

	workers := river.NewWorkers()
	river.AddWorker(workers, eligibilityjobs.Worker{Arrangements: app.Arrangements{
		Store: events, Grants: grantstore.New(pool), Customers: customers,
		Reader: arrangementstore.NewQueries(pool), Now: time.Now,
	}})
	client, err := jobs.NewWorkerClient(pool, workers)
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))
	t.Cleanup(func() { _ = client.StopAndCancel(context.WithoutCancel(t.Context())) })

	var state domain.ArrangementState

	require.Eventually(t, func() bool {
		state, _, err = app.ArrangementType.Load(t.Context(), events, "arrangement-1")
		require.NoError(t, err)

		return state.EndReason != ""
	}, waitForWork, 50*time.Millisecond)

	assert.Equal(t, domain.EndReasonConsumerNotEligible, state.EndReason)
	assert.Equal(t, closedAt, state.EndedAt)
}
