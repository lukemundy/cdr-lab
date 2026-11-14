package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// The list maps each field explicitly, because the product also has the components that only the detail shows.
type productJSON struct {
	ProductID       string     `json:"productId"`
	EffectiveFrom   *time.Time `json:"effectiveFrom,omitempty"`
	EffectiveTo     *time.Time `json:"effectiveTo,omitempty"`
	LastUpdated     time.Time  `json:"lastUpdated"`
	ProductCategory string     `json:"productCategory"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Brand           string     `json:"brand"`
	IsTailored      bool       `json:"isTailored"`
}

type productListJSON struct {
	Products []productJSON `json:"products"`
}

func productList(products []domain.Product) productListJSON {
	list := productListJSON{Products: make([]productJSON, 0, len(products))}
	for _, p := range products {
		list.Products = append(list.Products, productJSON{
			ProductID:       p.ProductID,
			EffectiveFrom:   p.EffectiveFrom,
			EffectiveTo:     p.EffectiveTo,
			LastUpdated:     p.LastUpdated,
			ProductCategory: string(p.ProductCategory),
			Name:            p.Name,
			Description:     p.Description,
			Brand:           p.Brand,
			IsTailored:      p.IsTailored,
		})
	}

	return list
}

func products(baseURL string, reader app.ProductReader, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, filters, ok := productQuery(w, r)
		if !ok {
			return
		}

		page, ok := cds.ParsePage(w, r)
		if !ok {
			return
		}

		query.At, query.Page, query.PageSize = now(), page.Number, page.Size

		found, total, err := reader.Products(r.Context(), query)
		if err != nil {
			unexpected(w, r, err)
			return
		}

		cds.WritePage(w, r, baseURL, page, total, filters, productList(found))
	})
}

// productQuery writes the error response and gives false when a filter is not valid.
func productQuery(w http.ResponseWriter, r *http.Request) (app.ProductQuery, url.Values, bool) {
	params := r.URL.Query()
	filters := url.Values{}

	for _, name := range []string{"effective", "updated-since", "brand", "product-category"} {
		if v := params.Get(name); v != "" {
			filters.Set(name, v)
		}
	}

	query := app.ProductQuery{Effective: app.EffectiveCurrent, Brand: filters.Get("brand")}

	if v := filters.Get("effective"); v != "" {
		effective, valid := app.ParseEffective(v)
		if !valid {
			cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, "effective")
			return app.ProductQuery{}, nil, false
		}

		query.Effective = effective
	}

	if v := filters.Get("updated-since"); v != "" {
		since, err := time.Parse(time.RFC3339, v)
		if err != nil {
			cds.WriteError(w, http.StatusBadRequest, cds.InvalidDateTime, "updated-since")
			return app.ProductQuery{}, nil, false
		}

		query.UpdatedSince = &since
	}

	if v := filters.Get("product-category"); v != "" {
		category, valid := domain.ParseProductCategory(v)
		if !valid {
			cds.WriteError(w, http.StatusBadRequest, cds.FieldInvalid, "product-category")
			return app.ProductQuery{}, nil, false
		}

		query.Category = category
	}

	return query, filters, true
}

func productDetail(baseURL string, reader app.ProductReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		productID := r.PathValue("productId")

		product, err := reader.Product(r.Context(), productID)
		if errors.Is(err, app.ErrProductNotFound) {
			cds.WriteError(w, http.StatusNotFound, cds.ResourceInvalid, productID)
			return
		}

		if err != nil {
			unexpected(w, r, err)
			return
		}

		jsonhttp.Write(w, http.StatusOK, cds.NewResponse(r, baseURL, product))
	})
}

func unexpected(w http.ResponseWriter, r *http.Request, err error) {
	cds.WriteUnexpected(w, r, err, "The data holder could not complete the request.")
}
