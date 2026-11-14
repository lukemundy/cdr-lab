package console

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type dataRecipientDetail struct {
	LegalEntityID       string                     `json:"legalEntityId"`
	LegalEntityName     string                     `json:"legalEntityName"`
	AccreditationNumber string                     `json:"accreditationNumber"`
	AccreditationLevel  domain.AccreditationLevel  `json:"accreditationLevel"`
	LogoURI             string                     `json:"logoUri"`
	Status              domain.DataRecipientStatus `json:"status"`
	LastUpdated         time.Time                  `json:"lastUpdated"`
	Version             int64                      `json:"version"`
	Actions             []dataRecipientAction      `json:"actions"`
	Brands              []dataRecipientBrand       `json:"dataRecipientBrands"`
}

type statusAction[A, S ~string, C any] struct {
	Action       A    `json:"action"`
	Status       S    `json:"status"`
	Irreversible bool `json:"irreversible"`
	Changes      []C  `json:"changes"`
}

type (
	dataRecipientAction   = statusAction[domain.DataRecipientAction, domain.DataRecipientStatus, softwareProductChange]
	recordAction          = statusAction[domain.RecordAction, domain.Status, softwareProductChange]
	dataHolderAction      = statusAction[domain.DataHolderAction, domain.DataHolderStatus, dataHolderBrandChange]
	dataHolderBrandAction = statusAction[domain.RecordAction, domain.Status, dataHolderBrandChange]
)

type dataHolderBrandChange struct {
	DataHolderBrandID string        `json:"dataHolderBrandId,omitempty"`
	InterimID         string        `json:"interimId,omitempty"`
	BrandName         string        `json:"brandName"`
	From              domain.Status `json:"from"`
	To                domain.Status `json:"to"`
}

type softwareProductChange struct {
	SoftwareProductID   string        `json:"softwareProductId"`
	SoftwareProductName string        `json:"softwareProductName"`
	From                domain.Status `json:"from"`
	To                  domain.Status `json:"to"`
}

type dataRecipientBrand struct {
	DataRecipientBrandID string            `json:"dataRecipientBrandId"`
	BrandName            string            `json:"brandName"`
	LogoURI              string            `json:"logoUri"`
	Status               domain.Status     `json:"status"`
	Actions              []recordAction    `json:"actions"`
	SoftwareProducts     []softwareProduct `json:"softwareProducts"`
}

type softwareProduct struct {
	SoftwareProductID            string                           `json:"softwareProductId"`
	SoftwareProductName          string                           `json:"softwareProductName"`
	SoftwareProductDescription   string                           `json:"softwareProductDescription"`
	LogoURI                      string                           `json:"logoUri"`
	Status                       domain.Status                    `json:"status"`
	ClientURI                    string                           `json:"clientUri"`
	SectorIdentifierURI          string                           `json:"sectorIdentifierUri,omitempty"`
	RedirectURIs                 []string                         `json:"redirectUris"`
	TOSURI                       string                           `json:"tosUri,omitempty"`
	PolicyURI                    string                           `json:"policyUri,omitempty"`
	JWKSURI                      string                           `json:"jwksUri"`
	RevocationURI                string                           `json:"revocationUri"`
	RecipientBaseURI             string                           `json:"recipientBaseUri"`
	Scopes                       []string                         `json:"scopes"`
	ClientCertificateThumbprints []string                         `json:"clientCertificateThumbprints"`
	Actions                      []recordAction                   `json:"actions"`
	ClientCertificateActions     []domain.ClientCertificateAction `json:"clientCertificateActions"`
}

type dataHolderDetail struct {
	LegalEntityID   string                  `json:"legalEntityId"`
	LegalEntityName string                  `json:"legalEntityName"`
	LogoURI         string                  `json:"logoUri"`
	Status          domain.DataHolderStatus `json:"status"`
	LastUpdated     time.Time               `json:"lastUpdated"`
	Version         int64                   `json:"version"`
	Actions         []dataHolderAction      `json:"actions"`
	Brands          []dataHolderBrand       `json:"dataHolderBrands"`
}

