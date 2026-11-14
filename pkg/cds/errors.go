package cds

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
)

// Error is a code of the CDS error list, with the title that the list gives for the code.
type Error struct {
	Code  string
	Title string
}

// The cds-all codes of the CDS error list, which apply to all participants.
var (
	Unexpected = Error{
		Code:  "urn:au-cds:error:cds-all:GeneralError/Unexpected",
		Title: "Unexpected Error Encountered",
	}
	Expected = Error{
		Code:  "urn:au-cds:error:cds-all:GeneralError/Expected",
		Title: "Expected Error Encountered",
	}
	FieldMissing = Error{
		Code:  "urn:au-cds:error:cds-all:Field/Missing",
		Title: "Missing Required Field",
	}
	FieldInvalid = Error{
		Code:  "urn:au-cds:error:cds-all:Field/Invalid",
		Title: "Invalid Field",
	}
	InvalidDateTime = Error{
		Code:  "urn:au-cds:error:cds-all:Field/InvalidDateTime",
		Title: "Invalid Date",
	}
	InvalidPageSize = Error{
		Code:  "urn:au-cds:error:cds-all:Field/InvalidPageSize",
		Title: "Invalid Page Size",
	}
	InvalidPage = Error{
		Code:  "urn:au-cds:error:cds-all:Field/InvalidPage",
		Title: "Invalid Page",
	}
	HeaderInvalid = Error{
		Code:  "urn:au-cds:error:cds-all:Header/Invalid",
		Title: "Invalid Header",
	}
	HeaderMissing = Error{
		Code:  "urn:au-cds:error:cds-all:Header/Missing",
		Title: "Missing Required Header",
	}
	InvalidVersion = Error{
		Code:  "urn:au-cds:error:cds-all:Header/InvalidVersion",
		Title: "Invalid Version",
	}
	UnsupportedVersion = Error{
		Code:  "urn:au-cds:error:cds-all:Header/UnsupportedVersion",
		Title: "Unsupported Version",
	}
	ADRStatusNotActive = Error{
		Code:  "urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive",
		Title: "ADR Status Is Not Active",
	}
	InvalidConsent = Error{
		Code:  "urn:au-cds:error:cds-all:Authorisation/InvalidConsent",
		Title: "Consent Is Invalid",
	}
	InvalidArrangement = Error{
		Code:  "urn:au-cds:error:cds-all:Authorisation/InvalidArrangement",
		Title: "Invalid Consent Arrangement",
	}
	InvalidBankingAccount = Error{
		Code:  "urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount",
		Title: "Invalid Banking Account",
	}
	ResourceInvalid = Error{
		Code:  "urn:au-cds:error:cds-all:Resource/Invalid",
		Title: "Invalid Resource",
	}
	ResourceNotFound = Error{
		Code:  "urn:au-cds:error:cds-all:Resource/NotFound",
		Title: "Resource Not Found",
	}
	ResourceNotImplemented = Error{
		Code:  "urn:au-cds:error:cds-all:Resource/NotImplemented",
		Title: "Resource Not Implemented",
	}
)

type errorListJSON struct {
	Errors []errorJSON `json:"errors"`
}

type errorJSON struct {
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func WriteError(w http.ResponseWriter, status int, e Error, detail string) {
	WriteErrors(w, status, e, detail)
}

func WriteErrors(w http.ResponseWriter, status int, e Error, details ...string) {
	list := errorListJSON{Errors: make([]errorJSON, 0, len(details))}
	for _, detail := range details {
		list.Errors = append(list.Errors, errorJSON{Code: e.Code, Title: e.Title, Detail: detail})
	}

	jsonhttp.Write(w, status, list)
}

// WriteUnexpected does not log or show the message of err, because database errors can contain query arguments.
func WriteUnexpected(w http.ResponseWriter, r *http.Request, err error, detail string) {
	slog.ErrorContext(r.Context(), "request failed", "error.type", fmt.Sprintf("%T", err))
	WriteError(w, http.StatusInternalServerError, Unexpected, detail)
}
