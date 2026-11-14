package customerstore_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var pageTime = time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

func person(number, first, last, email string, updated time.Time) domain.Customer {
	return domain.Customer{
		CustomerNumber: number, FirstName: first, LastName: last, EmailAddress: email, LastUpdated: updated,
	}
}

func addCustomers(t *testing.T, store *customerstore.Store, customers ...domain.Customer) {
	t.Helper()

	for _, c := range customers {
		require.NoError(t, store.AddCustomer(t.Context(), c))
	}
}

func page(t *testing.T, store *customerstore.Store, query app.CustomerPageQuery) ([]app.CustomerSummary, int) {
	t.Helper()

	if query.Sort == "" {
		query.Sort = app.SortByLastName
	}

	if query.Page == 0 {
		query.Page, query.PageSize = 1, 25
	}

	query.At = pageTime

	got, total, err := store.CustomerPage(t.Context(), query)
	require.NoError(t, err)

	return got, total
}

func numbersOf(page []app.CustomerSummary) []string {
	numbers := make([]string, 0, len(page))
	for _, s := range page {
		numbers = append(numbers, s.Customer.CustomerNumber)
	}

	return numbers
}

func TestACustomerPageGivesTheCustomersAndTheTotal(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	alex := person("00000001", "Alex", "Example", "alex@example.com", pageTime)
	addCustomers(t, store, alex, person("00000002", "", "Sample", "sam@example.com", pageTime))

	got, total := page(t, store, app.CustomerPageQuery{})

	assert.Equal(t, 2, total)
	assert.Equal(t, []app.CustomerSummary{
		{Customer: alex},
		{Customer: person("00000002", "", "Sample", "sam@example.com", pageTime)},
	}, got)
}

func TestTheSearchMatchesAPartOfEachFieldWithNoDifferenceForCase(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	addCustomers(t, store,
		person("12345678", "Alex", "Example", "alex@mail.example", pageTime),
		person("87654321", "Robin", "McSample", "robin@post.example", pageTime),
		person("11112222", "", "Other", "other@else.example", pageTime),
	)

	tests := map[string][]string{
		"3456":         {"12345678"},
		"ROBIN":        {"87654321"},
		"mcsam":        {"87654321"},
		"Post.EXAMPLE": {"87654321"},
		"example":      {"12345678", "87654321", "11112222"},
		// strpos, not LIKE: the characters % and _ are text.
		"%":   {},
		"_":   {},
		"zzz": {},
	}
	for search, want := range tests {
		t.Run(search, func(t *testing.T) {
			got, total := page(t, store, app.CustomerPageQuery{Contains: search})

			assert.Equal(t, want, numbersOf(got))
			assert.Equal(t, len(want), total)
		})
	}
}

func TestTheCustomersSortByEachKeyAndThenByTheCustomerNumber(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	addCustomers(t, store,
		person("00000003", "", "beta", "c@example.com", pageTime.Add(-time.Hour)),
		person("00000001", "", "Beta", "a@example.com", pageTime),
		person("00000002", "", "Alpha", "b@example.com", pageTime.Add(time.Hour)),
	)

	tests := []struct {
		sort       app.CustomerSort
		descending bool
		want       []string
	}{
		{sort: app.SortByLastName, want: []string{"00000002", "00000001", "00000003"}},
		{sort: app.SortByLastName, descending: true, want: []string{"00000001", "00000003", "00000002"}},
		{sort: app.SortByCustomerNumber, want: []string{"00000001", "00000002", "00000003"}},
		{sort: app.SortByCustomerNumber, descending: true, want: []string{"00000003", "00000002", "00000001"}},
		{sort: app.SortByLastUpdated, want: []string{"00000003", "00000001", "00000002"}},
		{sort: app.SortByLastUpdated, descending: true, want: []string{"00000002", "00000001", "00000003"}},
	}
	for _, tt := range tests {
		got, _ := page(t, store, app.CustomerPageQuery{Sort: tt.sort, Descending: tt.descending})

		assert.Equal(t, tt.want, numbersOf(got), "%s descending=%t", tt.sort, tt.descending)
	}
}

func TestAPageAfterTheLastPageGivesNoCustomersAndTheTotal(t *testing.T) {
	store := customerstore.New(newPool(t), nil, nil)
	addCustomers(t, store,
		person("00000001", "", "A", "a@example.com", pageTime),
		person("00000002", "", "B", "b@example.com", pageTime),
		person("00000003", "", "C", "c@example.com", pageTime),
	)

	second, total := page(t, store, app.CustomerPageQuery{Page: 2, PageSize: 2})
	after, afterTotal := page(t, store, app.CustomerPageQuery{Page: 3, PageSize: 2})

	assert.Equal(t, []string{"00000003"}, numbersOf(second))
	assert.Equal(t, 3, total)
	assert.Empty(t, after)
	assert.Equal(t, 3, afterTotal)
}