type dataHolderBrand struct {
	DataHolderBrandID        string                           `json:"dataHolderBrandId,omitempty"`
	InterimID                string                           `json:"interimId,omitempty"`
	BrandName                string                           `json:"brandName"`
	BrandGroup               string                           `json:"brandGroup,omitempty"`
	Industries               []domain.Industry                `json:"industries"`
	LogoURI                  string                           `json:"logoUri"`
	Status                   domain.Status                    `json:"status"`
	Actions                  []dataHolderBrandAction          `json:"actions"`
	ServerCertificateActions []domain.ServerCertificateAction `json:"serverCertificateActions"`
	EndpointDetail           endpointDetail                   `json:"endpointDetail"`
	JWKSEndpoint             string                           `json:"jwksEndpoint,omitempty"`
}

type endpointDetail struct {
	Version          string `json:"version,omitempty"`
	PublicBaseURI    string `json:"publicBaseUri"`
	ProductBaseURI   string `json:"productBaseUri,omitempty"`
	ResourceBaseURI  string `json:"resourceBaseUri,omitempty"`
	InfosecBaseURI   string `json:"infosecBaseUri,omitempty"`
	ExtensionBaseURI string `json:"extensionBaseUri,omitempty"`
	WebsiteURI       string `json:"websiteUri,omitempty"`
}

// The handlers call the reader through a function literal, because a method value of a nil reader panics.
func dataRecipient(reader app.DataRecipientDetailReader) http.Handler {
	return detail(
		func(ctx context.Context, id string) (app.DataRecipientItem, error) {
			return reader.DataRecipient(ctx, id)
		},
		noDataRecipient,
		dataRecipientDetailOf,
	)
}

func dataHolder(reader app.DataHolderDetailReader) http.Handler {
	return detail(
		func(ctx context.Context, id string) (app.DataHolderItem, error) { return reader.DataHolder(ctx, id) },
		noDataHolder,
		dataHolderDetailOf,
	)
}

const (
	noDataRecipient = "The Register has no data recipient with this legal entity ID."
	noDataHolder    = "The Register has no data holder with this legal entity ID."
)

// The detail of a problem does not contain the legal entity ID, because the ID is in the URL that the operator sent.
func detail[Item, Body any](
	read func(ctx context.Context, legalEntityID string) (Item, error),
	notFound string,
	bodyOf func(Item) Body,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item, err := read(r.Context(), r.PathValue("legalEntityId"))

		switch {
		case errors.Is(err, domain.ErrNotFound):
			problem.Write(w, http.StatusNotFound, notFound)
		case err != nil:
			problem.WriteUnexpected(w, r, err, notCompleted)
		default:
			jsonhttp.Write(w, http.StatusOK, bodyOf(item))
		}
	})
}

func dataRecipientDetailOf(item app.DataRecipientItem) dataRecipientDetail {
	dr := item.DataRecipient
	actions := domain.PermittedStatusActions(domain.DataRecipientState{Imported: true, DataRecipient: dr})
	body := dataRecipientDetail{
		LegalEntityID:       dr.LegalEntityID,
		LegalEntityName:     dr.LegalEntityName,
		AccreditationNumber: dr.AccreditationNumber,
		AccreditationLevel:  dr.AccreditationLevel,
		LogoURI:             dr.LogoURI,
		Status:              dr.Status,
		LastUpdated:         item.LastUpdated.UTC(),
		Version:             item.Version,
		Actions:             actionsOf(actions.DataRecipient, softwareProductChangeOf),
		Brands:              make([]dataRecipientBrand, 0, len(dr.Brands)),
	}

	for _, b := range dr.Brands {
		brand := dataRecipientBrand{
			DataRecipientBrandID: b.DataRecipientBrandID,
			BrandName:            b.BrandName,
			LogoURI:              b.LogoURI,
			Status:               b.Status,
			Actions:              actionsOf(actions.Brands[b.DataRecipientBrandID], softwareProductChangeOf),
			SoftwareProducts:     make([]softwareProduct, 0, len(b.SoftwareProducts)),
		}

		for _, sp := range b.SoftwareProducts {
			brand.SoftwareProducts = append(brand.SoftwareProducts, softwareProduct{
				SoftwareProductID:            sp.SoftwareProductID,
				SoftwareProductName:          sp.SoftwareProductName,
				SoftwareProductDescription:   sp.SoftwareProductDescription,
				LogoURI:                      sp.LogoURI,
				Status:                       sp.Status,
				ClientURI:                    sp.ClientURI,
				SectorIdentifierURI:          sp.SectorIdentifierURI,
				RedirectURIs:                 list(sp.RedirectURIs),
				TOSURI:                       sp.TOSURI,
				PolicyURI:                    sp.PolicyURI,
				JWKSURI:                      sp.JWKSURI,
				RevocationURI:                sp.RevocationURI,
				RecipientBaseURI:             sp.RecipientBaseURI,
				Scopes:                       list(sp.Scopes),
				ClientCertificateThumbprints: list(sp.ClientCertificates),
				Actions: actionsOf(
					actions.SoftwareProducts[sp.SoftwareProductID],
					softwareProductChangeOf,
				),
				ClientCertificateActions: domain.PermittedClientCertificateActions(sp),
			})
		}

		body.Brands = append(body.Brands, brand)
	}

	return body
}

