package console

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const (
	maxChangeBodyBytes   = 1 << 20
	noDataRecipientBrand = "The Register has no data recipient with this legal entity ID, " +
		"or the data recipient has no brand with this ID."
	noDataHolderBrand = "The Register has no data holder with this legal entity ID, " +
		"or the data holder has no brand with this ID."
	badChangeBody = "The request body must be a JSON object with only the version of the participant, " +
		"an integer of 1 or more, and the fields of the change."
)

// A change request embeds versioned next to the fields of the change.
type versioned struct {
	Version *int64 `json:"version"`
}

func (v versioned) version() (int64, bool) {
	if v.Version == nil || *v.Version < 1 {
		return 0, false
	}

	return *v.Version, true
}

type changeRequest interface{ version() (int64, bool) }

type executeAt[Command any] func(ctx context.Context, legalEntityID string, version int64, command Command) error

// The function literals call the operator, because a method value of a nil operator panics.
func dataRecipientsAt(operator Operator) executeAt[domain.DataRecipientCommand] {
	return func(ctx context.Context, id string, version int64, c domain.DataRecipientCommand) error {
		return operator.ExecuteAt(ctx, id, version, c)
	}
}

func dataHoldersAt(operator Operator) executeAt[domain.DataHolderCommand] {
	return func(ctx context.Context, id string, version int64, c domain.DataHolderCommand) error {
		return operator.ExecuteDataHolderAt(ctx, id, version, c)
	}
}

func registerChanges(routes *httproute.Routes, operator Operator) {
	type detailsChange struct {
		versioned
		domain.DataRecipientDetails
	}

	type brandChange struct {
		versioned
		domain.DataRecipientBrandDetails
	}

	type softwareProductChange struct {
		versioned
		domain.SoftwareProductDetails
	}

	type dataHolderDetailsChange struct {
		versioned
		domain.DataHolderDetails
	}

	type dataHolderBrandChange struct {
		versioned
		domain.DataHolderBrandDetails
	}

	dataRecipients, dataHolders := dataRecipientsAt(operator), dataHoldersAt(operator)

	routes.Handle("POST /api/data-recipients/{legalEntityId}/change", executeChange(dataRecipients, noDataRecipient,
		func(r *http.Request, c detailsChange) (string, domain.DataRecipientCommand) {
			return r.PathValue("legalEntityId"), domain.ChangeDataRecipientDetails{Details: c.DataRecipientDetails}
		}))
	routes.Handle("POST /api/data-recipients/{legalEntityId}/brands/{brandId}/change",
		executeChange(dataRecipients, noDataRecipientBrand,
			func(r *http.Request, c brandChange) (string, domain.DataRecipientCommand) {
				id := r.PathValue("brandId")
				return id, domain.ChangeDataRecipientBrandDetails{BrandID: id, Details: c.DataRecipientBrandDetails}
			}))
	routes.Handle("POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/change",
		executeChange(dataRecipients, noSoftwareProduct,
			func(r *http.Request, c softwareProductChange) (string, domain.DataRecipientCommand) {
				id := r.PathValue("softwareProductId")
				return id, domain.ChangeSoftwareProductDetails{SoftwareProductID: id, Details: c.SoftwareProductDetails}
			}))
	routes.Handle("POST /api/data-holders/{legalEntityId}/change", executeChange(dataHolders, noDataHolder,
		func(r *http.Request, c dataHolderDetailsChange) (string, domain.DataHolderCommand) {
			return r.PathValue("legalEntityId"), domain.ChangeDataHolderDetails{Details: c.DataHolderDetails}
		}))
	routes.Handle("POST /api/data-holders/{legalEntityId}/brands/{brandId}/change",
		executeChange(dataHolders, noDataHolderBrand,
			func(r *http.Request, c dataHolderBrandChange) (string, domain.DataHolderCommand) {
				id := r.PathValue("brandId")
				return id, domain.ChangeDataHolderBrandDetails{BrandID: id, Details: c.DataHolderBrandDetails}
			}))
}

// The command function also gives the ID of the record whose fields are in the body, because only its field errors
// have a pointer into the body.
func executeChange[C changeRequest, Command any](
	execute executeAt[Command],
	notFound string,
	command func(*http.Request, C) (string, Command),
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request C
		if jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxChangeBodyBytes), &request) != nil {
			problem.Write(w, http.StatusBadRequest, badChangeBody)
			return
		}

		version, ok := request.version()
		if !ok {
			problem.Write(w, http.StatusBadRequest, badChangeBody)
			return
		}

		recordID, c := command(r, request)
		err := execute(r.Context(), r.PathValue("legalEntityId"), version, c)

		var ruleErr domain.StatusRuleError

		switch {
		case errors.Is(err, domain.ErrNotValid):
			writeRefusedValues(w, http.StatusBadRequest, recordID, err)
		case errors.As(err, &ruleErr):
			writeRefusedValues(w, http.StatusConflict, recordID, err)
		default:
			writeResult(w, r, notFound, err)
		}
	})
}

func writeRefusedValues(w http.ResponseWriter, status int, recordID string, err error) {
	fields, others := domain.FieldErrors(err)

	var errs []problem.FieldError

	for _, f := range fields {
		if f.RecordID != recordID {
			others = append(others, f)
			continue
		}

		errs = append(errs, problem.FieldError{Pointer: problem.FieldPointer(f.Field), Detail: f.Detail})
	}

	messages := make([]string, 0, len(others))
	for _, o := range others {
		messages = append(messages, o.Error())
	}

	problem.WriteFieldErrors(w, status, strings.Join(messages, "\n"), errs)
}
