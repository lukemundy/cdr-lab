//go:build conformance

package dataholder

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	statusPath  = "/cds-au/v1/discovery/status"
	outagesPath = "/cds-au/v1/discovery/outages"
)

var discoveryPaths = []string{statusPath, outagesPath}

func xv(version string) http.Header { return http.Header{"X-V": {version}} }

// discovery validates each response against the pinned CDS common OpenAPI specification.
func discovery(t *testing.T, path string, header http.Header) response {
	t.Helper()

	resp := get(t, publicBaseURI(t)+path, header)

	spec, err := LoadCommonSpec()
	require.NoError(t, err)

	resp.Body = io.NopCloser(bytes.NewReader(resp.body))
	require.NoError(t, spec.ValidateResponse(resp.Response))

	return resp
}

type envelope[T any] struct {
	Data  T `json:"data"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
	Meta map[string]any `json:"meta"`
}

func TestGetStatusGivesOKWithTheTimeOfTheRequest(t *testing.T) {
	before := time.Now().Add(-time.Minute)
	resp := discovery(t, statusPath, xv("1"))

	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
	assert.Equal(t, "1", resp.Header.Get("x-v"))

	var body envelope[struct {
		Status     string    `json:"status"`
		UpdateTime time.Time `json:"updateTime"`
	}]
	require.NoError(t, json.Unmarshal(resp.body, &body))
	assert.Equal(t, "OK", body.Data.Status)
	assert.WithinRange(t, body.Data.UpdateTime, before, time.Now().Add(time.Minute))
	assert.Equal(t, publicBaseURI(t)+statusPath, body.Links.Self)
	assert.Empty(t, body.Meta)
}

func TestGetOutagesGivesNoOutages(t *testing.T) {
	resp := discovery(t, outagesPath, xv("1"))

	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
	assert.Equal(t, "1", resp.Header.Get("x-v"))

	var body envelope[struct {
		Outages []json.RawMessage `json:"outages"`
	}]
	require.NoError(t, json.Unmarshal(resp.body, &body))
	assert.NotNil(t, body.Data.Outages, "outages is an array, not null")
	assert.Empty(t, body.Data.Outages)
	assert.Equal(t, publicBaseURI(t)+outagesPath, body.Links.Self)
	assert.Empty(t, body.Meta)
}

func TestTheDiscoveryAPIsServeVersion1(t *testing.T) {
	tests := map[string]http.Header{
		"x-v 1":                            xv("1"),
		"x-v higher, x-min-v 1":            {"X-V": {"7"}, "X-Min-V": {"1"}},
		"x-min-v higher than x-v, ignored": {"X-V": {"1"}, "X-Min-V": {"3"}},
		"x-min-v equal to x-v, ignored":    {"X-V": {"1"}, "X-Min-V": {"1"}},
	}
	for _, path := range discoveryPaths {
		for name, header := range tests {
			t.Run(path+" "+name, func(t *testing.T) {
				resp := discovery(t, path, header)

				assert.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
				assert.Equal(t, "1", resp.Header.Get("x-v"))
			})
		}
	}
}

func TestTheDiscoveryAPIsRefuseARequestWithNoSupportedVersion(t *testing.T) {
	tests := map[string]struct {
		header http.Header
		status int
		code   string
		title  string
	}{
		"no x-v": {
			http.Header{},
			http.StatusBadRequest,
			"urn:au-cds:error:cds-all:Header/Missing", "Missing Required Header",
		},
		"x-v is not a positive integer": {
			xv("0"), http.StatusBadRequest,
			"urn:au-cds:error:cds-all:Header/InvalidVersion", "Invalid Version",
		},
		"x-min-v is not a positive integer": {
			http.Header{"X-V": {"1"}, "X-Min-V": {"a"}},
			http.StatusBadRequest,
			"urn:au-cds:error:cds-all:Header/InvalidVersion", "Invalid Version",
		},
		"x-v higher than version 1, no x-min-v": {
			xv("2"), http.StatusNotAcceptable,
			"urn:au-cds:error:cds-all:Header/UnsupportedVersion", "Unsupported Version",
		},
		"x-min-v higher than version 1": {
			http.Header{"X-V": {"3"}, "X-Min-V": {"2"}},
			http.StatusNotAcceptable,
			"urn:au-cds:error:cds-all:Header/UnsupportedVersion", "Unsupported Version",
		},
	}
	for _, path := range discoveryPaths {
		for name, tc := range tests {
			t.Run(path+" "+name, func(t *testing.T) {
				resp := discovery(t, path, tc.header)

				require.Equal(t, tc.status, resp.StatusCode, string(resp.body))

				var body struct {
					Errors []struct {
						Code   string `json:"code"`
						Title  string `json:"title"`
						Detail string `json:"detail"`
					} `json:"errors"`
				}
				require.NoError(t, json.Unmarshal(resp.body, &body))
				require.Len(t, body.Errors, 1)
				assert.Equal(t, tc.code, body.Errors[0].Code)
				assert.Equal(t, tc.title, body.Errors[0].Title)
				assert.NotEmpty(t, body.Errors[0].Detail)
			})
		}
	}
}

func TestTheDiscoveryAPIsAllowAnyOriginAndExposeTheCDSHeaders(t *testing.T) {
	tests := map[string]http.Header{
		"success": {"X-V": {"1"}, "Origin": {"https://app.example"}},
		"error":   {"Origin": {"https://app.example"}},
	}
	for _, path := range discoveryPaths {
		for name, header := range tests {
			t.Run(path+" "+name, func(t *testing.T) {
				resp := discovery(t, path, header)

				assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
				assert.Equal(
					t,
					"x-v, Retry-After, x-fapi-interaction-id",
					resp.Header.Get("Access-Control-Expose-Headers"),
				)
			})
		}
	}
}

func TestACORSPreflightRequestToADiscoveryAPIIsAllowed(t *testing.T) {
	for _, path := range discoveryPaths {
		t.Run(path, func(t *testing.T) {
			resp := send(t, http.MethodOptions, publicBaseURI(t)+path, http.Header{
				"Origin":                         {"https://app.example"},
				"Access-Control-Request-Method":  {"GET"},
				"Access-Control-Request-Headers": {"x-v,x-min-v,x-fapi-interaction-id"},
			})

			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET", resp.Header.Get("Access-Control-Allow-Methods"))
			assert.Equal(t, "x-v,x-min-v,x-fapi-interaction-id", resp.Header.Get("Access-Control-Allow-Headers"))
		})
	}
}
