package readmodel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/readmodel"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func namedHolder(id, name string, status domain.DataHolderStatus, brandNames ...string) domain.DataHolder {
	brands := make([][]domain.Industry, 0, len(brandNames))
	for range brandNames {
		brands = append(brands, []domain.Industry{domain.Banking})
	}

	dh := holder(id, status, brands...)
	dh.LegalEntityName = name

	for i, brandName := range brandNames {
		dh.Brands[i].BrandName = brandName
	}

	return dh
}

func seedPageHolders(t *testing.T) *readmodel.Queries {
	t.Helper()

	store, pool := newStore(t)
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{
		namedHolder("dh-1", "Mock Holdings Ltd", domain.DataHolderActive, "Mock Bank", "Mock Mutual"),
		namedHolder("dh-2", "former bank ltd", domain.DataHolderRemoved, "Former Bank", "Brand-1"),
		namedHolder("dh-3", "Bright Energy Pty Ltd", domain.DataHolderActive, "Bright Mutual Energy"),
		namedHolder("dh-4", "100% Ltd", domain.DataHolderActive),
		namedHolder("dh-5", "1000 Ltd", domain.DataHolderActive, "Brand_1"),
	}}))

	return readmodel.NewQueries(pool)
}

func holderIDsOf(items []app.DataHolderItem) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.DataHolder.LegalEntityID)
	}

	return ids
}

func holderPageQuery(mutate func(*app.DataHolderPageQuery)) app.DataHolderPageQuery {
	q := app.DataHolderPageQuery{Sort: app.SortByName, Page: 1, PageSize: 10}
	mutate(&q)

	return q
}

func TestDataHolderPageFiltersAndSearchesTheLegalEntityAndBrandNames(t *testing.T) {
	queries := seedPageHolders(t)
	active := domain.DataHolderActive

	tests := map[string]struct {
		query     app.DataHolderPageQuery
		want      []string
		wantTotal int
	}{
		"all, also with no brands": {
			holderPageQuery(func(*app.DataHolderPageQuery) {}), []string{"dh-4", "dh-5", "dh-3", "dh-2", "dh-1"}, 5,
		},
		"status": {
			holderPageQuery(func(q *app.DataHolderPageQuery) { q.Status = &active }),
			[]string{"dh-4", "dh-5", "dh-3", "dh-1"},
			4,
		},
		"part of the legal entity name with a different case": {
			holderPageQuery(func(q *app.DataHolderPageQuery) { q.NameContains = "HOLDINGS" }),
			[]string{"dh-1"},
			1,
		},
		"part of a brand name, one row for each legal entity": {
			holderPageQuery(func(q *app.DataHolderPageQuery) { q.NameContains = "mutual" }),
			[]string{"dh-3", "dh-1"},
			2,
		},
		"legal entity name or brand name": {
			holderPageQuery(func(q *app.DataHolderPageQuery) { q.NameContains = "bank" }),
			[]string{"dh-2", "dh-1"},
			2,
		},
		"status and name": {
			holderPageQuery(func(q *app.DataHolderPageQuery) { q.Status = &active; q.NameContains = "bank" }),
			[]string{"dh-1"},
			1,
		},
		"% is text": {
			holderPageQuery(func(q *app.DataHolderPageQuery) { q.NameContains = "0%" }),
			[]string{"dh-4"},
			1,
		},
		"_ is text": {
			holderPageQuery(func(q *app.DataHolderPageQuery) { q.NameContains = "d_1" }),
			[]string{"dh-5"},
			1,
		},
		"no match": {
			holderPageQuery(func(q *app.DataHolderPageQuery) { q.NameContains = "unknown" }),
			[]string{},
			0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, total, err := queries.DataHolderPage(t.Context(), tt.query)

			require.NoError(t, err)
			assert.Equal(t, tt.want, holderIDsOf(got))
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestDataHolderPageSortsAndThenUsesTheLegalEntityID(t *testing.T) {
	queries := seedPageHolders(t)

	tests := map[string]struct {
		sort       app.ParticipantSort
		descending bool
		want       []string
	}{
		"name, with no difference for case": {app.SortByName, false, []string{"dh-4", "dh-5", "dh-3", "dh-2", "dh-1"}},
		"name descending":                   {app.SortByName, true, []string{"dh-1", "dh-2", "dh-3", "dh-5", "dh-4"}},
		"status, then ID": {
			app.SortByStatus,
			false,
			[]string{"dh-1", "dh-3", "dh-4", "dh-5", "dh-2"},
		},
		"status descending, then ID": {app.SortByStatus, true, []string{"dh-2", "dh-1", "dh-3", "dh-4", "dh-5"}},
		"brand count, also of removed brands, then ID": {
			app.SortByBrandCount, false, []string{"dh-4", "dh-3", "dh-5", "dh-1", "dh-2"},
		},
		"brand count descending, then ID": {
			app.SortByBrandCount, true, []string{"dh-1", "dh-2", "dh-3", "dh-5", "dh-4"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, _, err := queries.DataHolderPage(t.Context(), holderPageQuery(func(q *app.DataHolderPageQuery) {
				q.Sort = tt.sort
				q.Descending = tt.descending
			}))

			require.NoError(t, err)
			assert.Equal(t, tt.want, holderIDsOf(got))
		})
	}
}

func TestDataHolderPageGivesOnePageAndTheTotalOfAllPages(t *testing.T) {
	queries := seedPageHolders(t)

	tests := map[string]struct {
		page int
		want []string
	}{
		"first page":        {1, []string{"dh-4", "dh-5"}},
		"last page":         {3, []string{"dh-1"}},
		"page out of range": {4, []string{}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, total, err := queries.DataHolderPage(t.Context(), holderPageQuery(func(q *app.DataHolderPageQuery) {
				q.Page = tt.page
				q.PageSize = 2
			}))

			require.NoError(t, err)
			assert.Equal(t, tt.want, holderIDsOf(got))
			assert.Equal(t, 5, total)
		})
	}
}

func TestDataHolderPageGivesTheFullDataHolder(t *testing.T) {
	store, pool := newStore(t)
	dh := holder("dh-1", domain.DataHolderActive, []domain.Industry{domain.Banking}, []domain.Industry{domain.Energy})
	dh.Brands[1].Status = domain.Inactive
	require.NoError(t, app.Seed(t.Context(), store, domain.Participants{DataHolders: []domain.DataHolder{dh}}))

	got, _, err := readmodel.NewQueries(pool).
		DataHolderPage(t.Context(), holderPageQuery(func(*app.DataHolderPageQuery) {}))

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, dh, got[0].DataHolder)
	assert.True(t, got[0].LastUpdated.Equal(lastRecordedAt(t, pool, "dh-1")))
}

func TestDataHolderPageRefusesASortThatTheListDoesNotHave(t *testing.T) {
	_, pool := newStore(t)

	_, _, err := readmodel.NewQueries(pool).DataHolderPage(
		t.Context(),
		holderPageQuery(func(q *app.DataHolderPageQuery) { q.Sort = app.SortByAccreditationNumber }),
	)

	assert.ErrorContains(t, err, "accreditation-number")
}
