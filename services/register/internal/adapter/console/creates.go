package console

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const (
	badCreateBody = "The request body must be a JSON object with only the fields of the record, " +
		"with no child records."
	badAddBody = "The request body must be a JSON object with only the version of the participant, " +
		"an integer of 1 or more, and the fields of the record, with no child records and no client certificates."
)

type newID struct{ field, value string }

// The body has no child records, so an ID that a different record has is always one of ids, and the first ID with a
// value is the record of the field errors.
type creation struct {
	ids      []newID
	location string
	execute  func(context.Context) error
}

func (c creation) recordID() string {
	for _, id := range c.ids {
		if id.value != "" {
			return id.value
		}
	}

	return ""
}

// fieldOf gives the field of the identifier that a different record has, or the first field when the error does not
// name the identifier.
func (c creation) fieldOf(notUnique domain.IDNotUniqueError) string {
	for _, id := range c.ids {
		if notUnique.ID != "" && id.value == notUnique.ID {
			return id.field
		}
	}

	return c.ids[0].field
}

func registerCreates(routes *httproute.Routes, operator Operator) {
	type dataRecipientCreate struct {
		LegalEntityID string `json:"legalEntityId"`
		domain.DataRecipientDetails
		Status domain.DataRecipientStatus `json:"status"`
	}

	type brandAdd struct {
		versioned
		DataRecipientBrandID string `json:"dataRecipientBrandId"`
		domain.DataRecipientBrandDetails
		Status domain.Status `json:"status"`
	}

	type dataHolderBrandAdd struct {
		versioned
		DataHolderBrandID string `json:"dataHolderBrandId"`
		InterimID         string `json:"interimId"`
		domain.DataHolderBrandDetails
		Status domain.Status `json:"status"`
	}

	type softwareProductAdd struct {
		versioned
		SoftwareProductID string `json:"softwareProductId"`
		domain.SoftwareProductDetails
		Status domain.Status `json:"status"`
	}

	routes.Handle("POST /api/data-recipients", executeCreate(func(c dataRecipientCreate) creation {
		return creation{
			ids:      []newID{{field: "legalEntityId", value: c.LegalEntityID}},
			location: "/api/data-recipients/" + url.PathEscape(c.LegalEntityID),
			execute: func(ctx context.Context) error {
				d := c.DataRecipientDetails

				return operator.Execute(
					ctx,
					c.LegalEntityID,
					domain.CreateDataRecipient{DataRecipient: domain.DataRecipient{
						LegalEntityID: c.LegalEntityID, LegalEntityName: d.LegalEntityName,
						AccreditationNumber: d.AccreditationNumber, AccreditationLevel: d.AccreditationLevel,
						LogoURI: d.LogoURI, Status: c.Status,
					}},
				)
			},
		}
	}))

	type dataHolderCreate struct {
		LegalEntityID string `json:"legalEntityId"`
		domain.DataHolderDetails
		Status domain.DataHolderStatus `json:"status"`
	}

	routes.Handle("POST /api/data-holders", executeCreate(func(c dataHolderCreate) creation {
		return creation{
			ids:      []newID{{field: "legalEntityId", value: c.LegalEntityID}},
			location: "/api/data-holders/" + url.PathEscape(c.LegalEntityID),
			execute: func(ctx context.Context) error {
				return operator.ExecuteDataHolder(ctx, c.LegalEntityID, domain.CreateDataHolder{
					DataHolder: domain.DataHolder{
						LegalEntityID: c.LegalEntityID, LegalEntityName: c.LegalEntityName, LogoURI: c.LogoURI,
						Status: c.Status,
					},
				})
			},
		}
	}))

	dataRecipients, dataHolders := dataRecipientsAt(operator), dataHoldersAt(operator)

	routes.Handle("POST /api/data-recipients/{legalEntityId}/brands", executeAdd(noDataRecipient,
		func(r *http.Request, a brandAdd, version int64) creation {
			return creation{
				ids: []newID{{field: "dataRecipientBrandId", value: a.DataRecipientBrandID}},
				location: "/api/data-recipients/" + url.PathEscape(r.PathValue("legalEntityId")) +
					"/brands/" + url.PathEscape(a.DataRecipientBrandID),
				execute: dataRecipients.addTo(r, version, domain.AddDataRecipientBrand{Brand: domain.DataRecipientBrand{
					DataRecipientBrandID: a.DataRecipientBrandID, BrandName: a.BrandName, LogoURI: a.LogoURI,
					Status: a.Status,
				}}),
			}
		}))

	routes.Handle("POST /api/data-recipients/{legalEntityId}/brands/{brandId}/software-products",
		executeAdd(noDataRecipientBrand, func(r *http.Request, a softwareProductAdd, version int64) creation {
			return creation{
				ids: []newID{{field: "softwareProductId", value: a.SoftwareProductID}},
				location: "/api/data-recipients/" + url.PathEscape(r.PathValue("legalEntityId")) +
					"/software-products/" + url.PathEscape(a.SoftwareProductID),
				execute: dataRecipients.addTo(r, version, domain.AddSoftwareProduct{
					BrandID:         r.PathValue("brandId"),
					SoftwareProduct: a.SoftwareProduct(a.SoftwareProductID, a.Status),
				}),
			}
		}))

	routes.Handle("POST /api/data-holders/{legalEntityId}/brands", executeAdd(noDataHolder,
		func(r *http.Request, a dataHolderBrandAdd, version int64) creation {
			b := domain.DataHolderBrand{
				DataHolderBrandID: a.DataHolderBrandID, InterimID: a.InterimID, BrandName: a.BrandName,
				BrandGroup: a.BrandGroup, Industries: a.Industries, LogoURI: a.LogoURI, Status: a.Status,
				EndpointDetail: a.EndpointDetail, JWKSEndpoint: a.JWKSEndpoint,
			}

			return creation{
				ids: []newID{
					{field: "dataHolderBrandId", value: a.DataHolderBrandID}, {field: "interimId", value: a.InterimID},
				},
				location: "/api/data-holders/" + url.PathEscape(r.PathValue("legalEntityId")) +
					"/brands/" + url.PathEscape(b.ID()),
				execute: dataHolders.addTo(r, version, domain.AddDataHolderBrand{Brand: b}),
			}
		}))
}

