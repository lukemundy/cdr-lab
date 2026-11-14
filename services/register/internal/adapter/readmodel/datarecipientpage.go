package readmodel

import (
	"context"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

var _ app.DataRecipientPageReader = (*Queries)(nil)

var dataRecipientSortColumns = sortColumns{
	app.SortByName:                 `lower(legal_entity_name) COLLATE "C"`,
	app.SortByAccreditationNumber:  `accreditation_number COLLATE "C"`,
	app.SortByStatus:               `status COLLATE "C"`,
	app.SortByBrandCount:           "brand_count",
	app.SortBySoftwareProductCount: "software_product_count",
}

// strpos, not LIKE, so that % and _ in the search are text.
const dataRecipientPage = `FROM data_recipients
	WHERE ($1::text IS NULL OR status = $1)
	AND strpos(lower(legal_entity_name), lower($2)) > 0`

func (q *Queries) DataRecipientPage(
	ctx context.Context,
	query app.DataRecipientPageQuery,
) ([]app.DataRecipientItem, int, error) {
	return queryPage[app.DataRecipientItem](ctx, q, dataRecipientSortColumns, pageQuery{
		selectList: "data_recipient, last_updated, version",
		from:       dataRecipientPage,
		args:       []any{query.Status, query.NameContains},
		sort:       query.Sort,
		descending: query.Descending,
		page:       query.Page,
		pageSize:   query.PageSize,
	})
}
