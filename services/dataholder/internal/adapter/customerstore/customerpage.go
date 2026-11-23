package customerstore

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var _ app.CustomerPageReader = (*Store)(nil)

// The "C" collation gives the same order in each database locale.
var customerSortColumns = map[app.CustomerSort]string{
	app.SortByLastName:       `lower(last_name) COLLATE "C"`,
	app.SortByCustomerNumber: `customer_number COLLATE "C"`,
	app.SortByLastUpdated:    "last_updated",
	// Names of the select list: no table has these columns.
	app.SortByOpenAccountCount:        "open_account_count",
	app.SortByCurrentArrangementCount: "current_arrangement_count",
}

// strpos, not LIKE, so that % and _ in the search are text. The search is an argument, because the database spans
// contain the SQL text and not the arguments.
const customerPage = `FROM customers
	WHERE strpos(customer_number, $1) > 0
	OR strpos(lower(first_name), lower($1)) > 0
	OR strpos(lower(last_name), lower($1)) > 0
	OR strpos(lower(email_address), lower($1)) > 0`

func (s *Store) CustomerPage(ctx context.Context, query app.CustomerPageQuery) ([]app.CustomerSummary, int, error) {
	column, ok := customerSortColumns[query.Sort]
	if !ok {
		return nil, 0, fmt.Errorf("unknown sort %q for the customers", query.Sort)
	}

	direction := "ASC"
	if query.Descending {
		direction = "DESC"
	}

	var total int
	if err := s.db.QueryRow(ctx, "SELECT count(*) "+customerPage, query.Contains).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT customer_number, first_name, last_name, email_address, last_updated,
			(SELECT count(*) FROM accounts a
				WHERE a.customer_number = customers.customer_number AND a.open_status = $2) AS open_account_count,
			(SELECT count(*) FROM arrangements r
				WHERE r.customer_number = customers.customer_number AND r.ends_at > $3) AS current_arrangement_count
		%s
		ORDER BY %s %s, customer_number
		LIMIT $4 OFFSET $5`, customerPage, column, direction),
		query.Contains, domain.OpenStatusOpen, query.At, query.PageSize, (query.Page-1)*query.PageSize,
	)
	if err != nil {
		return nil, 0, err
	}

	page, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (app.CustomerSummary, error) {
		var summary app.CustomerSummary

		c := &summary.Customer
		err := row.Scan(
			&c.CustomerNumber, &c.FirstName, &c.LastName, &c.EmailAddress, &c.LastUpdated,
			&summary.OpenAccounts, &summary.CurrentArrangements,
		)
		c.LastUpdated = c.LastUpdated.UTC()

		return summary, err
	})

	return page, total, err
}