func dataHolderDetailOf(item app.DataHolderItem) dataHolderDetail {
	dh := item.DataHolder
	actions := domain.PermittedDataHolderStatusActions(domain.DataHolderState{Imported: true, DataHolder: dh})
	body := dataHolderDetail{
		LegalEntityID:   dh.LegalEntityID,
		LegalEntityName: dh.LegalEntityName,
		LogoURI:         dh.LogoURI,
		Status:          dh.Status,
		LastUpdated:     item.LastUpdated.UTC(),
		Version:         item.Version,
		Actions:         actionsOf(actions.DataHolder, dataHolderBrandChangeOf),
		Brands:          make([]dataHolderBrand, 0, len(dh.Brands)),
	}

	for _, b := range dh.Brands {
		e := b.EndpointDetail
		body.Brands = append(body.Brands, dataHolderBrand{
			DataHolderBrandID:        b.DataHolderBrandID,
			InterimID:                b.InterimID,
			BrandName:                b.BrandName,
			BrandGroup:               b.BrandGroup,
			Industries:               list(b.Industries),
			LogoURI:                  b.LogoURI,
			Status:                   b.Status,
			Actions:                  actionsOf(actions.Brands[b.ID()], dataHolderBrandChangeOf),
			ServerCertificateActions: domain.PermittedServerCertificateActions(b),
			EndpointDetail: endpointDetail{
				Version:          e.Version,
				PublicBaseURI:    e.PublicBaseURI,
				ProductBaseURI:   e.ProductBaseURI,
				ResourceBaseURI:  e.ResourceBaseURI,
				InfosecBaseURI:   e.InfosecBaseURI,
				ExtensionBaseURI: e.ExtensionBaseURI,
				WebsiteURI:       e.WebsiteURI,
			},
			JWKSEndpoint: b.JWKSEndpoint,
		})
	}

	return body
}

func actionsOf[A, S ~string, C, Body any](
	actions []domain.StatusAction[A, S, C],
	changeOf func(C) Body,
) []statusAction[A, S, Body] {
	body := make([]statusAction[A, S, Body], 0, len(actions))
	for _, a := range actions {
		changes := make([]Body, 0, len(a.Changes))
		for _, c := range a.Changes {
			changes = append(changes, changeOf(c))
		}

		body = append(body, statusAction[A, S, Body]{
			Action: a.Action, Status: a.Status, Irreversible: a.Irreversible, Changes: changes,
		})
	}

	return body
}

func softwareProductChangeOf(c domain.SoftwareProductChange) softwareProductChange {
	return softwareProductChange(c)
}

func dataHolderBrandChangeOf(c domain.DataHolderBrandChange) dataHolderBrandChange {
	return dataHolderBrandChange(c)
}

// list gives [] and not null for a nil slice.
func list[T any](values []T) []T {
	return append([]T{}, values...)
}
