//go:build conformance

package register

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
	"github.com/lukemundy/cdr-lab/pkg/openapi"
)

const (
	brandsSummaryPath          = "/cdr-register/v1/all/data-holders/brands/summary"
	dataHoldersStatusPath      = "/cdr-register/v1/all/data-holders/status"
	dataRecipientsStatusPath   = "/cdr-register/v1/all/data-recipients/status"
	softwareProductsStatusPath = "/cdr-register/v1/all/data-recipients/brands/software-products/status"
	dataRecipientsPath         = "/cdr-register/v1/all/data-recipients"
)

// RFC 9110, section 8.8.3: entity-tag = [ weak ] opaque-tag.
var entityTag = regexp.MustCompile(`^(W/)?"[\x21\x23-\x7E\x80-\xFF]*"$`)

var statusEndpoints = map[string]struct {
	path     string
	version  string
	idField  string
	statuses []string
}{
	"data holders": {
		path:     dataHoldersStatusPath,
		version:  "2",
		idField:  "legalEntityId",
		statuses: []string{"ACTIVE", "REMOVED"},
	},
	"data recipients": {
		path:     dataRecipientsStatusPath,
		version:  "3",
		idField:  "legalEntityId",
		statuses: []string{"ACTIVE", "SUSPENDED", "REVOKED", "SURRENDERED"},
	},
	"software products": {
		path:     softwareProductsStatusPath,
		version:  "3",
		idField:  "softwareProductId",
		statuses: []string{"ACTIVE", "INACTIVE", "REMOVED"},
	},
}

// listEndpoints have the same general behaviour.
var listEndpoints = map[string]struct {
	path    string
	version string
}{
	"data holder brands summary": {brandsSummaryPath, "2"},
	"data holder statuses":       {dataHoldersStatusPath, "2"},
	"data recipient statuses":    {dataRecipientsStatusPath, "3"},
	"software product statuses":  {softwareProductsStatusPath, "3"},
	"data recipients":            {dataRecipientsPath, "4"},
}

func xv(version string) http.Header {
	return http.Header{"X-V": {version}}
}

type response struct {
	*http.Response
	body []byte
}

// call validates each response against the OpenAPI specification, except for the /unknown path.
func call(t *testing.T, path string, header http.Header) response {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, labtest.Env(t, "REGISTER_PUBLIC_URL")+path, nil)
	require.NoError(t, err)

	req.Header = header

	resp, err := labtest.Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	if path != "/unknown" {
		require.NoError(t, spec(t).ValidateResponse(resp))
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return response{Response: resp, body: body}
}

func spec(t *testing.T) *openapi.Spec {
	t.Helper()

	s, err := LoadSpec()
	require.NoError(t, err)

	return s
}

func v3() http.Header {
	return http.Header{"X-V": {"3"}}
}

