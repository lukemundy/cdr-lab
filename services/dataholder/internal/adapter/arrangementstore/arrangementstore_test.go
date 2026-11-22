package arrangementstore_test

import (
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
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

var confirmedAt = time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

func arrangement(id, customerNumber string, confirmed time.Time) domain.Arrangement {
	return domain.Arrangement{
		ArrangementID:          id,
		CustomerNumber:         customerNumber,
		ClientID:               "client-1",
		SoftwareID:             "software-1",
		ClientName:             "Recipient Web",
		Scopes:                 []string{"openid", "bank:accounts.basic:read"},
		AccountIDs:             []string{"account-1"},
		ConfirmedAt:            confirmed,
		SharingDurationSeconds: 3600,
	}
}

func create(t *testing.T, store *eventstore.Store, a domain.Arrangement) {
	t.Helper()

	require.NoError(t, app.ArrangementType.Execute(
		t.Context(), store, a.ArrangementID, domain.CreateArrangement{Arrangement: a},
	))
}

func TestTheArrangementsOfACustomerAreInTheSequenceOfTheirConfirmation(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, arrangementstore.Projections()...)
	later := arrangement("arrangement-2", "00000001", confirmedAt.Add(time.Minute))
	earlier := arrangement("arrangement-1", "00000001", confirmedAt)

	create(t, store, later)
	create(t, store, earlier)
	create(t, store, arrangement("arrangement-3", "00000002", confirmedAt))

	got, err := arrangementstore.NewQueries(pool).ArrangementsOf(t.Context(), "00000001")

	require.NoError(t, err)
	assert.Equal(t, []domain.ArrangementState{
		{Created: true, Arrangement: earlier},
		{Created: true, Arrangement: later},
	}, got)
}

func TestTheReadModelHasTheEndOfAnArrangement(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, arrangementstore.Projections()...)
	create(t, store, arrangement("arrangement-1", "00000001", confirmedAt))
	endedAt := confirmedAt.Add(time.Minute + time.Nanosecond)
	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "arrangement-1", domain.WithdrawArrangement{
		WithdrawnAt: endedAt, Method: domain.WithdrawalMethodAlternative,
	}))

	got, err := arrangementstore.NewQueries(pool).ArrangementsOf(t.Context(), "00000001")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, endedAt, got[0].EndedAt)
	assert.Equal(t, domain.EndReasonWithdrawnByConsumer, got[0].EndReason)
}

func TestTheReadModelHasTheMethodOfAWithdrawal(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, arrangementstore.Projections()...)
	create(t, store, arrangement("arrangement-1", "00000001", confirmedAt))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "arrangement-1", domain.WithdrawArrangement{
		WithdrawnAt: confirmedAt.Add(time.Minute), Method: domain.WithdrawalMethodAlternative,
	}))

	got, err := arrangementstore.NewQueries(pool).ArrangementsOf(t.Context(), "00000001")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, domain.EndReasonWithdrawnByConsumer, got[0].EndReason)
	assert.Equal(t, domain.WithdrawalMethodAlternative, got[0].WithdrawalMethod)
}

func TestACustomerWithNoArrangementHasAnEmptyList(t *testing.T) {
	got, err := arrangementstore.NewQueries(newPool(t)).ArrangementsOf(t.Context(), "00000001")

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestProjectAllWritesTheReadModelForArrangementsThatExistBeforeIt(t *testing.T) {
	pool := newPool(t)
	want := arrangement("arrangement-1", "00000001", confirmedAt)
	create(t, eventstore.New(pool), want)

	require.NoError(t, eventstore.New(pool, arrangementstore.Projections()...).ProjectAll(t.Context()))

	got, err := arrangementstore.NewQueries(pool).ArrangementsOf(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Equal(t, []domain.ArrangementState{{Created: true, Arrangement: want}}, got)
}

func TestTheReadModelHasTheOutcomeOfTheNoticeAndEachAttempt(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, arrangementstore.Projections()...)
	create(t, store, arrangement("arrangement-1", "00000001", confirmedAt))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "arrangement-1", domain.WithdrawArrangement{
		WithdrawnAt: confirmedAt.Add(time.Minute), Method: domain.WithdrawalMethodAlternative,
	}))
	first := domain.NoticeAttempt{Attempt: 1, At: confirmedAt.Add(time.Minute + time.Nanosecond), Result: "timeout"}
	last := domain.NoticeAttempt{Attempt: 2, At: confirmedAt.Add(2 * time.Minute), Result: "status 400"}

	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "arrangement-1",
		domain.RecordNoticeAttempt{Attempt: first}))

	pending, err := arrangementstore.NewQueries(pool).ArrangementsOf(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Equal(t, domain.NoticePending, pending[0].NoticeStatus())
	assert.Equal(t, []domain.NoticeAttempt{first}, pending[0].NoticeAttempts)

	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "arrangement-1", domain.RecordNotice{
		RecordedAt: confirmedAt.Add(2 * time.Minute), Status: domain.NoticeFailed, Attempt: last,
	}))

	failed, err := arrangementstore.NewQueries(pool).ArrangementsOf(t.Context(), "00000001")
	require.NoError(t, err)
	assert.Equal(t, domain.NoticeFailed, failed[0].NoticeStatus())
	assert.Equal(t, []domain.NoticeAttempt{first, last}, failed[0].NoticeAttempts)
}

