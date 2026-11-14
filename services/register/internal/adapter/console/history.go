package console

import (
	"errors"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type historyPage struct {
	Data         []historyEntry `json:"data"`
	TotalRecords int            `json:"totalRecords"`
}

// The OpenAPI document gives the fields of each event type.
type historyEntry struct {
	StreamVersion       int64     `json:"streamVersion"`
	EventType           string    `json:"eventType"`
	RecordedAt          time.Time `json:"recordedAt"`
	SoftwareProductID   string    `json:"softwareProductId,omitempty"`
	SoftwareProductName string    `json:"softwareProductName,omitempty"`
	Thumbprint          string    `json:"thumbprint,omitempty"`
	BrandID             string    `json:"brandId,omitempty"`
	BrandName           string    `json:"brandName,omitempty"`
}

func history(reader app.HistoryReader, streamType, notFound string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, pageSize, err := consolehttp.ParsePage(r.URL.Query())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		entries, total, err := reader.History(r.Context(), streamType, app.HistoryPageQuery{
			LegalEntityID: r.PathValue("legalEntityId"), Page: page, PageSize: pageSize,
		})

		switch {
		case errors.Is(err, domain.ErrNotFound):
			problem.Write(w, http.StatusNotFound, notFound)
			return
		case err != nil:
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		body := historyPage{Data: make([]historyEntry, 0, len(entries)), TotalRecords: total}
		for _, e := range entries {
			entry := historyEntry{
				StreamVersion: e.StreamVersion,
				EventType:     e.EventType,
				RecordedAt:    e.RecordedAt.UTC(),
				Thumbprint:    e.Thumbprint,
			}
			if e.SoftwareProduct != nil {
				entry.SoftwareProductID, entry.SoftwareProductName = e.SoftwareProduct.ID, e.SoftwareProduct.Name
			}

			if e.Brand != nil {
				entry.BrandID, entry.BrandName = e.Brand.ID, e.Brand.Name
			}

			body.Data = append(body.Data, entry)
		}

		jsonhttp.Write(w, http.StatusOK, body)
	})
}
