// Package httpapi serves the public CDR Register APIs.
package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var cdsInvalidIndustry = cds.Error{
	Code:  "urn:au-cds:error:cds-register:Field/InvalidIndustry",
	Title: "Invalid Industry",
}

type Readers interface {
	app.DataHolderBrandReader
	app.StatusReader
	app.DataRecipientReader
}

// KeySource gives the signing keys of the Software Statement Assertions.
type KeySource func(context.Context) ([]authserver.SigningKey, error)

func Register(mux *http.ServeMux, readers Readers, ssaKeys KeySource, baseURL string) {
	baseURL = strings.TrimSuffix(baseURL, "/")

	routes := httproute.New(mux)

	routes.Handle("GET /cdr-register/v1/jwks", jwks(ssaKeys))
	routes.With(cds.Versioned(2, 2)).
		Handle("GET /cdr-register/v1/{industry}/data-holders/brands/summary", dataHolderBrandsSummary(baseURL, readers))
	routes.With(cds.Versioned(2, 2)).
		Handle("GET /cdr-register/v1/{industry}/data-holders/status", dataHolderStatuses(baseURL, readers))
	routes.With(cds.Versioned(3, 3), recipientIndustry).
		Handle(
			"GET /cdr-register/v1/{industry}/data-recipients/status",
			list(baseURL, readers.DataRecipientStatuses, dataRecipientStatus),
		)
	routes.With(cds.Versioned(3, 3), recipientIndustry).
		Handle(
			"GET /cdr-register/v1/{industry}/data-recipients/brands/software-products/status",
			list(baseURL, readers.SoftwareProductStatuses, softwareProductStatus),
		)
	routes.With(cds.Versioned(4, 4), recipientIndustry).
		Handle("GET /cdr-register/v1/{industry}/data-recipients", list(baseURL, readers.DataRecipients, dataRecipient))

	mux.HandleFunc("/", notFound)
}

func jwks(keys KeySource) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signingKeys, err := keys(r.Context())
		if err != nil {
			unexpected(w, r, err)
			return
		}

		writeCacheable(w, r, authserver.PublicJWKS(signingKeys))
	})
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	cds.WriteError(w, http.StatusNotFound, cds.ResourceNotFound, "The URL is not a CDR Register API endpoint.")
}

func dataHolderStatuses(baseURL string, statuses app.StatusReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		industry, ok := holderIndustry(r)
		if !ok {
			cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, "industry")
			return
		}

		read := func(ctx context.Context) ([]app.DataHolderStatusItem, error) {
			return statuses.DataHolderStatuses(ctx, industry)
		}
		list(baseURL, read, dataHolderStatus).ServeHTTP(w, r)
	})
}

func dataHolderBrandsSummary(baseURL string, brands app.DataHolderBrandReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		industry, ok := holderIndustry(r)
		if !ok {
			// Only this endpoint gives 404: here the CDS error list and the OpenAPI specification agree.
			cds.WriteError(w, http.StatusNotFound, cdsInvalidIndustry, "industry")
			return
		}

		read := func(ctx context.Context) ([]app.DataHolderBrandItem, error) {
			return brands.DataHolderBrands(ctx, industry)
		}
		list(baseURL, read, dataHolderBrandSummary).ServeHTTP(w, r)
	})
}

// holderIndustry gives a nil industry for all.
func holderIndustry(r *http.Request) (*domain.Industry, bool) {
	v := r.PathValue("industry")
	if v == "all" {
		return nil, true
	}

	industry, ok := domain.ParseIndustry(v)

	return &industry, ok
}

// The summary maps each field explicitly, because the brand also has
// the endpoints that only Get Data Holder Brands shows.
type dataHolderBrandSummaryJSON struct {
	DataHolderBrandID string   `json:"dataHolderBrandId,omitempty"`
	InterimID         string   `json:"interimId,omitempty"`
	BrandName         string   `json:"brandName"`
	BrandGroup        string   `json:"brandGroup,omitempty"`
	PublicBaseURI     string   `json:"publicBaseUri"`
	ProductBaseURI    string   `json:"productBaseUri,omitempty"`
	LogoURI           string   `json:"logoUri"`
	Industries        []string `json:"industries"`
	LastUpdated       string   `json:"lastUpdated"`
}

func dataHolderBrandSummary(item app.DataHolderBrandItem) dataHolderBrandSummaryJSON {
	brand := item.Brand

	return dataHolderBrandSummaryJSON{
		DataHolderBrandID: brand.DataHolderBrandID,
		InterimID:         brand.InterimID,
		BrandName:         brand.BrandName,
		BrandGroup:        brand.BrandGroup,
		PublicBaseURI:     brand.EndpointDetail.PublicBaseURI,
		ProductBaseURI:    brand.EndpointDetail.ProductBaseURI,
		LogoURI:           brand.LogoURI,
		Industries:        industryNames(brand.Industries),
		LastUpdated:       item.LastUpdated.UTC().Format(time.RFC3339),
	}
}

