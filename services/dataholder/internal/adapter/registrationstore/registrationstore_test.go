package registrationstore_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/migrations"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	return pool
}

func registration(clientID, softwareID string) domain.Registration {
	return domain.Registration{
		ClientID:                clientID,
		ClientIDIssuedAt:        1791547200,
		ClientName:              "Recipient Web",
		RedirectURIs:            []string{"https://recipient.example/cb"},
		JWKSURI:                 "https://recipient.example/jwks",
		GrantTypes:              []domain.GrantType{domain.GrantClientCredentials, domain.GrantAuthorizationCode},
		ResponseTypes:           []string{"code"},
		RequestObjectSigningAlg: domain.ES256,
		SoftwareID:              softwareID,
		Scope:                   "openid bank:accounts.basic:read cdr:registration",
	}
}

func register(t *testing.T, store *eventstore.Store, r domain.Registration) {
	t.Helper()

	require.NoError(t, app.ClientRegistrationType.Execute(
		t.Context(), store, r.SoftwareID, domain.RegisterClient{Registration: r},
	))
}

func TestARegistrationIsInTheReadModelWhenItsAppendCommits(t *testing.T) {
	pool := newPool(t)
	want := registration("client-1", "software-1")
	register(t, eventstore.New(pool, registrationstore.Projections()...), want)
	register(t, eventstore.New(pool, registrationstore.Projections()...), registration("client-2", "software-2"))

	got, err := registrationstore.NewQueries(pool).Registration(t.Context(), "client-1")

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestAClientIDWithNoRegistrationIsNotFound(t *testing.T) {
	_, err := registrationstore.NewQueries(newPool(t)).Registration(t.Context(), "client-1")

	assert.ErrorIs(t, err, app.ErrRegistrationNotFound)
}

func TestProjectAllWritesTheReadModelForRegistrationsThatExistBeforeIt(t *testing.T) {
	pool := newPool(t)
	register(t, eventstore.New(pool), registration("client-1", "software-1"))

	_, err := registrationstore.NewQueries(pool).Registration(t.Context(), "client-1")
	require.ErrorIs(t, err, app.ErrRegistrationNotFound, "a store with no projection does not write the read model")

	require.NoError(t, eventstore.New(pool, registrationstore.Projections()...).ProjectAll(t.Context()))

	got, err := registrationstore.NewQueries(pool).Registration(t.Context(), "client-1")
	require.NoError(t, err)
	assert.Equal(t, "software-1", got.SoftwareID)
}

func TestARegistrationIsAClientOfTheAuthorisationServer(t *testing.T) {
	pool := newPool(t)
	register(t, eventstore.New(pool, registrationstore.Projections()...), registration("client-1", "software-1"))
	clients := registrationstore.Clients{Reader: registrationstore.NewQueries(pool)}

	client, err := clients.Client(t.Context(), "client-1")

	require.NoError(t, err)
	assert.Equal(t, authserver.DataHolderClient{
		ID:                      "client-1",
		JWKSURI:                 "https://recipient.example/jwks",
		Scopes:                  []string{"openid", "bank:accounts.basic:read", "cdr:registration"},
		GrantTypes:              []string{"client_credentials", "authorization_code"},
		RedirectURIs:            []string{"https://recipient.example/cb"},
		RequestObjectSigningAlg: "ES256",
	}, client)

	_, err = clients.Client(t.Context(), "client-2")
	assert.ErrorIs(t, err, authserver.ErrClientNotFound)
}

type failingReader struct{ err error }

func (r failingReader) Registration(context.Context, string) (domain.Registration, error) {
	return domain.Registration{}, r.err
}

func TestAReadFailureIsNotAClientThatDoesNotExist(t *testing.T) {
	failure := errors.New("database failure")

	_, err := registrationstore.Clients{Reader: failingReader{err: failure}}.Client(t.Context(), "client-1")

	require.ErrorIs(t, err, failure)
	assert.NotErrorIs(t, err, authserver.ErrClientNotFound)
}

func TestTheListGivesEachRegistrationByItsClientNameWithItsTimesAndTheStatusOfItsSoftwareProduct(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, registrationstore.Projections()...)
	before := time.Now().Truncate(time.Microsecond)

	zulu := registration("client-1", "software-1")
	zulu.ClientName = "Zulu Web"
	register(t, store, zulu)
	register(t, store, registration("client-2", "software-2"))

	updated := registration("", "software-1")
	updated.ClientName = "Alpha Web"
	require.NoError(t, app.ClientRegistrationType.Execute(
		t.Context(), store, "software-1", domain.UpdateClient{ClientID: "client-1", Registration: updated},
	))

	require.NoError(t, statusstore.New(pool, nil).Save(t.Context(), map[string]domain.SoftwareProductStatus{
		"software-2": domain.SoftwareProductInactive,
	}, `"v1"`))

	list, err := registrationstore.NewQueries(pool).RegistrationList(t.Context(), byClientName)

	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "client-1", list[0].Registration.ClientID)
	assert.Equal(t, "Alpha Web", list[0].Registration.ClientName)
	assert.Empty(t, list[0].Status, "the data holder has no status for the software product")
	assert.WithinRange(t, list[0].RegisteredAt, before, time.Now())
	assert.True(t, list[0].UpdatedAt.After(list[0].RegisteredAt), "the update changes the time of the last change")
	assert.Equal(t, time.UTC, list[0].RegisteredAt.Location())

	assert.Equal(t, "client-2", list[1].Registration.ClientID)
	assert.Equal(t, domain.SoftwareProductInactive, list[1].Status)
	assert.Equal(t, list[1].RegisteredAt, list[1].UpdatedAt, "a registration with no update has one time")
}

