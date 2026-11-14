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
	dataRecipientSorts = map[string]app.ParticipantSort{
		"legalEntityName":      app.SortByName,
		"accreditationNumber":  app.SortByAccreditationNumber,
		"status":               app.SortByStatus,
		"brandCount":           app.SortByBrandCount,
		"softwareProductCount": app.SortBySoftwareProductCount,
	}
	dataRecipientStatuses = []domain.DataRecipientStatus{
		domain.DataRecipientActive,
		domain.DataRecipientSuspended,
		domain.DataRecipientRevoked,
		domain.DataRecipientSurrendered,
	}
)

type dataRecipientList struct {
	Data         []dataRecipientSummary `json:"data"`
	TotalRecords int                    `json:"totalRecords"`
}

type dataRecipientSummary struct {
	LegalEntityID        string                     `json:"legalEntityId"`
	LegalEntityName      string                     `json:"legalEntityName"`
	AccreditationNumber  string                     `json:"accreditationNumber"`
	Status               domain.DataRecipientStatus `json:"status"`
	BrandCount           int                        `json:"brandCount"`
	SoftwareProductCount int                        `json:"softwareProductCount"`
}

func dataRecipients(reader app.DataRecipientPageReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, err := dataRecipientQuery(r.URL.Query())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		page, total, err := reader.DataRecipientPage(r.Context(), query)
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		list := dataRecipientList{Data: make([]dataRecipientSummary, 0, len(page)), TotalRecords: total}
		for _, item := range page {
			list.Data = append(list.Data, dataRecipientSummaryOf(item.DataRecipient))
		}

		jsonhttp.Write(w, http.StatusOK, list)
	})
}

func dataRecipientSummaryOf(dr domain.DataRecipient) dataRecipientSummary {
	return dataRecipientSummary{
		LegalEntityID:        dr.LegalEntityID,
		LegalEntityName:      dr.LegalEntityName,
		AccreditationNumber:  dr.AccreditationNumber,
		Status:               dr.Status,
		BrandCount:           len(dr.Brands),
		SoftwareProductCount: dr.SoftwareProductCount(),
	}
}

func dataRecipientQuery(v url.Values) (app.DataRecipientPageQuery, error) {
	status, err := consolehttp.ParseEnum(v, "status", dataRecipientStatuses)
	if err != nil {
		return app.DataRecipientPageQuery{}, err
	}

	params, err := parseList(v, dataRecipientSorts)
	if err != nil {
		return app.DataRecipientPageQuery{}, err
	}

	return app.DataRecipientPageQuery{
		Status:       status,
		NameContains: params.nameContains,
		Sort:         params.sort,
		Descending:   params.descending,
		Page:         params.page,
		PageSize:     params.pageSize,
	}, nil
}
