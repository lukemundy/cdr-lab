package console

import (
	"net/http"
	"net/url"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

var customerSorts = map[string]app.CustomerSort{
	"lastName":                app.SortByLastName,
	"customerNumber":          app.SortByCustomerNumber,
	"lastUpdated":             app.SortByLastUpdated,
	"openAccountCount":        app.SortByOpenAccountCount,
	"currentArrangementCount": app.SortByCurrentArrangementCount,
}

type customerList struct {
	Data         []customerSummary `json:"data"`
	TotalRecords int               `json:"totalRecords"`
}

type customerSummary struct {
	CustomerNumber          string    `json:"customerNumber"`
	FirstName               string    `json:"firstName,omitempty"`
	LastName                string    `json:"lastName"`
	EmailAddress            string    `json:"emailAddress"`
	LastUpdated             time.Time `json:"lastUpdated"`
	OpenAccountCount        int       `json:"openAccountCount"`
	CurrentArrangementCount int       `json:"currentArrangementCount"`
}

func customers(reader app.CustomerPageReader, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, err := customerQuery(r.URL.Query())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		query.At = now()

		page, total, err := reader.CustomerPage(r.Context(), query)
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		list := customerList{Data: make([]customerSummary, 0, len(page)), TotalRecords: total}
		for _, s := range page {
			list.Data = append(list.Data, customerSummary{
				CustomerNumber:          s.Customer.CustomerNumber,
				FirstName:               s.Customer.FirstName,
				LastName:                s.Customer.LastName,
				EmailAddress:            s.Customer.EmailAddress,
				LastUpdated:             s.Customer.LastUpdated,
				OpenAccountCount:        s.OpenAccounts,
				CurrentArrangementCount: s.CurrentArrangements,
			})
		}

		jsonhttp.Write(w, http.StatusOK, list)
	})
}

func customerQuery(v url.Values) (app.CustomerPageQuery, error) {
	query := app.CustomerPageQuery{Contains: v.Get("q")}

	var err error

	if query.Sort, err = consolehttp.ParseSort(v, customerSorts, app.SortByLastName); err != nil {
		return query, err
	}

	if query.Descending, err = consolehttp.ParseOrder(v); err != nil {
		return query, err
	}

	query.Page, query.PageSize, err = consolehttp.ParsePage(v)

	return query, err
}
