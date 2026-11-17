package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

const baseURL = "https://bank.example"

// The zone is not UTC, to prove that the responses give times in UTC.
var requestTime = time.Date(2026, 10, 8, 21, 30, 15, 0, time.FixedZone("AWST", 8*60*60))

func xv(version string) http.Header { return http.Header{"X-V": {version}} }

func send(t *testing.T, method, path string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()

	return sendTo(t, &fakeReader{}, method, path, header)
}

func sendTo(
	t *testing.T,
	reader app.ProductReader,
	method, path string,
	header http.Header,
) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	httpapi.Register(mux, baseURL+"/", reader, func() time.Time { return requestTime })

	req := httptest.NewRequestWithContext(t.Context(), method, path, nil)
	req.Header = header

	rec := httptest.NewRecorder()
	httpapi.AllowAnyOrigin(cds.InteractionID(mux), "/authorize").ServeHTTP(rec, req)

	return rec
}

// Verifies: DH-0455
func TestAnUnknownPathGivesResourceNotFound(t *testing.T) {
	for _, path := range []string{"/", "/cds-au/v1/banking/accounts"} {
		t.Run(path, func(t *testing.T) {
			rec := send(t, http.MethodGet, path, http.Header{})

			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.JSONEq(t, `{"errors": [{
				"code": "urn:au-cds:error:cds-all:Resource/NotFound",
				"title": "Resource Not Found",
				"detail": "The URL is not a data holder API endpoint."
			}]}`, rec.Body.String())
		})
	}
}

func TestEachResponseHasAnInteractionID(t *testing.T) {
	t.Run("from the request", func(t *testing.T) {
		const id = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

		rec := send(t, http.MethodGet, "/", http.Header{"X-Fapi-Interaction-Id": {id}})

		assert.Equal(t, id, rec.Header().Get("x-fapi-interaction-id"))
	})

	t.Run("new when the request has none", func(t *testing.T) {
		rec := send(t, http.MethodGet, "/", http.Header{})

		_, err := uuid.Parse(rec.Header().Get("x-fapi-interaction-id"))
		require.NoError(t, err)
	})
}

// Verifies: DH-0456, DH-0457
func TestEachResponseAllowsAnyOriginAndExposesTheCDSHeaders(t *testing.T) {
	rec := send(t, http.MethodGet, "/", http.Header{"Origin": {"https://app.example"}})

	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "x-v, Retry-After, x-fapi-interaction-id", rec.Header().Get("Access-Control-Expose-Headers"))
}

func TestACORSPreflightRequestIsAllowed(t *testing.T) {
	rec := send(t, http.MethodOptions, "/cds-au/v1/discovery/status", http.Header{
		"Origin":                         {"https://app.example"},
		"Access-Control-Request-Method":  {"GET"},
		"Access-Control-Request-Headers": {"x-v,x-min-v"},
	})

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "x-v,x-min-v", rec.Header().Get("Access-Control-Allow-Headers"))
}

func TestGetStatusGivesOKWithTheTimeOfTheRequest(t *testing.T) {
	rec := send(t, http.MethodGet, "/cds-au/v1/discovery/status", xv("1"))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "1", rec.Header().Get("x-v"))
	assert.JSONEq(t, `{
		"data": {"status": "OK", "updateTime": "2026-10-08T13:30:15Z"},
		"links": {"self": "https://bank.example/cds-au/v1/discovery/status"},
		"meta": {}
	}`, rec.Body.String())
}

func TestGetOutagesGivesNoOutages(t *testing.T) {
	rec := send(t, http.MethodGet, "/cds-au/v1/discovery/outages", xv("1"))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "1", rec.Header().Get("x-v"))
	assert.JSONEq(t, `{
		"data": {"outages": []},
		"links": {"self": "https://bank.example/cds-au/v1/discovery/outages"},
		"meta": {}
	}`, rec.Body.String())
}

var discoveryPaths = []string{"/cds-au/v1/discovery/status", "/cds-au/v1/discovery/outages"}

