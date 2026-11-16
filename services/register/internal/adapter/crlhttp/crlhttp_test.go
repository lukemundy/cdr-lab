package crlhttp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/crlhttp"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

type source struct {
	crl app.CRL
	err error
}

func (s source) CurrentCRL(context.Context) (app.CRL, error) { return s.crl, s.err }

const crlPath = "/crl/cdr-intermediate.crl"

func serve(t *testing.T, s source, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	crlhttp.Register(mux, s)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, path, nil))

	return rec
}

var current = source{crl: app.CRL{Number: 3, DER: []byte{0x30, 0x03, 0x01, 0x02, 0x03}}}

// Verifies: REG-0718
func TestGetGivesTheCurrentCRLInDER(t *testing.T) {
	rec := serve(t, current, http.MethodGet, crlPath)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/pkix-crl", rec.Header().Get("Content-Type"))
	assert.Equal(t, "5", rec.Header().Get("Content-Length"))
	assert.Equal(t, current.crl.DER, rec.Body.Bytes())
}

// Verifies: REG-0719
func TestHeadGivesTheHeadersOfGetWithNoBody(t *testing.T) {
	rec := serve(t, current, http.MethodHead, crlPath)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/pkix-crl", rec.Header().Get("Content-Type"))
	assert.Equal(t, "5", rec.Header().Get("Content-Length"))
	assert.Empty(t, rec.Body.Bytes())
}

// Verifies: REG-0720
func TestAnotherMethodGives405WithTheAllowedMethods(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		rec := serve(t, current, method, crlPath)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, method)
		assert.Equal(t, "GET, HEAD", rec.Header().Get("Allow"), method)
	}
}

// Verifies: REG-0721
func TestAnotherPathGives404(t *testing.T) {
	for _, path := range []string{"/", "/crl/", "/crl/other.crl", crlPath + "/x"} {
		assert.Equal(t, http.StatusNotFound, serve(t, current, http.MethodGet, path).Code, path)
	}
}

// Verifies: REG-0722
func TestNoCurrentCRLGives503(t *testing.T) {
	for _, err := range []error{app.ErrNoCurrentCRL, errors.New("database failure")} {
		rec := serve(t, source{err: err}, http.MethodGet, crlPath)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.NotContains(t, rec.Body.String(), "database")
	}
}
