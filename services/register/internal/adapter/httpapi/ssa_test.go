package httpapi_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

const ssaPath = "/cdr-register/v1/all/data-recipients/brands/drb-1/software-products/sp-1/ssa"

type fakeSSAIssuer struct {
	ssa            string
	err            error
	product, brand string
	calls          int
}

func (f *fakeSSAIssuer) Issue(_ context.Context, softwareProductID, dataRecipientBrandID string) (string, error) {
	f.calls++
	f.product, f.brand = softwareProductID, dataRecipientBrandID

	return f.ssa, f.err
}

func authorisedV4() http.Header {
	return http.Header{"Authorization": {"Bearer valid"}, "X-V": {"4"}}
}

// Verifies: REG-0414
func TestGetSSAGivesTheSSAOfTheCallerAsACompactJWT(t *testing.T) {
	l := newSecureLab()

	rec := l.get(t, ssaPath, authorisedV4())

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "header.claims.signature", rec.Body.String(), "the body is the JWT with no JSON quotes")
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "4", rec.Header().Get("x-v"))
	assert.Empty(t, rec.Header().Get("ETag"))
	assert.Equal(t, callerID, l.ssa.product)
	assert.Equal(t, "drb-1", l.ssa.brand)
}

// Verifies: REG-0403, REG-0404
func TestGetSSARefusesAPathThatIsNotOfTheCaller(t *testing.T) {
	tests := map[string]struct {
		path                string
		issueErr            error
		code, title, detail string
		issued              int
	}{
		"another software product": {
			"/cdr-register/v1/all/data-recipients/brands/drb-1/software-products/sp-other/ssa", nil,
			"urn:au-cds:error:cds-register:Field/InvalidSoftwareProduct", "Invalid Software Product", "sp-other", 0,
		},
		"another brand": {
			"/cdr-register/v1/all/data-recipients/brands/drb-other/software-products/sp-1/ssa",
			app.ErrNotBrandOfSoftwareProduct,
			"urn:au-cds:error:cds-register:Field/InvalidBrand",
			"Invalid Brand",
			"drb-other",
			1,
		},
		"another software product and another brand": {
			"/cdr-register/v1/all/data-recipients/brands/drb-other/software-products/sp-other/ssa", nil,
			"urn:au-cds:error:cds-register:Field/InvalidSoftwareProduct", "Invalid Software Product", "sp-other", 0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newSecureLab()
			l.ssa.err = tt.issueErr

			rec := l.get(t, tt.path, authorisedV4())

			detail := requireSecureError(t, rec, http.StatusNotFound, tt.code, tt.title)
			assert.Equal(t, tt.detail, detail)
			assert.Equal(t, tt.issued, l.ssa.calls)
		})
	}
}

func TestGetSSAIssueErrors(t *testing.T) {
	tests := map[string]struct {
		err         error
		status      int
		code, title string
	}{
		"a status change after the caller check": {
			err:    fmt.Errorf("software product %w", domain.ErrNotActive),
			status: http.StatusForbidden,
			code:   "urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive",
			title:  "ADR Status Is Not Active",
		},
		"a failure": {
			err:    errors.New("connection to 10.0.0.1 failed"),
			status: http.StatusInternalServerError,
			code:   "urn:au-cds:error:cds-all:GeneralError/Unexpected",
			title:  "Unexpected Error Encountered",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newSecureLab()
			l.ssa.err = tt.err

			rec := l.get(t, ssaPath, authorisedV4())

			detail := requireSecureError(t, rec, tt.status, tt.code, tt.title)
			assert.NotContains(t, detail, "10.0.0.1")
		})
	}
}

// Verifies: REG-0402, REG-0323
func TestGetSSARefusesACallerThatIsNotActive(t *testing.T) {
	l := newSecureLab()
	l.reader.caller.Status = domain.Inactive

	rec := l.get(t, ssaPath, authorisedV4())

	assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	assert.Zero(t, l.ssa.calls)
}

func TestGetSSAChecksTheTokenTheVersionAndTheIndustry(t *testing.T) {
	tests := map[string]struct {
		path   string
		header http.Header
		status int
		issued int
	}{
		"no token": {path: ssaPath, header: http.Header{"X-V": {"4"}}, status: http.StatusUnauthorized},
		"no x-v": {
			path:   ssaPath,
			header: http.Header{"Authorization": {"Bearer valid"}},
			status: http.StatusBadRequest,
		},
		"x-v 3": {
			path:   ssaPath,
			header: http.Header{"Authorization": {"Bearer valid"}, "X-V": {"3"}},
			status: http.StatusNotAcceptable,
		},
		"x-v 5 with x-min-v 4": {
			path:   ssaPath,
			header: http.Header{"Authorization": {"Bearer valid"}, "X-V": {"5"}, "X-Min-V": {"4"}},
			status: http.StatusOK,
			issued: 1,
		},
		"industry banking": {
			path:   "/cdr-register/v1/banking/data-recipients/brands/drb-1/software-products/sp-1/ssa",
			header: authorisedV4(), status: http.StatusBadRequest,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newSecureLab()

			rec := l.get(t, tt.path, tt.header)

			assert.Equal(t, tt.status, rec.Code, rec.Body.String())
			assert.Equal(t, tt.issued, l.ssa.calls)
		})
	}
}

func jwksRequest(t *testing.T, keys httpapi.KeySource, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	httpapi.Register(mux, newReader(), keys, baseURL)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/cdr-register/v1/jwks", nil)
	req.Header = header
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

// Verifies: REG-0401, REG-0326
func TestGetJWKSGivesThePublicSSAKeysWithAnETag(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keys := func(context.Context) ([]authserver.SigningKey, error) {
		return []authserver.SigningKey{{ID: "ssa-1", Key: key}}, nil
	}

	rec := jwksRequest(t, keys, http.Header{})
	again := jwksRequest(t, keys, http.Header{"If-None-Match": {rec.Header().Get("ETag")}})

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Regexp(t, `^"[0-9a-f]+"$`, rec.Header().Get("ETag"))

	var body struct {
		Keys []map[string]any `json:"keys"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Keys, 1)
	assert.Equal(t, "ssa-1", body.Keys[0]["kid"])
	assert.Equal(t, "RSA", body.Keys[0]["kty"])
	assert.Equal(t, "PS256", body.Keys[0]["alg"])
	assert.Equal(t, []any{"verify"}, body.Keys[0]["key_ops"])
	assert.NotEmpty(t, body.Keys[0]["n"])
	assert.NotEmpty(t, body.Keys[0]["e"])
	assert.NotContains(t, body.Keys[0], "d", "the JWKS has no private key")
	assert.Equal(t, http.StatusNotModified, again.Code)
	assert.Empty(t, again.Body.String())
}

func TestGetJWKSKeyFailureIsAnUnexpectedError(t *testing.T) {
	rec := jwksRequest(t, func(context.Context) ([]authserver.SigningKey, error) {
		return nil, errors.New("connection to 10.0.0.1 failed")
	}, http.Header{})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "10.0.0.1")
}
