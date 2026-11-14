package console

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var registrationSorts = map[string]app.RegistrationSort{
	"clientName":              app.SortRegistrationsByClientName,
	"registeredAt":            app.SortRegistrationsByRegisteredAt,
	"currentArrangementCount": app.SortRegistrationsByCurrentArrangements,
	"arrangementCount":        app.SortRegistrationsByArrangements,
}

// ClientArrangements gives the arrangements of a client ID newest first, and the number of them in all pages.
type ClientArrangements interface {
	ArrangementPageOfClient(
		ctx context.Context,
		clientID string,
		page, pageSize int,
	) ([]app.ClientArrangement, int, error)
}

type registrationList struct {
	Data []registration `json:"data"`
}

type registration struct {
	RegisteredAt            time.Time                    `json:"registeredAt"`
	LastUpdated             time.Time                    `json:"lastUpdated"`
	SoftwareProductStatus   domain.SoftwareProductStatus `json:"softwareProductStatus,omitempty"`
	ArrangementCount        int                          `json:"arrangementCount"`
	CurrentArrangementCount int                          `json:"currentArrangementCount"`
	Metadata                metadata                     `json:"metadata"`
}

type clientArrangementPage struct {
	Data         []clientArrangement `json:"data"`
	TotalRecords int                 `json:"totalRecords"`
}

type clientArrangement struct {
	ArrangementID  string            `json:"arrangementId"`
	CustomerNumber string            `json:"customerNumber"`
	FirstName      string            `json:"firstName,omitempty"`
	LastName       string            `json:"lastName"`
	Status         arrangementStatus `json:"status"`
	ConfirmedAt    time.Time         `json:"confirmedAt"`
	SharingEndsAt  time.Time         `json:"sharingEndsAt"`
}

// metadata has no software_statement, because the console shows no SSA.
type metadata struct {
	ClientID                       string             `json:"client_id"`
	ClientIDIssuedAt               int64              `json:"client_id_issued_at"`
	ClientName                     string             `json:"client_name"`
	ClientDescription              string             `json:"client_description"`
	ClientURI                      string             `json:"client_uri"`
	LegalEntityID                  string             `json:"legal_entity_id,omitempty"`
	LegalEntityName                string             `json:"legal_entity_name,omitempty"`
	OrgID                          string             `json:"org_id"`
	OrgName                        string             `json:"org_name"`
	RedirectURIs                   []string           `json:"redirect_uris"`
	SectorIdentifierURI            string             `json:"sector_identifier_uri,omitempty"`
	LogoURI                        string             `json:"logo_uri"`
	TOSURI                         string             `json:"tos_uri,omitempty"`
	PolicyURI                      string             `json:"policy_uri,omitempty"`
	JWKSURI                        string             `json:"jwks_uri"`
	RevocationURI                  string             `json:"revocation_uri"`
	RecipientBaseURI               string             `json:"recipient_base_uri"`
	TokenEndpointAuthMethod        string             `json:"token_endpoint_auth_method"`
	TokenEndpointAuthSigningAlg    domain.SigningAlg  `json:"token_endpoint_auth_signing_alg"`
	GrantTypes                     []domain.GrantType `json:"grant_types"`
	ResponseTypes                  []string           `json:"response_types"`
	ApplicationType                string             `json:"application_type"`
	IDTokenSignedResponseAlg       domain.SigningAlg  `json:"id_token_signed_response_alg"`
	AuthorizationSignedResponseAlg domain.SigningAlg  `json:"authorization_signed_response_alg"`
	RequestObjectSigningAlg        domain.SigningAlg  `json:"request_object_signing_alg"`
	SoftwareID                     string             `json:"software_id"`
	SoftwareRoles                  string             `json:"software_roles"`
	Scope                          string             `json:"scope"`
}

