package console

import (
	"net/url"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

const notCompleted = "The Register could not complete the request."

type listParams struct {
	nameContains   string
	sort           app.ParticipantSort
	descending     bool
	page, pageSize int
}

func parseList(v url.Values, sorts map[string]app.ParticipantSort) (listParams, error) {
	params := listParams{nameContains: v.Get("q")}

	var err error

	if params.sort, err = consolehttp.ParseSort(v, sorts, app.SortByName); err != nil {
		return params, err
	}

	if params.descending, err = consolehttp.ParseOrder(v); err != nil {
		return params, err
	}

	params.page, params.pageSize, err = consolehttp.ParsePage(v)

	return params, err
}