func industryNames(industries []domain.Industry) []string {
	names := make([]string, 0, len(industries))
	for _, industry := range industries {
		names = append(names, string(industry))
	}

	return names
}

type dataHolderStatusJSON struct {
	LegalEntityID string `json:"legalEntityId"`
	Status        string `json:"status"`
}

func dataHolderStatus(item app.DataHolderStatusItem) dataHolderStatusJSON {
	return dataHolderStatusJSON{LegalEntityID: item.LegalEntityID, Status: string(item.Status)}
}

type dataRecipientStatusJSON struct {
	LegalEntityID string `json:"legalEntityId"`
	Status        string `json:"status"`
}

func dataRecipientStatus(item app.DataRecipientStatusItem) dataRecipientStatusJSON {
	return dataRecipientStatusJSON{LegalEntityID: item.LegalEntityID, Status: string(item.Status)}
}

type softwareProductStatusJSON struct {
	SoftwareProductID string `json:"softwareProductId"`
	Status            string `json:"status"`
}

func softwareProductStatus(item app.SoftwareProductStatusItem) softwareProductStatusJSON {
	return softwareProductStatusJSON{SoftwareProductID: item.SoftwareProductID, Status: string(item.Status)}
}

// The JSON types of Get Data Recipients map each field explicitly, because the
// domain model also has SSA metadata that this API must not show.
type dataRecipientJSON struct {
	LegalEntityID       string                   `json:"legalEntityId"`
	LegalEntityName     string                   `json:"legalEntityName"`
	AccreditationNumber string                   `json:"accreditationNumber"`
	AccreditationLevel  string                   `json:"accreditationLevel"`
	LogoURI             string                   `json:"logoUri"`
	DataRecipientBrands []dataRecipientBrandJSON `json:"dataRecipientBrands"`
	Status              string                   `json:"status"`
	LastUpdated         string                   `json:"lastUpdated"`
}

type dataRecipientBrandJSON struct {
	DataRecipientBrandID string                `json:"dataRecipientBrandId"`
	BrandName            string                `json:"brandName"`
	LogoURI              string                `json:"logoUri"`
	SoftwareProducts     []softwareProductJSON `json:"softwareProducts"`
	Status               string                `json:"status"`
}

type softwareProductJSON struct {
	SoftwareProductID          string `json:"softwareProductId"`
	SoftwareProductName        string `json:"softwareProductName"`
	SoftwareProductDescription string `json:"softwareProductDescription"`
	LogoURI                    string `json:"logoUri"`
	Status                     string `json:"status"`
}

func dataRecipient(item app.DataRecipientItem) dataRecipientJSON {
	dr := item.DataRecipient

	brands := make([]dataRecipientBrandJSON, 0, len(dr.Brands))
	for _, b := range dr.Brands {
		products := make([]softwareProductJSON, 0, len(b.SoftwareProducts))
		for _, sp := range b.SoftwareProducts {
			products = append(products, softwareProductJSON{
				SoftwareProductID:          sp.SoftwareProductID,
				SoftwareProductName:        sp.SoftwareProductName,
				SoftwareProductDescription: sp.SoftwareProductDescription,
				LogoURI:                    sp.LogoURI,
				Status:                     string(sp.Status),
			})
		}

		brands = append(brands, dataRecipientBrandJSON{
			DataRecipientBrandID: b.DataRecipientBrandID,
			BrandName:            b.BrandName,
			LogoURI:              b.LogoURI,
			SoftwareProducts:     products,
			Status:               string(b.Status),
		})
	}

	return dataRecipientJSON{
		LegalEntityID:       dr.LegalEntityID,
		LegalEntityName:     dr.LegalEntityName,
		AccreditationNumber: dr.AccreditationNumber,
		AccreditationLevel:  string(dr.AccreditationLevel),
		LogoURI:             dr.LogoURI,
		DataRecipientBrands: brands,
		Status:              string(dr.Status),
		LastUpdated:         item.LastUpdated.UTC().Format(time.RFC3339),
	}
}

func recipientIndustry(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("industry") != "all" {
			cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, "industry")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// list serves a list that is not paginated.
func list[I, T any](baseURL string, read func(context.Context) ([]I, error), convert func(I) T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, err := read(r.Context())
		if err != nil {
			unexpected(w, r, err)
			return
		}

		data := make([]T, 0, len(items))
		for _, item := range items {
			data = append(data, convert(item))
		}

		writeCacheable(w, r, cds.NewResponse(r, baseURL, data))
	})
}