func registrationOf(record app.RegistrationRecord) registration {
	r := record.Registration

	return registration{
		RegisteredAt:            record.RegisteredAt.UTC(),
		LastUpdated:             record.UpdatedAt.UTC(),
		SoftwareProductStatus:   record.Status,
		ArrangementCount:        record.Arrangements,
		CurrentArrangementCount: record.CurrentArrangements,
		Metadata: metadata{
			ClientID:                       r.ClientID,
			ClientIDIssuedAt:               r.ClientIDIssuedAt,
			ClientName:                     r.ClientName,
			ClientDescription:              r.ClientDescription,
			ClientURI:                      r.ClientURI,
			LegalEntityID:                  r.LegalEntityID,
			LegalEntityName:                r.LegalEntityName,
			OrgID:                          r.OrgID,
			OrgName:                        r.OrgName,
			RedirectURIs:                   r.RedirectURIs,
			SectorIdentifierURI:            r.SectorIdentifierURI,
			LogoURI:                        r.LogoURI,
			TOSURI:                         r.TOSURI,
			PolicyURI:                      r.PolicyURI,
			JWKSURI:                        r.JWKSURI,
			RevocationURI:                  r.RevocationURI,
			RecipientBaseURI:               r.RecipientBaseURI,
			TokenEndpointAuthMethod:        r.TokenEndpointAuthMethod,
			TokenEndpointAuthSigningAlg:    r.TokenEndpointAuthSigningAlg,
			GrantTypes:                     r.GrantTypes,
			ResponseTypes:                  r.ResponseTypes,
			ApplicationType:                r.ApplicationType,
			IDTokenSignedResponseAlg:       r.IDTokenSignedResponseAlg,
			AuthorizationSignedResponseAlg: r.AuthorizationSignedResponseAlg,
			RequestObjectSigningAlg:        r.RequestObjectSigningAlg,
			SoftwareID:                     r.SoftwareID,
			SoftwareRoles:                  r.SoftwareRoles,
			Scope:                          r.Scope,
		},
	}
}

func registrationQuery(r *http.Request, at time.Time) (app.RegistrationListQuery, error) {
	query := app.RegistrationListQuery{At: at}
	v := r.URL.Query()

	var err error

	if query.Sort, err = consolehttp.ParseSort(v, registrationSorts, app.SortRegistrationsByClientName); err != nil {
		return query, err
	}

	query.Descending, err = consolehttp.ParseOrder(v)

	return query, err
}

func registrations(records app.RegistrationRecords, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, err := registrationQuery(r, now())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		found, err := records.RegistrationList(r.Context(), query)
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		list := registrationList{Data: make([]registration, 0, len(found))}
		for _, record := range found {
			list.Data = append(list.Data, registrationOf(record))
		}

		jsonhttp.Write(w, http.StatusOK, list)
	})
}

const registrationNotFound = "The data holder has no registration with this client ID."

func readRegistration(records app.RegistrationRecords, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record, err := records.RegistrationRecord(r.Context(), r.PathValue("clientId"), now())
		if errors.Is(err, app.ErrRegistrationNotFound) {
			problem.Write(w, http.StatusNotFound, registrationNotFound)
			return
		}

		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		jsonhttp.Write(w, http.StatusOK, registrationOf(record))
	})
}

func readClientArrangements(
	records app.RegistrationRecords,
	arrangements ClientArrangements,
	now func() time.Time,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, pageSize, err := consolehttp.ParsePage(r.URL.Query())
		if err != nil {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		at := now()
		clientID := r.PathValue("clientId")

		_, err = records.RegistrationRecord(r.Context(), clientID, at)
		if errors.Is(err, app.ErrRegistrationNotFound) {
			problem.Write(w, http.StatusNotFound, registrationNotFound)
			return
		}

		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		found, total, err := arrangements.ArrangementPageOfClient(r.Context(), clientID, page, pageSize)
		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		list := clientArrangementPage{Data: make([]clientArrangement, 0, len(found)), TotalRecords: total}
		for _, f := range found {
			a := f.State.Arrangement
			list.Data = append(list.Data, clientArrangement{
				ArrangementID:  a.ArrangementID,
				CustomerNumber: a.CustomerNumber,
				FirstName:      f.FirstName,
				LastName:       f.LastName,
				Status:         statusOf(f.State, at),
				ConfirmedAt:    a.ConfirmedAt.UTC(),
				SharingEndsAt:  a.SharingEndsAt().UTC(),
			})
		}

		jsonhttp.Write(w, http.StatusOK, list)
	})
}
