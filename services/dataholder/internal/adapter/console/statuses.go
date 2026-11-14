package console

import (
	"net/http"
	"net/url"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var statusSorts = map[string]app.StatusSort{
	"name":   app.SortStatusesByName,
	"status": app.SortStatusesByStatus,
}

var registrationFilters = []string{"registered", "notRegistered"}

type statusList struct {
	Data       []productStatus `json:"data"`
	ReceivedAt *time.Time      `json:"receivedAt,omitempty"`
}

type productStatus struct {
	SoftwareProductID string                       `json:"softwareProductId"`
	Status            domain.SoftwareProductStatus `json:"status"`
	ClientID          string                       `json:"clientId,omitempty"`
	Name              string                       `json:"name,omitempty"`
}

func softwareProductStatuses(reader app.StatusListReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, err := statusQuery(r.URL.Query())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		list, err := reader.StatusList(r.Context(), query)
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		body := statusList{Data: make([]productStatus, 0, len(list.Products))}
		if !list.ReceivedAt.IsZero() {
			body.ReceivedAt = &list.ReceivedAt
		}

		for _, p := range list.Products {
			body.Data = append(body.Data, productStatus{
				SoftwareProductID: p.SoftwareProductID, Status: p.Status, ClientID: p.ClientID, Name: p.ClientName,
			})
		}

		jsonhttp.Write(w, http.StatusOK, body)
	})
}

func statusQuery(v url.Values) (app.StatusListQuery, error) {
	var query app.StatusListQuery

	registration, err := consolehttp.ParseEnum(v, "registration", registrationFilters)
	if err != nil {
		return query, err
	}

	if registration != nil {
		registered := *registration == "registered"
		query.Registered = &registered
	}

	if query.Sort, err = consolehttp.ParseSort(v, statusSorts, app.SortStatusesByName); err != nil {
		return query, err
	}

	query.Descending, err = consolehttp.ParseOrder(v)

	return query, err
}
