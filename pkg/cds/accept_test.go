package cds_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withAccept(accept ...string) http.Header {
	header := http.Header{"X-V": {"2"}}
	if len(accept) > 0 {
		header["Accept"] = accept
	}

	return header
}

// Verifies: CDSL-0071, CDSL-0072, CDSL-0073, CDSL-0079
func TestVersionedServesARequestThatAcceptsJSONInUTF8(t *testing.T) {
	tests := map[string]http.Header{
		"no Accept":                          withAccept(),
		"an empty Accept":                    withAccept(""),
		"application/json":                   withAccept("application/json"),
		"application/json in upper case":     withAccept("APPLICATION/JSON"),
		"application/json with UTF-8":        withAccept("application/json;charset=UTF-8"),
		"application/json with utf-8":        withAccept("application/json; charset=utf-8"),
		"any media type":                     withAccept("*/*"),
		"any application media type":         withAccept("application/*"),
		"a list with JSON":                   withAccept("application/xml, application/json;q=0.5"),
		"two header lines, one with JSON":    withAccept("text/html", "application/json"),
		"a wildcard after a refused type":    withAccept("application/xml;q=1, */*;q=0.1"),
		"a media range with other parameter": withAccept("application/json; version=2"),
	}
	for name, header := range tests {
		t.Run(name, func(t *testing.T) {
			rec := requestVersions(t, header)

			assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		})
	}
}

// Verifies: CDSL-0071, CDSL-0072, CDSL-0073, CDSL-0074
func TestVersionedRefusesARequestThatDoesNotAcceptJSONInUTF8(t *testing.T) {
	tests := map[string]http.Header{
		"application/xml":                   withAccept("application/xml"),
		"text/html":                         withAccept("text/html"),
		"text/*":                            withAccept("text/*"),
		"a different character set":         withAccept("application/json;charset=iso-8859-1"),
		"any media type in a different set": withAccept("*/*; charset=us-ascii"),
		"JSON with q=0":                     withAccept("application/json;q=0"),
		"any media type with q=0.000":       withAccept("*/*;q=0.000, text/html"),
		"a malformed media range":           withAccept("json"),
		"a malformed quality value":         withAccept("application/json;q=high"),
	}
	for name, header := range tests {
		t.Run(name, func(t *testing.T) {
			rec := requestVersions(t, header)

			require.Equal(t, http.StatusNotAcceptable, rec.Code)

			var body struct {
				Errors []struct {
					Code   string `json:"code"`
					Title  string `json:"title"`
					Detail string `json:"detail"`
				} `json:"errors"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Len(t, body.Errors, 1)
			assert.Equal(t, "urn:au-cds:error:cds-all:Header/Invalid", body.Errors[0].Code)
			assert.Equal(t, "Invalid Header", body.Errors[0].Title)
			assert.Equal(t, "Accept", body.Errors[0].Detail)
			assert.Empty(t, rec.Header().Get("x-v"))
		})
	}
}

// Verifies: CDSL-0075
func TestVersionedChecksTheAcceptHeaderBeforeTheVersionHeaders(t *testing.T) {
	rec := requestVersions(t, http.Header{"Accept": {"application/xml"}})

	assert.Equal(t, http.StatusNotAcceptable, rec.Code)
	assert.Contains(t, rec.Body.String(), "urn:au-cds:error:cds-all:Header/Invalid")
}
