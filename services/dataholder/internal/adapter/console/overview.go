package console

import (
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type overviewBody struct {
	CustomerCount                 int          `json:"customerCount"`
	OpenAccountCount              int          `json:"openAccountCount"`
	ClosedAccountCount            int          `json:"closedAccountCount"`
	CurrentArrangementCount       int          `json:"currentArrangementCount"`
	EndedArrangementCount         int          `json:"endedArrangementCount"`
	RegistrationCount             int          `json:"registrationCount"`
	NotActiveSoftwareProductCount int          `json:"notActiveSoftwareProductCount"`
	Changes                       []changeBody `json:"changes"`
}

type changeBody struct {
	Time             time.Time               `json:"time"`
	Kind             app.ChangeKind          `json:"kind"`
	CustomerNumber   string                  `json:"customerNumber,omitempty"`
	ClientID         string                  `json:"clientId,omitempty"`
	ClientName       string                  `json:"clientName,omitempty"`
	ClientRegistered *bool                   `json:"clientRegistered,omitempty"`
	ProductName      string                  `json:"productName,omitempty"`
	SharingEndsAt    *time.Time              `json:"sharingEndsAt,omitempty"`
	EndReason        domain.EndReason        `json:"endReason,omitempty"`
	WithdrawalMethod domain.WithdrawalMethod `json:"withdrawalMethod,omitempty"`
}

func overview(reader app.OverviewReader, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o, err := reader.Overview(r.Context(), now(), app.NewestChanges)
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		body := overviewBody{
			CustomerCount: o.Customers, OpenAccountCount: o.OpenAccounts, ClosedAccountCount: o.ClosedAccounts,
			CurrentArrangementCount: o.CurrentArrangements, EndedArrangementCount: o.EndedArrangements,
			RegistrationCount: o.Registrations, NotActiveSoftwareProductCount: o.SoftwareProductsNotActive,
			Changes: make([]changeBody, 0, len(o.Changes)),
		}

		for _, c := range o.Changes {
			change := changeBody{
				Time: c.Time, Kind: c.Kind, CustomerNumber: c.CustomerNumber, ClientID: c.ClientID,
				ClientName: c.ClientName, ProductName: c.ProductName, EndReason: c.EndReason,
				WithdrawalMethod: c.WithdrawalMethod,
			}
			if c.ClientID != "" {
				change.ClientRegistered = &c.ClientRegistered
			}

			if !c.SharingEndsAt.IsZero() {
				change.SharingEndsAt = &c.SharingEndsAt
			}

			body.Changes = append(body.Changes, change)
		}

		jsonhttp.Write(w, http.StatusOK, body)
	})
}
