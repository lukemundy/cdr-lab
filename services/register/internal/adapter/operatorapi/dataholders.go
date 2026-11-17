package operatorapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func registerDataHolders(routes *httproute.Routes, operator Operator) {
	const (
		dataHolder = "/v1/data-holders/{legalEntityId}"
		brand      = dataHolder + "/brands/{brandId}"
	)

	dataHolders := executor[domain.DataHolderCommand](operator.ExecuteDataHolder)

	routes.Handle("POST /v1/data-holders", create(dataHolders, "data holder", func(dh domain.DataHolder) (
		string, domain.DataHolderCommand,
	) {
		return dh.LegalEntityID, domain.CreateDataHolder{DataHolder: dh}
	}))
	routes.Handle(
		"PUT "+dataHolder,
		executeBody(dataHolders, "data holder", func(r *http.Request, d domain.DataHolderDetails) (
			domain.DataHolderCommand, func(http.ResponseWriter),
		) {
			return domain.ChangeDataHolderDetails{Details: d}, noContent
		}),
	)
	routes.Handle(
		"POST "+dataHolder+"/remove",
		executeCommandOf(dataHolders, "data holder", func(*http.Request) domain.DataHolderCommand {
			return domain.RemoveDataHolder{}
		}),
	)
	routes.Handle("POST "+dataHolder+"/brands", executeBody(dataHolders, "data holder", func(
		r *http.Request,
		b domain.DataHolderBrand,
	) (domain.DataHolderCommand, func(http.ResponseWriter)) {
		location := r.URL.Path + "/" + url.PathEscape(b.ID())
		return domain.AddDataHolderBrand{Brand: b}, created(location)
	}))
	routes.Handle(
		"PUT "+brand,
		executeBody(dataHolders, "data holder", func(r *http.Request, d domain.DataHolderBrandDetails) (
			domain.DataHolderCommand, func(http.ResponseWriter),
		) {
			return domain.ChangeDataHolderBrandDetails{BrandID: r.PathValue("brandId"), Details: d}, noContent
		}),
	)

	for _, a := range domain.RecordActions {
		routes.Handle(
			"POST "+brand+"/"+string(a),
			executeCommandOf(dataHolders, "data holder", func(r *http.Request) domain.DataHolderCommand {
				return a.DataHolderBrandCommand(r.PathValue("brandId"))
			}),
		)
	}
}

// create executes the command on the legal entity ID of the body, and gives the path of the new legal entity.
func create[B, C any](execute executor[C], participant string, command func(B) (string, C)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body B
		if !decodeBody(w, r, &body) {
			return
		}

		legalEntityID, c := command(body)
		err := execute(r.Context(), legalEntityID, c)
		writeResult(w, r, err, participant, created(r.URL.Path+"/"+url.PathEscape(legalEntityID)))
	})
}

func created(location string) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.Header().Set("Location", location)
		w.WriteHeader(http.StatusCreated)
	}
}

type executor[C any] func(ctx context.Context, legalEntityID string, command C) error

// participant names the aggregate in the detail of a version conflict.
func executeCommandOf[C any](execute executor[C], participant string, command func(*http.Request) C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := execute(r.Context(), r.PathValue("legalEntityId"), command(r))
		writeResult(w, r, err, participant, noContent)
	})
}

func executeBody[B, C any](
	execute executor[C],
	participant string,
	command func(*http.Request, B) (C, func(http.ResponseWriter)),
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body B
		if !decodeBody(w, r, &body) {
			return
		}

		c, success := command(r, body)
		err := execute(r.Context(), r.PathValue("legalEntityId"), c)
		writeResult(w, r, err, participant, success)
	})
}

// maxBodyBytes is much more than a legal entity in the seed data, which is less than 20 KiB.
const maxBodyBytes = 1 << 20

// decodeBody writes the problem when the body is not one JSON object with only the fields of v.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	err := jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxBodyBytes), v)
	if err == nil {
		return true
	}

	problem.WriteBodyRefusal(w, err, fmt.Sprintf("The request body is not valid JSON for the action: %s.", err))

	return false
}
