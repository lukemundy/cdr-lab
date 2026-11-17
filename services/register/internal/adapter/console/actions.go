package console

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// Operator methods with a version give eventsourcing.ErrVersionConflict when the participant has a different version.
type Operator interface {
	Execute(ctx context.Context, legalEntityID string, command domain.DataRecipientCommand) error
	ExecuteAt(ctx context.Context, legalEntityID string, version int64, command domain.DataRecipientCommand) error
	ExecuteDataHolder(ctx context.Context, legalEntityID string, command domain.DataHolderCommand) error
	ExecuteDataHolderAt(
		ctx context.Context,
		legalEntityID string,
		version int64,
		command domain.DataHolderCommand,
	) error
}

const (
	maxActionBodyBytes = 1 << 10
	noSoftwareProduct  = "The Register has no data recipient with this legal entity ID, " +
		"or the data recipient has no software product with this ID."
	staleVersion = "The participant changed after the version in the request. The Register changed nothing."
	badBody      = "The request body must be a JSON object with only the version of the participant, " +
		"an integer of 1 or more."
)

func registerStatusActions(routes *httproute.Routes, operator Operator) {
	dataRecipients, dataHolders := dataRecipientsAt(operator), dataHoldersAt(operator)

	for _, a := range domain.DataRecipientActions {
		routes.Handle("POST /api/data-recipients/{legalEntityId}/"+string(a),
			executeAction(dataRecipients, noDataRecipient, func(*http.Request) domain.DataRecipientCommand {
				return a.Command()
			}))
	}

	for _, a := range domain.DataHolderActions {
		routes.Handle("POST /api/data-holders/{legalEntityId}/"+string(a),
			executeAction(dataHolders, noDataHolder, func(*http.Request) domain.DataHolderCommand {
				return a.Command()
			}))
	}

	for _, a := range domain.RecordActions {
		routes.Handle("POST /api/data-recipients/{legalEntityId}/brands/{brandId}/"+string(a),
			executeAction(dataRecipients, noDataRecipientBrand, func(r *http.Request) domain.DataRecipientCommand {
				return a.DataRecipientBrandCommand(r.PathValue("brandId"))
			}))
		routes.Handle("POST /api/data-recipients/{legalEntityId}/software-products/{softwareProductId}/"+string(a),
			executeAction(dataRecipients, noSoftwareProduct, func(r *http.Request) domain.DataRecipientCommand {
				return a.SoftwareProductCommand(r.PathValue("softwareProductId"))
			}))
		routes.Handle("POST /api/data-holders/{legalEntityId}/brands/{brandId}/"+string(a),
			executeAction(dataHolders, noDataHolderBrand, func(r *http.Request) domain.DataHolderCommand {
				return a.DataHolderBrandCommand(r.PathValue("brandId"))
			}))
	}
}

// The detail of a problem does not contain an ID, because the IDs are in the URL that the operator sent.
func executeAction[Command any](
	execute executeAt[Command],
	notFound string,
	command func(*http.Request) Command,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version, err := decodeVersion(http.MaxBytesReader(w, r.Body, maxActionBodyBytes))
		if err != nil {
			problem.WriteBodyRefusal(w, err, badBody)
			return
		}

		writeResult(w, r, notFound, execute(r.Context(), r.PathValue("legalEntityId"), version, command(r)))
	})
}

func writeResult(w http.ResponseWriter, r *http.Request, notFound string, err error) {
	var ruleErr domain.StatusRuleError

	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, eventsourcing.ErrVersionConflict):
		problem.Write(w, http.StatusConflict, staleVersion)
	case errors.As(err, &ruleErr):
		problem.Write(w, http.StatusConflict, ruleErr.Error())
	case errors.Is(err, domain.ErrNotFound):
		problem.Write(w, http.StatusNotFound, notFound)
	case errors.Is(err, domain.ErrNotValid):
		problem.Write(w, http.StatusBadRequest, thumbprintNotValid)
	default:
		problem.WriteUnexpected(w, r, err, notCompleted)
	}
}

var errFieldsNotValid = errors.New("the body does not have the fields of the action")

func decodeVersion(body io.Reader) (int64, error) {
	var request struct {
		Version *int64 `json:"version"`
	}

	if err := jsonhttp.Decode(body, &request); err != nil {
		return 0, err
	}

	if request.Version == nil || *request.Version < 1 {
		return 0, errFieldsNotValid
	}

	return *request.Version, nil
}
