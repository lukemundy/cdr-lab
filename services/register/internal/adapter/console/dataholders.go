package console

import (
	"net/http"
	"net/url"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var (
	dataHolderSorts = map[string]app.ParticipantSort{
		"legalEntityName": app.SortByName,
		"status":          app.SortByStatus,
		"brandCount":      app.SortByBrandCount,
	}
	dataHolderStatuses = []domain.DataHolderStatus{domain.DataHolderActive, domain.DataHolderRemoved}
)

type dataHolderList struct {
	Data         []dataHolderSummary `json:"data"`
	TotalRecords int                 `json:"totalRecords"`
}

type dataHolderSummary struct {
	LegalEntityID   string                  `json:"legalEntityId"`
	LegalEntityName string                  `json:"legalEntityName"`
	Status          domain.DataHolderStatus `json:"status"`
	BrandCount      int                     `json:"brandCount"`
	Industries      []domain.Industry       `json:"industries"`
}

func dataHolders(reader app.DataHolderPageReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, err := dataHolderQuery(r.URL.Query())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		page, total, err := reader.DataHolderPage(r.Context(), query)
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		list := dataHolderList{Data: make([]dataHolderSummary, 0, len(page)), TotalRecords: total}
		for _, item := range page {
			list.Data = append(list.Data, dataHolderSummaryOf(item.DataHolder))
		}

		jsonhttp.Write(w, http.StatusOK, list)
	})
}

func dataHolderSummaryOf(dh domain.DataHolder) dataHolderSummary {
	return dataHolderSummary{
		LegalEntityID:   dh.LegalEntityID,
		LegalEntityName: dh.LegalEntityName,
		Status:          dh.Status,
		BrandCount:      len(dh.Brands),
		Industries:      list(dh.Industries()),
	}
}

func dataHolderQuery(v url.Values) (app.DataHolderPageQuery, error) {
	status, err := consolehttp.ParseEnum(v, "status", dataHolderStatuses)
	if err != nil {
		return app.DataHolderPageQuery{}, err
	}

	params, err := parseList(v, dataHolderSorts)
	if err != nil {
		return app.DataHolderPageQuery{}, err
	}

	return app.DataHolderPageQuery{
		Status:       status,
		NameContains: params.nameContains,
		Sort:         params.sort,
		Descending:   params.descending,
		Page:         params.page,
		PageSize:     params.pageSize,
	}, nil
}
