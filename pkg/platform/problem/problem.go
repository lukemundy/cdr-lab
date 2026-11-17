// Package problem writes RFC 9457 problem details for the HTTP interfaces that the CDR does not define.
package problem

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

const (
	// The anchor is the identifier of the rule, so the URI stays when a specification moves.
	fieldErrorsType   = "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002"
	fieldErrorsTitle  = "Values in the request are not valid"
	fieldErrorsDetail = "One or more values in the request are not valid."
	sessionEndedType  = "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0010"
	sessionEndedTitle = "The session ended"
)

// Write uses the type about:blank, so that the title is the HTTP status phrase.
func Write(w http.ResponseWriter, status int, detail string) {
	writeTyped(w, "about:blank", http.StatusText(status), status, detail)
}

func writeTyped(w http.ResponseWriter, typeURI, title string, status int, detail string) {
	write(w, status, struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}{Type: typeURI, Title: title, Status: status, Detail: detail})
}

// WriteBodyRefusal writes 413 when err comes from a body that is larger than its limit, and 400 with detail for each
// other err, also nil.
func WriteBodyRefusal(w http.ResponseWriter, err error, detail string) {
	if tooLarge, ok := errors.AsType[*http.MaxBytesError](err); ok {
		Write(w, http.StatusRequestEntityTooLarge, "The request body is larger than "+formatSize(tooLarge.Limit)+".")
		return
	}

	Write(w, http.StatusBadRequest, detail)
}

func formatSize(limit int64) string {
	switch {
	case limit%(1<<20) == 0:
		return strconv.FormatInt(limit>>20, 10) + " MiB"
	case limit%(1<<10) == 0:
		return strconv.FormatInt(limit>>10, 10) + " KiB"
	default:
		return strconv.FormatInt(limit, 10) + " bytes"
	}
}

// WriteSessionEnded is for a request with a session cookie whose session ended, so that a UI can tell it apart from a
// request with no session.
func WriteSessionEnded(w http.ResponseWriter, status int, detail string) {
	writeTyped(w, sessionEndedType, sessionEndedTitle, status, detail)
}

// WriteUnexpected does not log or show the message of err, because database errors can contain query arguments.
func WriteUnexpected(w http.ResponseWriter, r *http.Request, err error, detail string) {
	slog.ErrorContext(r.Context(), "request failed", "error.type", fmt.Sprintf("%T", err))
	Write(w, http.StatusInternalServerError, detail)
}

type FieldError struct {
	Pointer string `json:"pointer"`
	Detail  string `json:"detail"`
}

// FieldPointer gives the JSON Pointer of a field name with a dot between the names of nested fields.
func FieldPointer(field string) string {
	return "#/" + strings.ReplaceAll(field, ".", "/")
}

// WriteFieldErrors writes the problem of Write when errs is empty. An empty detail gives a general detail.
func WriteFieldErrors(w http.ResponseWriter, status int, detail string, errs []FieldError) {
	if len(errs) == 0 {
		Write(w, status, detail)
		return
	}

	if detail == "" {
		detail = fieldErrorsDetail
	}

	write(w, status, struct {
		Type   string       `json:"type"`
		Title  string       `json:"title"`
		Status int          `json:"status"`
		Detail string       `json:"detail"`
		Errors []FieldError `json:"errors"`
	}{Type: fieldErrorsType, Title: fieldErrorsTitle, Status: status, Detail: detail, Errors: errs})
}

func write(w http.ResponseWriter, status int, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
