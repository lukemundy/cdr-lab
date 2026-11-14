package httpapi

import (
	"net/http"
	"net/url"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

type registeredBrandJSON struct {
	DataHolderBrandID string             `json:"dataHolderBrandId"`
	BrandName         string             `json:"brandName"`
	BrandGroup        string             `json:"brandGroup,omitempty"`
	Industries        []string           `json:"industries"`
	LogoURI           string             `json:"logoUri"`
	LegalEntity       legalEntityJSON    `json:"legalEntity"`
	Status            string             `json:"status"`
	EndpointDetail    endpointDetailJSON `json:"endpointDetail"`
	AuthDetails       []authDetailJSON   `json:"authDetails"`
	LastUpdated       string             `json:"lastUpdated"`
}

type legalEntityJSON struct {
	LegalEntityID   string `json:"legalEntityId"`
	LegalEntityName string `json:"legalEntityName"`
	LogoURI         string `json:"logoUri"`
	Status          string `json:"status"`
}

type endpointDetailJSON struct {
	Version          string `json:"version"`
	PublicBaseURI    string `json:"publicBaseUri"`
	ProductBaseURI   string `json:"productBaseUri,omitempty"`
	ResourceBaseURI  string `json:"resourceBaseUri"`
	InfosecBaseURI   string `json:"infosecBaseUri"`
	ExtensionBaseURI string `json:"extensionBaseUri,omitempty"`
	WebsiteURI       string `json:"websiteUri"`
}

type authDetailJSON struct {
	RegisterUType string `json:"registerUType"`
	JWKSEndpoint  string `json:"jwksEndpoint"`
}

func registeredBrand(item app.RegisteredDataHolderBrandItem) registeredBrandJSON {
	brand, endpoints := item.Brand, item.Brand.EndpointDetail

	return registeredBrandJSON{
		DataHolderBrandID: brand.DataHolderBrandID,
		BrandName:         brand.BrandName,
		BrandGroup:        brand.BrandGroup,
		Industries:        industryNames(brand.Industries),
		LogoURI:           brand.LogoURI,
		LegalEntity: legalEntityJSON{
			LegalEntityID:   item.LegalEntity.LegalEntityID,
			LegalEntityName: item.LegalEntity.LegalEntityName,
			LogoURI:         item.LegalEntity.LogoURI,
			Status:          string(item.LegalEntity.Status),
		},
		Status: string(brand.Status),
		EndpointDetail: endpointDetailJSON{
			Version:          endpoints.Version,
			PublicBaseURI:    endpoints.PublicBaseURI,
			ProductBaseURI:   endpoints.ProductBaseURI,
			ResourceBaseURI:  endpoints.ResourceBaseURI,
			InfosecBaseURI:   endpoints.InfosecBaseURI,
			ExtensionBaseURI: endpoints.ExtensionBaseURI,
			WebsiteURI:       endpoints.WebsiteURI,
		},
		AuthDetails: []authDetailJSON{{RegisterUType: "SIGNED-JWT", JWKSEndpoint: brand.JWKSEndpoint}},
		LastUpdated: item.LastUpdated.UTC().Format(time.RFC3339),
	}
}

func dataHolderBrands(baseURL string, brands app.RegisteredDataHolderBrandReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		industry, ok := holderIndustry(r)
		if !ok {
			cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, "industry")
			return
		}

		page, ok := cds.ParsePage(w, r)
		if !ok {
			return
		}

		query := app.DataHolderBrandQuery{Industry: industry, Page: page.Number, PageSize: page.Size}
		filters := url.Values{}

		if v := r.URL.Query().Get("updated-since"); v != "" {
			since, err := time.Parse(time.RFC3339, v)
			if err != nil {
				cds.WriteError(w, http.StatusBadRequest, cds.InvalidDateTime, "updated-since")
				return
			}

			query.UpdatedSince = &since

			filters.Set("updated-since", v)
		}

		items, total, err := brands.RegisteredDataHolderBrands(r.Context(), query)
		if err != nil {
			unexpected(w, r, err)
			return
		}

		data := make([]registeredBrandJSON, 0, len(items))
		for _, item := range items {
			data = append(data, registeredBrand(item))
		}

		cds.WritePage(w, r, baseURL, page, total, filters, data)
	})
}
