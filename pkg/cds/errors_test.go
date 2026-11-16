package cds_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

// Verifies: CDSL-0012, CDSL-0013, CDSL-0014
func TestWriteErrorWritesTheCDSErrorEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()

	cds.WriteError(rec, http.StatusNotFound, cds.ResourceNotFound, "The URL is not an endpoint.")

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"errors": [{
		"code": "urn:au-cds:error:cds-all:Resource/NotFound",
		"title": "Resource Not Found",
		"detail": "The URL is not an endpoint."
	}]}`, rec.Body.String())
}

// Verifies: CDSL-0012
func TestWriteErrorsWritesOneErrorForEachDetail(t *testing.T) {
	rec := httptest.NewRecorder()

	cds.WriteErrors(rec, http.StatusUnprocessableEntity, cds.InvalidBankingAccount, "account-1", "account-2")

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	invalid := `"code": "urn:au-cds:error:cds-banking:Authorisation/InvalidBankingAccount", ` +
		`"title": "Invalid Banking Account"`
	assert.JSONEq(t, `{"errors": [
		{`+invalid+`, "detail": "account-1"},
		{`+invalid+`, "detail": "account-2"}
	]}`, rec.Body.String())
}