func TestARegistrationAfterADeletionHasTheTimeOfTheNewRegistration(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, registrationstore.Projections()...)
	register(t, store, registration("client-1", "software-1"))
	require.NoError(t, app.ClientRegistrationType.Execute(
		t.Context(), store, "software-1", domain.DeleteClient{ClientID: "client-1"},
	))

	before := time.Now().Truncate(time.Microsecond)

	register(t, store, registration("client-2", "software-1"))

	queries := registrationstore.NewQueries(pool)
	record, err := queries.RegistrationRecord(t.Context(), "client-2", time.Now())
	require.NoError(t, err)
	assert.Equal(t, "software-1", record.Registration.SoftwareID)
	assert.WithinRange(t, record.RegisteredAt, before, time.Now())
	assert.Equal(t, record.RegisteredAt, record.UpdatedAt)

	_, err = queries.RegistrationRecord(t.Context(), "client-1", time.Now())
	assert.ErrorIs(t, err, app.ErrRegistrationNotFound, "a deleted registration is not found")
}

var (
	at           = time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	byClientName = app.RegistrationListQuery{Sort: app.SortRegistrationsByClientName, At: at}
)

func arrangement(id, clientID string, confirmed time.Time, sharing time.Duration) domain.Arrangement {
	return domain.Arrangement{
		ArrangementID: id, CustomerNumber: "00000001", ClientID: clientID, SoftwareID: "software-of-" + clientID,
		ClientName: "Recipient Web", Scopes: []string{"openid"}, AccountIDs: []string{}, ConfirmedAt: confirmed,
		SharingDurationSeconds: int64(sharing.Seconds()),
	}
}

func createArrangement(t *testing.T, store *eventstore.Store, a domain.Arrangement) {
	t.Helper()

	require.NoError(t, app.ArrangementType.Execute(
		t.Context(), store, a.ArrangementID, domain.CreateArrangement{Arrangement: a},
	))
}

func projections() []eventstore.Projection {
	return append(registrationstore.Projections(), arrangementstore.Projections()...)
}

type counts struct{ all, current int }

func countsOf(t *testing.T, pool *pgxpool.Pool) map[string]counts {
	t.Helper()

	list, err := registrationstore.NewQueries(pool).RegistrationList(t.Context(), byClientName)
	require.NoError(t, err)

	got := map[string]counts{}
	for _, r := range list {
		got[r.Registration.ClientID] = counts{all: r.Arrangements, current: r.CurrentArrangements}
	}

	return got
}

func TestARegistrationHasTheNumberOfAllTheArrangementsOfItsClientAndOfTheArrangementsThatDidNotEnd(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, projections()...)
	register(t, store, registration("client-1", "software-1"))
	register(t, store, registration("client-2", "software-2"))

	createArrangement(t, store, arrangement("current", "client-1", at.Add(-time.Hour), 24*time.Hour))
	createArrangement(t, store, arrangement("ended", "client-1", at.Add(-time.Hour), 24*time.Hour))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), store, "ended", domain.EndArrangement{
		EndedAt: at.Add(-time.Minute), Reason: domain.EndReasonWithdrawnByConsumer,
	}))
	// The sharing ends at the time of the request, so the arrangement ended with no event.
	createArrangement(t, store, arrangement("expired", "client-1", at.Add(-time.Hour), time.Hour))
	createArrangement(t, store, arrangement("other", "client-3", at.Add(-time.Hour), 24*time.Hour))

	assert.Equal(t, map[string]counts{"client-1": {all: 3, current: 1}, "client-2": {}}, countsOf(t, pool))

	record, err := registrationstore.NewQueries(pool).RegistrationRecord(t.Context(), "client-1", at)
	require.NoError(t, err)
	assert.Equal(t, 3, record.Arrangements)
	assert.Equal(t, 1, record.CurrentArrangements)
}