func requireCDSError(t *testing.T, resp response, status int, code, title string) {
	t.Helper()
	require.Equal(t, status, resp.StatusCode, string(resp.body))

	var body struct {
		Errors []struct {
			Code  string `json:"code"`
			Title string `json:"title"`
		} `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &body))
	require.Len(t, body.Errors, 1)
	assert.Equal(t, code, body.Errors[0].Code)
	assert.Equal(t, title, body.Errors[0].Title)
}

// Verifies: REG-0338
func TestStatusListsHaveEntitiesInEachStatus(t *testing.T) {
	for name, e := range statusEndpoints {
		t.Run(name, func(t *testing.T) {
			resp := call(t, e.path, xv(e.version))

			require.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, e.version, resp.Header.Get("x-v"))

			var body struct {
				Data []map[string]string `json:"data"`
			}
			require.NoError(t, json.Unmarshal(resp.body, &body))

			seen := map[string]bool{}

			for _, item := range body.Data {
				assert.NotEmpty(t, item[e.idField])
				seen[item["status"]] = true
			}

			for _, status := range e.statuses {
				assert.True(t, seen[status], "no entity has status %s", status)
			}
		})
	}
}

// Verifies: REG-0316
func TestListsRejectAnIndustryThatIsNotValid(t *testing.T) {
	for name, e := range listEndpoints {
		invalid := []string{"foo", "Banking"}
		if strings.Contains(e.path, "/data-recipients") {
			invalid = append(invalid, "banking")
		}

		for _, industry := range invalid {
			t.Run(name+"/"+industry, func(t *testing.T) {
				resp := call(t, strings.Replace(e.path, "/all/", "/"+industry+"/", 1), xv(e.version))

				if e.path == brandsSummaryPath {
					requireCDSError(
						t,
						resp,
						http.StatusNotFound,
						"urn:au-cds:error:cds-register:Field/InvalidIndustry",
						"Invalid Industry",
					)
				} else {
					requireCDSError(
						t,
						resp,
						http.StatusBadRequest,
						"urn:au-cds:error:cds-all:Field/Invalid",
						"Invalid Field",
					)
				}
			})
		}
	}
}

func TestListsNegotiateTheVersion(t *testing.T) {
	tests := map[string]func(version string) http.Header{
		"x-v is the version":      xv,
		"x-v 99 with x-min-v 2":   func(string) http.Header { return http.Header{"X-V": {"99"}, "X-Min-V": {"2"}} },
		"x-min-v higher than x-v": func(v string) http.Header { return http.Header{"X-V": {v}, "X-Min-V": {"5"}} },
	}
	for name, header := range tests {
		for endpoint, e := range listEndpoints {
			t.Run(name+"/"+endpoint, func(t *testing.T) {
				resp := call(t, e.path, header(e.version))

				require.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Equal(t, e.version, resp.Header.Get("x-v"))
			})
		}
	}
}

func TestListsGiveVersionErrors(t *testing.T) {
	tests := map[string]struct {
		header http.Header
		status int
		code   string
		title  string
	}{
		"no x-v": {
			http.Header{}, http.StatusBadRequest, "urn:au-cds:error:cds-all:Header/Missing", "Missing Required Header",
		},
		"x-v not a positive integer": {
			http.Header{
				"X-V": {"abc"},
			},
			http.StatusBadRequest,
			"urn:au-cds:error:cds-all:Header/InvalidVersion",
			"Invalid Version",
		},
		"x-min-v not a positive integer": {
			http.Header{
				"X-V":     {"99"},
				"X-Min-V": {"0"},
			},
			http.StatusBadRequest,
			"urn:au-cds:error:cds-all:Header/InvalidVersion",
			"Invalid Version",
		},
		"x-v not supported": {
			http.Header{
				"X-V": {"99"},
			},
			http.StatusNotAcceptable,
			"urn:au-cds:error:cds-all:Header/UnsupportedVersion",
			"Unsupported Version",
		},
	}
	for name, tt := range tests {
		for endpoint, e := range listEndpoints {
			t.Run(name+"/"+endpoint, func(t *testing.T) {
				resp := call(t, e.path, tt.header)

				requireCDSError(t, resp, tt.status, tt.code, tt.title)
			})
		}
	}
}

// Verifies: REG-0326
func TestListsHaveAnETagThatIfNoneMatchUses(t *testing.T) {
	for name, e := range listEndpoints {
		t.Run(name, func(t *testing.T) {
			first := call(t, e.path, xv(e.version))
			etag := first.Header.Get("ETag")
			require.Regexp(t, entityTag, etag)
			assert.Equal(t, "no-cache", first.Header.Get("Cache-Control"))

			header := xv(e.version)
			header.Set("If-None-Match", etag)
			notModified := call(t, e.path, header)

			assert.Equal(t, http.StatusNotModified, notModified.StatusCode)
			assert.Empty(t, notModified.body)
			assert.Equal(t, etag, notModified.Header.Get("ETag"))
			assert.Equal(t, "no-cache", notModified.Header.Get("Cache-Control"))

			header.Set("If-None-Match", `"not-the-current-tag"`)
			assert.Equal(t, http.StatusOK, call(t, e.path, header).StatusCode)
		})
	}
}

// Verifies: REG-0344
func TestAPreflightRequestPermitsAnyOriginToSendTheCDSHeaders(t *testing.T) {
	for name, e := range listEndpoints {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodOptions,
				labtest.Env(t, "REGISTER_PUBLIC_URL")+e.path,
				nil,
			)
			require.NoError(t, err)
			req.Header.Set("Origin", "https://comparison.example")
			req.Header.Set("Access-Control-Request-Method", "GET")
			req.Header.Set("Access-Control-Request-Headers", "x-v,x-min-v,if-none-match,x-fapi-interaction-id")

			resp, err := labtest.Client(t).Do(req)
			require.NoError(t, err)

			_ = resp.Body.Close()

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET", resp.Header.Get("Access-Control-Allow-Methods"))
			assert.Equal(
				t,
				"x-v,x-min-v,if-none-match,x-fapi-interaction-id",
				resp.Header.Get("Access-Control-Allow-Headers"),
			)
		})
	}
}

func TestPublicResponsesPermitAnyOrigin(t *testing.T) {
	tests := map[string]struct {
		path   string
		header http.Header
	}{
		"success":        {dataRecipientsStatusPath, v3()},
		"version error":  {softwareProductsStatusPath, http.Header{}},
		"industry error": {"/cdr-register/v1/banking/data-recipients/status", v3()},
		"unknown path":   {"/unknown", v3()},
		"not modified":   {dataRecipientsStatusPath, http.Header{"X-V": {"3"}, "If-None-Match": {"*"}}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := call(t, tt.path, tt.header)

			assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "x-v, etag, x-fapi-interaction-id", resp.Header.Get("Access-Control-Expose-Headers"))
		})
	}
}

// Verifies: REG-0318
func TestAnUnknownPathIsNotFound(t *testing.T) {
	resp := call(t, "/unknown", v3())

	requireCDSError(t, resp, http.StatusNotFound, "urn:au-cds:error:cds-all:Resource/NotFound", "Resource Not Found")
}

// Verifies: REG-0314
func TestPublicResponsesGiveTheInteractionID(t *testing.T) {
	for name, header := range map[string]http.Header{"success": v3(), "error": {}} {
		t.Run(name, func(t *testing.T) {
			id := uuid.NewString()
			header.Set("x-fapi-interaction-id", id)

			resp := call(t, dataRecipientsStatusPath, header)

			assert.Equal(t, id, resp.Header.Get("x-fapi-interaction-id"))
		})
	}
}

func TestPublicResponsesGiveANewInteractionIDWhenTheRequestHasNone(t *testing.T) {
	resp := call(t, dataRecipientsStatusPath, v3())

	id, err := uuid.Parse(resp.Header.Get("x-fapi-interaction-id"))
	require.NoError(t, err)
	assert.Equal(t, uuid.RFC4122, id.Variant())
}
