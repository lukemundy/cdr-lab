package overviewstore_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/overviewstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

var start = time.Date(2026, 11, 3, 1, 0, 0, 0, time.UTC)

type lab struct {
	pool      *pgxpool.Pool
	customers *customerstore.Store
	events    *eventstore.Store
}

func newLab(t *testing.T) lab {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	projections := append(registrationstore.Projections(), arrangementstore.Projections()...)

	return lab{pool: pool, customers: customerstore.New(pool, nil, nil), events: eventstore.New(pool, projections...)}
}

func (l lab) customer(t *testing.T, number string, at time.Time) {
	t.Helper()

	require.NoError(t, l.customers.AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: number, FirstName: "Alex", LastName: "Example", EmailAddress: number + "@mail.test",
		LastUpdated: at,
	}))
}

func (l lab) account(t *testing.T, number, id string, at time.Time) domain.Account {
	t.Helper()

	account := domain.Account{
		AccountID: id, DisplayName: "Everyday", ProductCategory: "TRANS_AND_SAVINGS_ACCOUNTS",
		ProductName: "Wattle Pocket", BSB: "372100", AccountNumber: id[len(id)-6:], OpenStatus: domain.OpenStatusOpen,
		CreatedAt: at,
	}
	require.NoError(t, l.customers.AddAccount(t.Context(), number, account))

	return account
}

func (l lab) arrangement(t *testing.T, id, number string, at time.Time) domain.Arrangement {
	t.Helper()

	a := domain.Arrangement{
		ArrangementID: id, CustomerNumber: number, ClientID: "client-1", SoftwareID: "software-1",
		ClientName: "Recipient Web", Scopes: []string{"openid"}, ConfirmedAt: at, SharingDurationSeconds: 3600,
	}
	require.NoError(t, app.ArrangementType.Execute(t.Context(), l.events, id, domain.CreateArrangement{Arrangement: a}))

	return a
}

func (l lab) register(t *testing.T, clientID, softwareID, name string) {
	t.Helper()

	require.NoError(t, app.ClientRegistrationType.Execute(t.Context(), l.events, softwareID, domain.RegisterClient{
		Registration: domain.Registration{ClientID: clientID, SoftwareID: softwareID, ClientName: name},
	}))
}

func overview(t *testing.T, l lab, at time.Time, changes int) app.Overview {
	t.Helper()

	got, err := overviewstore.New(l.pool).Overview(t.Context(), at, changes)
	require.NoError(t, err)

	return got
}

func TestAnEmptyDataHolderHasZeroCountsAndNoChanges(t *testing.T) {
	assert.Equal(t, app.Overview{Changes: []app.Change{}}, overview(t, newLab(t), start, app.NewestChanges))
}

func TestTheCountsAgreeWithTheRecords(t *testing.T) {
	l := newLab(t)
	l.customer(t, "00000001", start)
	l.customer(t, "00000002", start)
	closing := l.account(t, "00000001", "account-000001", start)
	l.account(t, "00000001", "account-000002", start)
	l.account(t, "00000002", "account-000003", start)
	_, err := l.customers.CloseAccount(t.Context(), closing, start.Add(time.Minute), nil)
	require.NoError(t, err)
	l.arrangement(t, "arrangement-1", "00000001", start)
	l.arrangement(t, "arrangement-2", "00000001", start.Add(-2*time.Hour))
	l.register(t, "client-1", "software-1", "Recipient Web")
	l.register(t, "client-2", "software-2", "Recipient Mobile")
	l.register(t, "client-3", "software-3", "Recipient Beta")
	require.NoError(t, statusstore.New(l.pool, nil).Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-1": domain.SoftwareProductActive, "software-2": domain.SoftwareProductInactive,
		"software-4": domain.SoftwareProductRemoved,
	}, `"1"`))

	got := overview(t, l, start.Add(time.Hour/2), 0)

	assert.Equal(t, app.Overview{
		Customers: 2, OpenAccounts: 2, ClosedAccounts: 1, CurrentArrangements: 1, EndedArrangements: 1,
		Registrations: 3, SoftwareProductsNotActive: 1, Changes: []app.Change{},
	}, got)
}

