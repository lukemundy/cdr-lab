package banksite

import (
	"context"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// ConsentReader gives the arrangements of one customer as the consumer dashboard shows them.
type ConsentReader interface {
	// Consents gives app.ErrCustomerNotFound.
	Consents(ctx context.Context, customerNumber string) ([]app.Consent, error)
}

type consentStatus string

const (
	consentCurrent consentStatus = "CURRENT"
	consentEnded   consentStatus = "ENDED"
)

type consentJSON struct {
	ArrangementID          string               `json:"arrangementId"`
	ClientName             string               `json:"clientName"`
	BrandName              string               `json:"brandName"`
	LegalEntityName        string               `json:"legalEntityName,omitempty"`
	Status                 consentStatus        `json:"status"`
	ConfirmedAt            time.Time            `json:"confirmedAt"`
	SharingDurationSeconds int64                `json:"sharingDurationSeconds"`
	SharingEndsAt          time.Time            `json:"sharingEndsAt"`
	EndedAt                *time.Time           `json:"endedAt,omitempty"`
	Accounts               []consentAccountJSON `json:"accounts"`
	DataClusters           []dataClusterJSON    `json:"dataClusters"`
}

// consentAccountJSON has the fields of accountJSON with no balance and no actions.
type consentAccountJSON struct {
	AccountID       string                 `json:"accountId"`
	CreationDate    string                 `json:"creationDate"`
	DisplayName     string                 `json:"displayName"`
	MaskedNumber    string                 `json:"maskedNumber"`
	OpenStatus      domain.OpenStatus      `json:"openStatus"`
	ProductCategory domain.ProductCategory `json:"productCategory"`
	ProductName     string                 `json:"productName"`
}

type dataClusterJSON struct {
	Name             string     `json:"name"`
	Permissions      []string   `json:"permissions"`
	FirstDisclosedAt *time.Time `json:"firstDisclosedAt,omitempty"`
	LastDisclosedAt  *time.Time `json:"lastDisclosedAt,omitempty"`
}

type consentListJSON struct {
	Consents []consentJSON `json:"consents"`
}

func listConsents(reader ConsentReader, s sessions, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerNumber, ok := s.customerOf(r)
		if !ok {
			s.refuse(w, r)
			return
		}

		found, err := reader.Consents(r.Context(), customerNumber)
		if err != nil {
			writeAccountError(w, r, s, err)
			return
		}

		at := now()
		body := consentListJSON{Consents: make([]consentJSON, 0, len(found))}

		for _, consent := range found {
			body.Consents = append(body.Consents, newConsentJSON(consent, at))
		}

		jsonhttp.Write(w, http.StatusOK, body)
	})
}

func newConsentJSON(consent app.Consent, now time.Time) consentJSON {
	a := consent.Arrangement
	view := consentJSON{
		ArrangementID:          a.ArrangementID,
		ClientName:             a.ClientName,
		BrandName:              a.OrgName,
		LegalEntityName:        a.LegalEntityName,
		Status:                 consentCurrent,
		ConfirmedAt:            a.ConfirmedAt.UTC(),
		SharingDurationSeconds: a.SharingDurationSeconds,
		SharingEndsAt:          a.SharingEndsAt().UTC(),
		Accounts:               make([]consentAccountJSON, 0, len(consent.Accounts)),
		DataClusters:           make([]dataClusterJSON, 0, len(consent.Clusters)),
	}

	if consent.Ended(now) {
		view.Status = consentEnded
		view.EndedAt = new(consent.EndsAt().UTC())
	}

	for _, account := range consent.Accounts {
		view.Accounts = append(view.Accounts, consentAccountJSON{
			AccountID:       account.AccountID,
			CreationDate:    account.CreatedAt.UTC().Format(time.DateOnly),
			DisplayName:     account.DisplayName,
			MaskedNumber:    account.MaskedNumber(),
			OpenStatus:      account.OpenStatus,
			ProductCategory: account.ProductCategory,
			ProductName:     account.ProductName,
		})
	}

	for _, cluster := range consent.Clusters {
		c := dataClusterJSON{Name: cluster.Name, Permissions: cluster.Permissions}
		if cluster.Disclosed != nil {
			c.FirstDisclosedAt = new(cluster.Disclosed.First.UTC())
			c.LastDisclosedAt = new(cluster.Disclosed.Last.UTC())
		}

		view.DataClusters = append(view.DataClusters, c)
	}

	return view
}
