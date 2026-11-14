package operatorapi

import (
	"net/http"
	"net/url"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func registerDataRecipientChanges(routes *httproute.Routes, operator Operator) {
	const (
		dataRecipient = "/v1/data-recipients/{legalEntityId}"
		brand         = dataRecipient + "/brands/{brandId}"
	)

	dataRecipients := executor[domain.DataRecipientCommand](operator.Execute)

	routes.Handle("POST /v1/data-recipients", create(dataRecipients, "data recipient", func(dr domain.DataRecipient) (
		string, domain.DataRecipientCommand,
	) {
		return dr.LegalEntityID, domain.CreateDataRecipient{DataRecipient: dr}
	}))
	routes.Handle("PUT "+dataRecipient, executeBody(dataRecipients, "data recipient", func(
		_ *http.Request,
		d domain.DataRecipientDetails,
	) (domain.DataRecipientCommand, func(http.ResponseWriter)) {
		return domain.ChangeDataRecipientDetails{Details: d}, noContent
	}))
	routes.Handle("POST "+dataRecipient+"/brands", executeBody(dataRecipients, "data recipient", func(
		r *http.Request,
		b domain.DataRecipientBrand,
	) (domain.DataRecipientCommand, func(http.ResponseWriter)) {
		return domain.AddDataRecipientBrand{
			Brand: b,
		}, created(
			r.URL.Path + "/" + url.PathEscape(b.DataRecipientBrandID),
		)
	}))
	routes.Handle("PUT "+brand, executeBody(dataRecipients, "data recipient", func(
		r *http.Request,
		d domain.DataRecipientBrandDetails,
	) (domain.DataRecipientCommand, func(http.ResponseWriter)) {
		return domain.ChangeDataRecipientBrandDetails{BrandID: r.PathValue("brandId"), Details: d}, noContent
	}))

	for _, a := range domain.RecordActions {
		routes.Handle(
			"POST "+brand+"/"+string(a),
			executeCommandOf(dataRecipients, "data recipient", func(r *http.Request) domain.DataRecipientCommand {
				return a.DataRecipientBrandCommand(r.PathValue("brandId"))
			}),
		)
	}

	routes.Handle("POST "+brand+"/software-products", executeBody(dataRecipients, "data recipient", func(
		r *http.Request,
		sp domain.SoftwareProduct,
	) (domain.DataRecipientCommand, func(http.ResponseWriter)) {
		location := "/v1/data-recipients/" + url.PathEscape(r.PathValue("legalEntityId")) +
			"/software-products/" + url.PathEscape(sp.SoftwareProductID)

		return domain.AddSoftwareProduct{BrandID: r.PathValue("brandId"), SoftwareProduct: sp}, created(location)
	}))
	routes.Handle("PUT "+dataRecipient+"/software-products/{softwareProductId}", executeBody(
		dataRecipients,
		"data recipient",
		func(
			r *http.Request,
			d domain.SoftwareProductDetails,
		) (domain.DataRecipientCommand, func(http.ResponseWriter)) {
			return domain.ChangeSoftwareProductDetails{
				SoftwareProductID: r.PathValue("softwareProductId"), Details: d,
			}, noContent
		},
	))
}
