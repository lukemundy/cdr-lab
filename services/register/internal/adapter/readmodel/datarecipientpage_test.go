package readmodel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func named(
	id, name, accreditationNumber string,
	status domain.DataRecipientStatus,
	brands ...[]domain.SoftwareProduct,
) domain.DataRecipient {
	dr := recipient(id, status, brands...)
	dr.LegalEntityName = name
	dr.AccreditationNumber = accreditationNumber

	return dr
}

func seedPageRecipients(t *testing.T) *readmodel.Queries {
	t.Helper()

	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{
		named("dr-1", "Mock Budgeting Pty Ltd", "ADR-300", domain.DataRecipientActive,
			[]domain.SoftwareProduct{product("sp-1", domain.Active)}, nil),
		named("dr-2", "mock savings pty ltd", "ADR-100", domain.DataRecipientSuspended,
			[]domain.SoftwareProduct{
				product("sp-2", domain.Inactive), product("sp-3", domain.Inactive), product("sp-4", domain.Removed),
			}),
		named("dr-3", "Other Pty Ltd", "ADR-200", domain.DataRecipientActive),
		named("dr-4", "100% Pty Ltd", "ADR-500", domain.DataRecipientRevoked,
			[]domain.SoftwareProduct{product("sp-5", domain.Removed)}),
		named("dr-5", "1000 Pty Ltd", "ADR-400", domain.DataRecipientActive),
	}}))

	return readmodel.NewQueries(pool)
}

func idsOf(items []app.DataRecipientItem) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.DataRecipient.LegalEntityID)
	}

	return ids
}

func pageQuery(mutate func(*app.DataRecipientPageQuery)) app.DataRecipientPageQuery {
	q := app.DataRecipientPageQuery{Sort: app.SortByName, Page: 1, PageSize: 10}
	mutate(&q)

	return q
}

func TestDataRecipientPageFiltersSearchesAndCounts(t *testing.T) {
	queries := seedPageRecipients(t)
	active := domain.DataRecipientActive

	tests := map[string]struct {
		query     app.DataRecipientPageQuery
		want      []string
		wantTotal int
	}{
		"all": {pageQuery(func(*app.DataRecipientPageQuery) {}), []string{"dr-4", "dr-5", "dr-1", "dr-2", "dr-3"}, 5},
		"status": {
			pageQuery(func(q *app.DataRecipientPageQuery) { q.Status = &active }),
			[]string{"dr-5", "dr-1", "dr-3"},
			3,
		},
		"part of the name with a different case": {
			pageQuery(func(q *app.DataRecipientPageQuery) { q.NameContains = "MOCK" }),
			[]string{"dr-1", "dr-2"},
			2,
		},
		"status and name": {
			pageQuery(func(q *app.DataRecipientPageQuery) { q.Status = &active; q.NameContains = "mock" }),
			[]string{"dr-1"},
			1,
		},
		"pattern characters are text": {
			pageQuery(func(q *app.DataRecipientPageQuery) { q.NameContains = "0%" }),
			[]string{"dr-4"},
			1,
		},
		"no match": {
			pageQuery(func(q *app.DataRecipientPageQuery) { q.NameContains = "unknown" }),
			[]string{},
			0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, total, err := queries.DataRecipientPage(t.Context(), tt.query)

			require.NoError(t, err)
			assert.Equal(t, tt.want, idsOf(got))
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestDataRecipientPageSortsAndThenUsesTheLegalEntityID(t *testing.T) {
	queries := seedPageRecipients(t)

	tests := map[string]struct {
		sort       app.ParticipantSort
		descending bool
		want       []string
	}{
		"name, with no difference for case": {app.SortByName, false, []string{"dr-4", "dr-5", "dr-1", "dr-2", "dr-3"}},
		"name descending":                   {app.SortByName, true, []string{"dr-3", "dr-2", "dr-1", "dr-5", "dr-4"}},
		"accreditation number": {
			app.SortByAccreditationNumber, false, []string{"dr-2", "dr-3", "dr-1", "dr-5", "dr-4"},
		},
		"status, then ID":            {app.SortByStatus, false, []string{"dr-1", "dr-3", "dr-5", "dr-4", "dr-2"}},
		"status descending, then ID": {app.SortByStatus, true, []string{"dr-2", "dr-4", "dr-1", "dr-3", "dr-5"}},
		"brand count, then ID":       {app.SortByBrandCount, false, []string{"dr-3", "dr-5", "dr-2", "dr-4", "dr-1"}},
		"brand count descending, then ID": {
			app.SortByBrandCount, true, []string{"dr-1", "dr-2", "dr-4", "dr-3", "dr-5"},
		},
		"software product count, then ID": {
			app.SortBySoftwareProductCount, false, []string{"dr-3", "dr-5", "dr-1", "dr-4", "dr-2"},
		},
		"software product count descending, then ID": {
			app.SortBySoftwareProductCount, true, []string{"dr-2", "dr-1", "dr-4", "dr-3", "dr-5"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, _, err := queries.DataRecipientPage(t.Context(), pageQuery(func(q *app.DataRecipientPageQuery) {
				q.Sort = tt.sort
				q.Descending = tt.descending
			}))

			require.NoError(t, err)
			assert.Equal(t, tt.want, idsOf(got))
		})
	}
}

func TestDataRecipientPageGivesOnePageAndTheTotalOfAllPages(t *testing.T) {
	queries := seedPageRecipients(t)

	tests := map[string]struct {
		page int
		want []string
	}{
		"first page":        {1, []string{"dr-4", "dr-5"}},
		"last page":         {3, []string{"dr-3"}},
		"page out of range": {4, []string{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, total, err := queries.DataRecipientPage(t.Context(), pageQuery(func(q *app.DataRecipientPageQuery) {
				q.Page = tt.page
				q.PageSize = 2
			}))

			require.NoError(t, err)
			assert.Equal(t, tt.want, idsOf(got))
			assert.Equal(t, 5, total)
		})
	}
}

func TestDataRecipientPageGivesTheFullDataRecipient(t *testing.T) {
	store, pool := newStore(t)
	dr := recipient("dr-1", domain.DataRecipientActive,
		[]domain.SoftwareProduct{product("sp-1", domain.Active), product("sp-2", domain.Removed)},
		[]domain.SoftwareProduct{product("sp-3", domain.Inactive)})
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataRecipients: []domain.DataRecipient{dr}}))

	got, _, err := readmodel.NewQueries(pool).
		DataRecipientPage(t.Context(), pageQuery(func(*app.DataRecipientPageQuery) {}))

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, dr, got[0].DataRecipient)
	assert.True(t, got[0].LastUpdated.Equal(lastRecordedAt(t, pool, "dr-1")))
}

func TestDataRecipientPageRefusesAnUnknownSort(t *testing.T) {
	_, pool := newStore(t)

	_, _, err := readmodel.NewQueries(pool).DataRecipientPage(
		t.Context(),
		pageQuery(func(q *app.DataRecipientPageQuery) { q.Sort = "unknown" }),
	)

	assert.ErrorContains(t, err, "unknown")
}