func TestTheChangesAreNewestFirstAndHaveNoPersonalInformation(t *testing.T) {
	l := newLab(t)
	l.customer(t, "00000001", start)
	account := l.account(t, "00000001", "account-000001", start.Add(time.Minute))
	confirmed := l.arrangement(t, "arrangement-1", "00000001", start.Add(2*time.Minute))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), l.events, "arrangement-1", domain.WithdrawArrangement{
		WithdrawnAt: start.Add(3 * time.Minute), Method: domain.WithdrawalMethodConsumerDashboard,
	}))
	_, err := l.customers.CloseAccount(t.Context(), account, start.Add(4*time.Minute), nil)
	require.NoError(t, err)

	got := overview(t, l, start.Add(time.Hour), app.NewestChanges)

	assert.Equal(t, []app.Change{
		{
			Time: start.Add(4 * time.Minute), Kind: app.ChangeAccountClosed, CustomerNumber: "00000001",
			ProductName: "Wattle Pocket",
		},
		{
			Time: start.Add(3 * time.Minute), Kind: app.ChangeArrangementEnded, CustomerNumber: "00000001",
			ClientID: "client-1", ClientName: "Recipient Web", EndReason: domain.EndReasonWithdrawnByConsumer,
			WithdrawalMethod: domain.WithdrawalMethodConsumerDashboard,
		},
		{
			Time: start.Add(2 * time.Minute), Kind: app.ChangeArrangementConfirmed, CustomerNumber: "00000001",
			ClientID: "client-1", ClientName: "Recipient Web", SharingEndsAt: confirmed.SharingEndsAt(),
		},
		{
			Time: start.Add(time.Minute), Kind: app.ChangeAccountCreated, CustomerNumber: "00000001",
			ProductName: "Wattle Pocket",
		},
		{Time: start, Kind: app.ChangeCustomerCreated, CustomerNumber: "00000001"},
	}, got.Changes)
}

func TestTheEndOfTheSharingIsAChangeAtItsTime(t *testing.T) {
	l := newLab(t)
	a := l.arrangement(t, "arrangement-1", "00000001", start)

	before := overview(t, l, a.SharingEndsAt().Add(-time.Second), app.NewestChanges)
	at := overview(t, l, a.SharingEndsAt(), app.NewestChanges)

	assert.Equal(t, []app.ChangeKind{app.ChangeArrangementConfirmed}, kindsOf(before))
	assert.Equal(t, []app.ChangeKind{app.ChangeArrangementEnded, app.ChangeArrangementConfirmed}, kindsOf(at))
	assert.Equal(t, a.SharingEndsAt(), at.Changes[0].Time)
	assert.Empty(t, at.Changes[0].EndReason)
	assert.Equal(t, 1, at.EndedArrangements)
}

func TestAChangeWithAClientSaysIfARegistrationHasTheClientIDNow(t *testing.T) {
	l := newLab(t)
	l.arrangement(t, "arrangement-1", "00000001", start)
	l.register(t, "client-1", "software-1", "Recipient Web")

	registered := confirmationOf(t, overview(t, l, start, app.NewestChanges))
	require.NoError(t, app.ClientRegistrationType.Execute(
		t.Context(), l.events, "software-1", domain.DeleteClient{ClientID: "client-1"},
	))
	deleted := confirmationOf(t, overview(t, l, start, app.NewestChanges))

	assert.True(t, registered.ClientRegistered)
	assert.False(t, deleted.ClientRegistered)
}

func TestAChangeOfARegistrationHasTheNameOfTheRegistrationAtThatChange(t *testing.T) {
	l := newLab(t)
	l.register(t, "client-1", "software-1", "Recipient Web")
	require.NoError(t, app.ClientRegistrationType.Execute(t.Context(), l.events, "software-1", domain.UpdateClient{
		ClientID: "client-1", Registration: domain.Registration{SoftwareID: "software-1", ClientName: "Recipient Two"},
	}))
	require.NoError(t, app.ClientRegistrationType.Execute(
		t.Context(), l.events, "software-1", domain.DeleteClient{ClientID: "client-1"},
	))

	got := overview(t, l, time.Now().Add(time.Hour), app.NewestChanges)

	require.Len(t, got.Changes, 3)
	assert.Equal(t, []app.ChangeKind{app.ChangeClientDeleted, app.ChangeClientUpdated, app.ChangeClientRegistered},
		kindsOf(got))
	assert.Equal(t, []string{"Recipient Two", "Recipient Two", "Recipient Web"}, namesOf(got))

	for _, c := range got.Changes {
		assert.Equal(t, "client-1", c.ClientID)
		assert.Empty(t, c.CustomerNumber)
	}
}

func TestTheOverviewGivesOnlyTheNewestChanges(t *testing.T) {
	l := newLab(t)
	for i := range 4 {
		l.customer(t, fmt.Sprintf("%08d", i+1), start.Add(time.Duration(i)*time.Minute))
	}

	got := overview(t, l, start.Add(time.Hour), 2)

	assert.Equal(t, 4, got.Customers)
	assert.Equal(t, []string{"00000004", "00000003"}, []string{
		got.Changes[0].CustomerNumber, got.Changes[1].CustomerNumber,
	})
	assert.Len(t, got.Changes, 2)
}

func confirmationOf(t *testing.T, o app.Overview) app.Change {
	t.Helper()

	for _, c := range o.Changes {
		if c.Kind == app.ChangeArrangementConfirmed {
			return c
		}
	}

	require.Fail(t, "the overview has no confirmation")

	return app.Change{}
}

func kindsOf(o app.Overview) []app.ChangeKind {
	kinds := make([]app.ChangeKind, 0, len(o.Changes))
	for _, c := range o.Changes {
		kinds = append(kinds, c.Kind)
	}

	return kinds
}

func namesOf(o app.Overview) []string {
	names := make([]string, 0, len(o.Changes))
	for _, c := range o.Changes {
		names = append(names, c.ClientName)
	}

	return names
}