func (execute executeAt[Command]) addTo(r *http.Request, version int64, command Command) func(context.Context) error {
	return func(ctx context.Context) error {
		return execute(ctx, r.PathValue("legalEntityId"), version, command)
	}
}

func executeCreate[R any](create func(R) creation) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A pointer is nil after the JSON null, which decodes into a struct with no error.
		var request *R

		err := jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxChangeBodyBytes), &request)
		if err != nil || request == nil {
			problem.WriteBodyRefusal(w, err, badCreateBody)
			return
		}

		writeCreation(w, r, create(*request), noDataRecipient)
	})
}

func executeAdd[R changeRequest](notFound string, add func(*http.Request, R, int64) creation) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request R
		if err := jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxChangeBodyBytes), &request); err != nil {
			problem.WriteBodyRefusal(w, err, badAddBody)
			return
		}

		version, ok := request.version()
		if !ok {
			problem.Write(w, http.StatusBadRequest, badAddBody)
			return
		}

		writeCreation(w, r, add(r, request, version), notFound)
	})
}

func writeCreation(w http.ResponseWriter, r *http.Request, c creation, notFound string) {
	err := c.execute(r.Context())

	var (
		ruleErr   domain.StatusRuleError
		notUnique domain.IDNotUniqueError
	)

	switch {
	case err == nil:
		w.Header().Set("Location", c.location)
		w.WriteHeader(http.StatusCreated)
	case errors.Is(err, eventsourcing.ErrVersionConflict):
		problem.Write(w, http.StatusConflict, staleVersion)
	case errors.Is(err, domain.ErrNotFound):
		problem.Write(w, http.StatusNotFound, notFound)
	case errors.Is(err, domain.ErrExists), errors.As(err, &notUnique):
		field := c.fieldOf(notUnique)
		problem.WriteFieldErrors(w, http.StatusConflict, "", []problem.FieldError{{
			Pointer: problem.FieldPointer(field), Detail: field + " is not unique: a different record has it",
		}})
	case errors.Is(err, domain.ErrNotValid):
		writeRefusedValues(w, http.StatusBadRequest, c.recordID(), err)
	case errors.As(err, &ruleErr):
		writeRefusedValues(w, http.StatusConflict, c.recordID(), err)
	default:
		problem.WriteUnexpected(w, r, err, notCompleted)
	}
}