func arrangementOf(t *testing.T, events *eventstore.Store, id, customerNumber string, confirmed time.Time) {
	t.Helper()

	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, id, domain.CreateArrangement{
		Arrangement: domain.Arrangement{
			ArrangementID: id, CustomerNumber: customerNumber, ClientID: "client-1", SoftwareID: "software-1",
			ClientName: "Recipient Web", Scopes: []string{"openid"}, AccountIDs: []string{},
			ConfirmedAt: confirmed, SharingDurationSeconds: 3600,
		},
	}))
}

func TestTheCountsHaveOnlyTheOpenAccountsAndTheArrangementsThatDidNotEnd(t *testing.T) {
	pool := newPool(t)
	store := customerstore.New(pool, nil, nil)
	events := eventstore.New(pool, arrangementstore.Projections()...)

	addCustomers(t, store,
		person("00000001", "", "A", "a@example.com", pageTime),
		person("00000002", "", "B", "b@example.com", pageTime),
	)

	open, closed := account("111111"), account("222222")
	require.NoError(t, store.AddAccount(t.Context(), "00000001", open))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", closed))
	require.NoError(t, store.AddAccount(t.Context(), "00000001", account("333333")))
	_, err := store.CloseAccount(t.Context(), closed, closedAt, nil)
	require.NoError(t, err)

	arrangementOf(t, events, "current", "00000001", pageTime.Add(-time.Minute))
	arrangementOf(t, events, "expired", "00000001", pageTime.Add(-time.Hour))
	arrangementOf(t, events, "withdrawn", "00000001", pageTime.Add(-time.Minute))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "withdrawn", domain.EndArrangement{
		EndedAt: pageTime.Add(-time.Second), Reason: domain.EndReasonWithdrawnByConsumer,
	}))
	arrangementOf(t, events, "other", "00000002", pageTime.Add(-time.Minute))

	got, _ := page(t, store, app.CustomerPageQuery{Sort: app.SortByCustomerNumber})

	require.Len(t, got, 2)
	assert.Equal(t, 2, got[0].OpenAccounts)
	assert.Equal(t, 1, got[0].CurrentArrangements)
	assert.Equal(t, 0, got[1].OpenAccounts)
	assert.Equal(t, 1, got[1].CurrentArrangements)
}

func TestTheCustomersSortByTheCountsAndThenByTheCustomerNumber(t *testing.T) {
	pool := newPool(t)
	store := customerstore.New(pool, nil, nil)
	events := eventstore.New(pool, arrangementstore.Projections()...)

	addCustomers(t, store,
		person("00000001", "", "A", "a@example.com", pageTime),
		person("00000002", "", "B", "b@example.com", pageTime),
		person("00000003", "", "C", "c@example.com", pageTime),
		person("00000004", "", "D", "d@example.com", pageTime),
	)

	closed := account("222223")
	for number, accounts := range map[string][]domain.Account{
		"00000001": {account("111111")},
		"00000002": {account("222221"), account("222222"), closed},
		"00000003": {account("333331")},
	} {
		for _, a := range accounts {
			require.NoError(t, store.AddAccount(t.Context(), number, a))
		}
	}

	_, err := store.CloseAccount(t.Context(), closed, closedAt, nil)
	require.NoError(t, err)

	// The arrangements of 00000001 ended: the first at the end of its sharing, and the second with an end event.
	arrangementOf(t, events, "expired", "00000001", pageTime.Add(-time.Hour))
	arrangementOf(t, events, "withdrawn", "00000001", pageTime.Add(-time.Minute))
	require.NoError(t, app.ArrangementType.Execute(t.Context(), events, "withdrawn", domain.EndArrangement{
		EndedAt: pageTime.Add(-time.Second), Reason: domain.EndReasonWithdrawnByConsumer,
	}))
	arrangementOf(t, events, "current-2", "00000002", pageTime.Add(-time.Minute))
	arrangementOf(t, events, "current-3a", "00000003", pageTime.Add(-time.Minute))
	arrangementOf(t, events, "current-3b", "00000003", pageTime.Add(-time.Minute))
	arrangementOf(t, events, "current-4", "00000004", pageTime.Add(-time.Minute))

	tests := []struct {
		sort       app.CustomerSort
		descending bool
		want       []string
	}{
		{sort: app.SortByOpenAccountCount, want: []string{"00000004", "00000001", "00000003", "00000002"}},
		{
			sort: app.SortByOpenAccountCount, descending: true,
			want: []string{"00000002", "00000001", "00000003", "00000004"},
		},
		{sort: app.SortByCurrentArrangementCount, want: []string{"00000001", "00000002", "00000004", "00000003"}},
		{
			sort: app.SortByCurrentArrangementCount, descending: true,
			want: []string{"00000003", "00000002", "00000004", "00000001"},
		},
	}
	for _, tt := range tests {
		got, _ := page(t, store, app.CustomerPageQuery{Sort: tt.sort, Descending: tt.descending})

		assert.Equal(t, tt.want, numbersOf(got), "%s descending=%t", tt.sort, tt.descending)
	}
}
