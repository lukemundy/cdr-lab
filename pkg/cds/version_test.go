package cds_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

func requestVersions(t *testing.T, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.Header = header

	rec := httptest.NewRecorder()
	cds.Versioned(2, 3)(http.HandlerFunc(noContent)).ServeHTTP(rec, req)

	return rec
}

func TestVersionedServesTheHighestSupportedVersionOfTheRequest(t *testing.T) {
	tests := map[string]struct {
		header http.Header
		want   string
	}{
		"x-v is supported":                     {http.Header{"X-V": {"2"}}, "2"},
		"x-v has leading zeros":                {http.Header{"X-V": {"003"}}, "3"},
		"x-v is higher, x-min-v is supported":  {http.Header{"X-V": {"9"}, "X-Min-V": {"2"}}, "3"},
		"x-v is too large for an integer":      {http.Header{"X-V": {"99999999999999999999"}, "X-Min-V": {"1"}}, "3"},
		"x-min-v is higher than x-v":           {http.Header{"X-V": {"2"}, "X-Min-V": {"5"}}, "2"},
		"x-min-v is lower than each supported": {http.Header{"X-V": {"3"}, "X-Min-V": {"1"}}, "3"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rec := requestVersions(t, tc.header)

			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(t, tc.want, rec.Header().Get("x-v"))
		})
	}
}

func TestVersionedRefusesARequestWithNoSupportedVersion(t *testing.T) {
	tests := map[string]struct {
		header http.Header
		status int
		code   string
	}{
		"no x-v": {http.Header{}, http.StatusBadRequest, "urn:au-cds:error:cds-all:Header/Missing"},
		"x-v is zero": {
			http.Header{"X-V": {"0"}}, http.StatusBadRequest, "urn:au-cds:error:cds-all:Header/InvalidVersion",
		},
		"x-v has a sign": {
			http.Header{"X-V": {"+2"}}, http.StatusBadRequest, "urn:au-cds:error:cds-all:Header/InvalidVersion",
		},
		"x-min-v is not a number": {
			http.Header{"X-V": {"2"}, "X-Min-V": {"a"}},
			http.StatusBadRequest,
			"urn:au-cds:error:cds-all:Header/InvalidVersion",
		},
		"x-v is lower than each supported": {
			http.Header{"X-V": {"1"}}, http.StatusNotAcceptable, "urn:au-cds:error:cds-all:Header/UnsupportedVersion",
		},
		"x-min-v is higher than each supported": {
			http.Header{"X-V": {"9"}, "X-Min-V": {"4"}},
			http.StatusNotAcceptable,
			"urn:au-cds:error:cds-all:Header/UnsupportedVersion",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rec := requestVersions(t, tc.header)

			assert.Equal(t, tc.status, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.code)
			assert.Empty(t, rec.Header().Get("x-v"))
		})
	}
}