func TestTheArrangementsOfASoftwareProductAreInTheSequenceOfTheirConfirmation(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, arrangementstore.Projections()...)
	later := arrangement("arrangement-2", "00000002", confirmedAt.Add(time.Minute))
	earlier := arrangement("arrangement-1", "00000001", confirmedAt)
	other := arrangement("arrangement-3", "00000001", confirmedAt)
	other.SoftwareID = "software-2"

	create(t, store, later)
	create(t, store, earlier)
	create(t, store, other)

	got, err := arrangementstore.NewQueries(pool).ArrangementsOfSoftwareProduct(t.Context(), "software-1")

	require.NoError(t, err)
	assert.Equal(t, []domain.ArrangementState{
		{Created: true, Arrangement: earlier},
		{Created: true, Arrangement: later},
	}, got)
}

func addCustomer(t *testing.T, pool *pgxpool.Pool, number, first, last string) {
	t.Helper()

	require.NoError(t, customerstore.New(pool, nil, nil).AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: number, FirstName: first, LastName: last, EmailAddress: number + "@mail.test",
		LastUpdated: confirmedAt,
	}))
}

func TestTheArrangementsOfAClientAreNewestFirstInPagesWithTheNameOfTheCustomer(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, arrangementstore.Projections()...)
	addCustomer(t, pool, "00000001", "Alex", "Example")
	addCustomer(t, pool, "00000002", "", "Sample")

	oldest := arrangement("arrangement-1", "00000001", confirmedAt)
	sameTimeB := arrangement("arrangement-b", "00000002", confirmedAt.Add(time.Minute))
	sameTimeA := arrangement("arrangement-a", "00000001", confirmedAt.Add(time.Minute))
	newest := arrangement("arrangement-0", "00000002", confirmedAt.Add(time.Hour))
	other := arrangement("arrangement-9", "00000001", confirmedAt.Add(2*time.Hour))
	other.ClientID = "client-2"

	for _, a := range []domain.Arrangement{oldest, sameTimeB, sameTimeA, newest, other} {
		create(t, store, a)
	}

	queries := arrangementstore.NewQueries(pool)
	first, total, err := queries.ArrangementPageOfClient(t.Context(), "client-1", 1, 3)

	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Equal(t, []app.ClientArrangement{
		{State: domain.ArrangementState{Created: true, Arrangement: newest}, LastName: "Sample"},
		{State: domain.ArrangementState{Created: true, Arrangement: sameTimeA}, FirstName: "Alex", LastName: "Example"},
		{State: domain.ArrangementState{Created: true, Arrangement: sameTimeB}, LastName: "Sample"},
	}, first, "newest first by the time of the confirmation, then by the arrangement ID")

	second, total, err := queries.ArrangementPageOfClient(t.Context(), "client-1", 2, 3)

	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Equal(t, []app.ClientArrangement{
		{State: domain.ArrangementState{Created: true, Arrangement: oldest}, FirstName: "Alex", LastName: "Example"},
	}, second)
}

func TestTheArrangementPageOfAClientGivesTheEndOfAnArrangement(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, arrangementstore.Projections()...)
	addCustomer(t, pool, "00000001", "Alex", "Example")

	a := arrangement("arrangement-1", "00000001", confirmedAt)
	create(t, store, a)

	endedAt := confirmedAt.Add(time.Minute)
	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, a.ArrangementID, domain.WithdrawArrangement{
		WithdrawnAt: endedAt, Method: domain.WithdrawalMethodAlternative,
	}))

	page, _, err := arrangementstore.NewQueries(pool).ArrangementPageOfClient(t.Context(), "client-1", 1, 25)

	require.NoError(t, err)
	require.Len(t, page, 1)
	assert.Equal(t, domain.EndReasonWithdrawnByConsumer, page[0].State.EndReason)
	assert.Equal(t, endedAt, page[0].State.EndedAt.UTC())
}

func TestAClientWithNoArrangementsHasAnEmptyPage(t *testing.T) {
	page, total, err := arrangementstore.NewQueries(newPool(t)).ArrangementPageOfClient(
		t.Context(), "client-zebracorn", 1, 25,
	)

	require.NoError(t, err)
	assert.Empty(t, page)
	assert.Zero(t, total)
}