func TestTheDiscoveryAPIsServeVersion1(t *testing.T) {
	tests := map[string]http.Header{
		"x-v 1":                  xv("1"),
		"x-v with leading zeros": xv("001"),
		"x-v higher, x-min-v 1":  {"X-V": {"7"}, "X-Min-V": {"1"}},
		"x-v too large for an integer, x-min-v 1": {"X-V": {"99999999999999999999"}, "X-Min-V": {"1"}},
		"x-min-v higher than x-v, ignored":        {"X-V": {"1"}, "X-Min-V": {"3"}},
		"x-min-v equal to x-v, ignored":           {"X-V": {"1"}, "X-Min-V": {"1"}},
	}
	for _, path := range discoveryPaths {
		for name, header := range tests {
			t.Run(path+" "+name, func(t *testing.T) {
				rec := send(t, http.MethodGet, path, header)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "1", rec.Header().Get("x-v"))
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
			http.Header{"X-V": {"1"}, "X-Min-V": {"-1"}},
			http.StatusBadRequest,
			"urn:au-cds:error:cds-all:Header/InvalidVersion", "Invalid Version",
		},
		"x-v higher than version 1, no x-min-v": {
			xv("7"), http.StatusNotAcceptable,
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
				rec := send(t, http.MethodGet, path, tc.header)

				assert.Equal(t, tc.status, rec.Code)
				assert.Empty(t, rec.Header().Get("x-v"))
				assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))

				var body struct {
					Errors []struct {
						Code   string `json:"code"`
						Title  string `json:"title"`
						Detail string `json:"detail"`
					} `json:"errors"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				require.Len(t, body.Errors, 1)
				assert.Equal(t, tc.code, body.Errors[0].Code)
				assert.Equal(t, tc.title, body.Errors[0].Title)
				assert.NotEmpty(t, body.Errors[0].Detail)
			})
		}
	}
}

// Verifies: DH-2616
func TestAPublicAPIRefusesAnAcceptHeaderThatPermitsNoJSON(t *testing.T) {
	for _, path := range append([]string{"/cds-au/v1/banking/products"}, discoveryPaths...) {
		t.Run(path, func(t *testing.T) {
			rec := send(t, http.MethodGet, path, http.Header{"X-V": {"1"}, "Accept": {"application/xml"}})

			assert.Equal(t, http.StatusNotAcceptable, rec.Code)
			assert.JSONEq(t, `{"errors": [{
				"code": "urn:au-cds:error:cds-all:Header/Invalid",
				"title": "Invalid Header",
				"detail": "Accept"
			}]}`, rec.Body.String())
		})
	}
}

// Verifies: DH-2617
func TestAnotherMethodOnAPublicAPIPathIsNotAllowed(t *testing.T) {
	paths := append([]string{"/cds-au/v1/banking/products", "/cds-au/v1/banking/products/p1"}, discoveryPaths...)
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			rec := send(t, http.MethodPost, path, xv("1"))

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
			assert.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))
			assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
			assert.JSONEq(t, `{"errors": [{
				"code": "urn:au-cds:error:cds-all:GeneralError/Expected",
				"title": "Expected Error Encountered",
				"detail": "The endpoint does not support the method POST."
			}]}`, rec.Body.String())
		})
	}
}

// Verifies: DH-0459
func TestThePagesOfAConsumerDoNotPermitCrossOriginRequests(t *testing.T) {
	page := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := httpapi.AllowAnyOrigin(page, "/authorize")

	for target, allowed := range map[string]string{
		"/authorize":                  "",
		"/authorize/flow-1":           "",
		"/authorize?client_id=1":      "",
		"/authorized":                 "*",
		"/cds-au/v1/discovery/status": "*",
	} {
		t.Run(target, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
			req.Header.Set("Origin", "https://app.example")

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, allowed, rec.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

// Verifies: DH-0459
func TestAPreflightRequestToAPageOfAConsumerGetsNoCORSHeaders(t *testing.T) {
	reached := false
	page := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true

		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodOptions, "/authorize/flow-1", nil)
	req.Header.Set("Origin", "https://app.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	rec := httptest.NewRecorder()
	httpapi.AllowAnyOrigin(page, "/authorize").ServeHTTP(rec, req)

	assert.True(t, reached, "the wrapper does not answer the preflight request")
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
}
