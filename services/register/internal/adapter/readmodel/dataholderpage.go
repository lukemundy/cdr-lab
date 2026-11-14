package readmodel

import (
	"context"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

var _ app.DataHolderPageReader = (*Queries)(nil)

var dataHolderSortColumns = sortColumns{
	app.SortByName:       `lower(legal_entity_name) COLLATE "C"`,
	app.SortByStatus:     `status COLLATE "C"`,
	app.SortByBrandCount: "brand_count",
}

// strpos, not LIKE, so that % and _ in the search are text.
const dataHolderPage = `FROM data_holders
	WHERE ($1::text IS NULL OR status = $1)
	AND (strpos(lower(legal_entity_name), lower($2)) > 0
		OR EXISTS (SELECT 1 FROM unnest(brand_names) AS brand_name WHERE strpos(lower(brand_name), lower($2)) > 0))`

func (q *Queries) DataHolderPage(
	ctx context.Context,
	query app.DataHolderPageQuery,
) ([]app.DataHolderItem, int, error) {
	return queryPage[app.DataHolderItem](ctx, q, dataHolderSortColumns, pageQuery{
		selectList: "data_holder, last_updated, version",
		from:       dataHolderPage,
		args:       []any{query.Status, query.NameContains},
		sort:       query.Sort,
		descending: query.Descending,
		page:       query.Page,
		pageSize:   query.PageSize,
	})
}