func TestAnUpdateOfARegistrationKeepsTheNumberOfTheArrangementsOfItsClient(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, projections()...)
	register(t, store, registration("client-1", "software-1"))
	createArrangement(t, store, arrangement("arrangement-1", "client-1", at.Add(-time.Hour), 24*time.Hour))

	require.NoError(t, app.ClientRegistrationType.Execute(
		t.Context(), store, "software-1",
		domain.UpdateClient{ClientID: "client-1", Registration: registration("", "software-1")},
	))

	assert.Equal(t, map[string]counts{"client-1": {all: 1, current: 1}}, countsOf(t, pool))
}

func TestProjectAllCountsTheArrangementsOfEachClientInEachSequenceOfTheProjections(t *testing.T) {
	sequences := map[string][]eventstore.Projection{
		"registrations first": projections(),
		"arrangements first":  append(arrangementstore.Projections(), registrationstore.Projections()...),
	}
	for name, sequence := range sequences {
		t.Run(name, func(t *testing.T) {
			pool := newPool(t)
			events := eventstore.New(pool)
			register(t, events, registration("client-1", "software-1"))
			createArrangement(t, events, arrangement("arrangement-1", "client-1", at.Add(-time.Hour), time.Minute))
			createArrangement(t, events, arrangement("arrangement-2", "client-1", at.Add(-time.Hour), 24*time.Hour))

			require.NoError(t, eventstore.New(pool, sequence...).ProjectAll(t.Context()))

			assert.Equal(t, map[string]counts{"client-1": {all: 2, current: 1}}, countsOf(t, pool))
		})
	}
}

func clientIDs(list []app.RegistrationRecord) []string {
	ids := make([]string, 0, len(list))
	for _, r := range list {
		ids = append(ids, r.Registration.ClientID)
	}

	return ids
}

func TestTheListSortsByEachKeyInBothDirectionsAndThenByTheClientID(t *testing.T) {
	pool := newPool(t)
	store := eventstore.New(pool, projections()...)

	// client-b and client-c have the same name in a different case, and the same numbers of arrangements.
	for _, r := range []struct{ clientID, name string }{
		{"client-c", "beta web"}, {"client-a", "Zulu Web"}, {"client-b", "Beta Web"},
	} {
		reg := registration(r.clientID, "software-of-"+r.clientID)
		reg.ClientName = r.name
		register(t, store, reg)
	}

	createArrangement(t, store, arrangement("a-1", "client-a", at.Add(-time.Hour), 24*time.Hour))
	createArrangement(t, store, arrangement("a-2", "client-a", at.Add(-time.Hour), 24*time.Hour))
	createArrangement(t, store, arrangement("a-3", "client-a", at.Add(-time.Hour), time.Hour))
	createArrangement(t, store, arrangement("b-1", "client-b", at.Add(-time.Hour), time.Hour))
	createArrangement(t, store, arrangement("c-1", "client-c", at.Add(-time.Hour), time.Hour))

	queries := registrationstore.NewQueries(pool)

	tests := []struct {
		sort       app.RegistrationSort
		descending bool
		want       []string
	}{
		{app.SortRegistrationsByClientName, false, []string{"client-b", "client-c", "client-a"}},
		{app.SortRegistrationsByClientName, true, []string{"client-a", "client-b", "client-c"}},
		{app.SortRegistrationsByRegisteredAt, false, []string{"client-c", "client-a", "client-b"}},
		{app.SortRegistrationsByRegisteredAt, true, []string{"client-b", "client-a", "client-c"}},
		{app.SortRegistrationsByCurrentArrangements, false, []string{"client-b", "client-c", "client-a"}},
		{app.SortRegistrationsByCurrentArrangements, true, []string{"client-a", "client-b", "client-c"}},
		{app.SortRegistrationsByArrangements, false, []string{"client-b", "client-c", "client-a"}},
		{app.SortRegistrationsByArrangements, true, []string{"client-a", "client-b", "client-c"}},
	}
	for _, tt := range tests {
		list, err := queries.RegistrationList(
			t.Context(), app.RegistrationListQuery{Sort: tt.sort, Descending: tt.descending, At: at},
		)

		require.NoError(t, err)
		assert.Equal(t, tt.want, clientIDs(list), "%s descending %t", tt.sort, tt.descending)
	}
}

func TestTheListRefusesAnUnknownSort(t *testing.T) {
	_, err := registrationstore.NewQueries(newPool(t)).RegistrationList(
		t.Context(), app.RegistrationListQuery{Sort: "zebracorn", At: at},
	)

	assert.Error(t, err)
}
