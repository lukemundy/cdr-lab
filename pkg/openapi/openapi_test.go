package openapi_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/openapi"
)

// The servers must not affect the match, so the specification gives servers that no test calls.
const specJSON = `{
  "openapi": "3.0.3",
  "info": {"title": "Test", "version": "1"},
  "servers": [{"url": "https://mtls.example.com/api"}],
  "paths": {
    "/statuses": {
      "get": {
        "servers": [{"url": "https://tls.example.com/api"}],
        "responses": {
          "200": {
            "description": "OK",
            "headers": {"x-v": {"required": true, "schema": {"type": "string"}}},
            "content": {"application/json": {"schema": {
              "type": "object",
              "required": ["data"],
              "properties": {"data": {"type": "array", "items": {
                "type": "object",
                "required": ["legalEntityId", "status"],
                "properties": {
                  "legalEntityId": {"type": "string"},
                  "status": {"type": "string", "enum": ["ACTIVE", "REMOVED"]}
                }
              }}}
            }}}
          }
        }
      }
    }
  }
}`

const validStatuses = `{"data":[{"legalEntityId":"a","status":"ACTIVE"}]}`

type fakeResponse struct {
	status  int
	version string
	body    string
}

func validResponse() fakeResponse {
	return fakeResponse{status: http.StatusOK, version: "3", body: validStatuses}
}

func get(t *testing.T, path string, fake fakeResponse) *http.Response {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if fake.version != "" {
			w.Header().Set("x-v", fake.version)
		}

		w.WriteHeader(fake.status)
		_, _ = io.WriteString(w, fake.body)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+path, nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)

	return resp
}

func validate(t *testing.T, basePath, path string, fake fakeResponse) error {
	t.Helper()

	resp := get(t, path, fake)
	defer func() { _ = resp.Body.Close() }()

	return load(t, basePath).ValidateResponse(resp)
}

func load(t *testing.T, basePath string) *openapi.Spec {
	t.Helper()

	spec, err := openapi.Load([]byte(specJSON), basePath)
	require.NoError(t, err)

	return spec
}

func TestValidateResponseAcceptsAValidResponse(t *testing.T) {
	resp := get(t, "/statuses", validResponse())
	defer func() { _ = resp.Body.Close() }()

	require.NoError(t, load(t, "").ValidateResponse(resp))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, validStatuses, string(body), "the caller can still read the body")
}

func TestValidateResponseMatchesThePathBelowTheBasePath(t *testing.T) {
	require.NoError(t, validate(t, "/cds-au/v1", "/cds-au/v1/statuses", validResponse()))

	err := validate(t, "/cds-au/v1", "/statuses", validResponse())
	require.Error(t, err, "a path that is not below the base path is not in the specification")
}

func TestValidateResponseNamesTheInvalidField(t *testing.T) {
	tests := map[string]struct {
		body      string
		wantField string
	}{
		"value not in enum": {
			body:      `{"data":[{"legalEntityId":"a","status":"DELETED"}]}`,
			wantField: "/data/0/status",
		},
		"required field missing": {
			body:      `{"data":[{"status":"ACTIVE"}]}`,
			wantField: "legalEntityId",
		},
		"wrong type": {
			body:      `{"data":[{"legalEntityId":1,"status":"ACTIVE"}]}`,
			wantField: "/data/0/legalEntityId",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fake := validResponse()
			fake.body = tt.body

			err := validate(t, "", "/statuses", fake)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantField)
		})
	}
}

func TestValidateResponseRejectsAMissingRequiredHeader(t *testing.T) {
	fake := validResponse()
	fake.version = ""

	err := validate(t, "", "/statuses", fake)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-v")
}

func TestValidateResponseRejectsAnUndocumentedStatus(t *testing.T) {
	fake := validResponse()
	fake.status = http.StatusTeapot

	err := validate(t, "", "/statuses", fake)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "418")
}

func TestValidateResponseRejectsAnUndocumentedPath(t *testing.T) {
	err := validate(t, "", "/unknown", validResponse())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "/unknown")
}

const closedSpecJSON = `{
  "openapi": "3.0.3",
  "info": {"title": "Test", "version": "1"},
  "paths": {
    "/items": {
      "get": {
        "responses": {
          "200": {
            "description": "OK",
            "content": {"application/json": {"schema": {"$ref": "#/components/schemas/List"}}}
          }
        }
      }
    }
  },
  "components": {"schemas": {
    "List": {
      "type": "object",
      "additionalProperties": false,
      "properties": {"data": {"type": "array", "items": {"$ref": "#/components/schemas/Item"}}}
    },
    "Item": {
      "type": "object",
      "additionalProperties": false,
      "properties": {"id": {"type": "string"}, "meta": %s}
    }
  }}
}`

func loadClosed(t *testing.T, meta string) *openapi.Spec {
	t.Helper()

	spec, err := openapi.Load(fmt.Appendf(nil, closedSpecJSON, meta), "")
	require.NoError(t, err)

	return spec
}

func TestCheckClosedAcceptsADocumentWhereEachObjectIsClosed(t *testing.T) {
	spec := loadClosed(t, `{"type": "object", "additionalProperties": false, "properties": {"a": {"type": "string"}}}`)

	assert.NoError(t, spec.CheckClosed())
}

func TestCheckClosedNamesEachObjectThatPermitsOtherProperties(t *testing.T) {
	tests := map[string]string{
		"no additionalProperties":   `{"type": "object", "properties": {"a": {"type": "string"}}}`,
		"additionalProperties true": `{"type": "object", "additionalProperties": true}`,
		"properties with no type":   `{"properties": {"a": {"type": "string"}}}`,
	}
	for name, meta := range tests {
		t.Run(name, func(t *testing.T) {
			err := loadClosed(t, meta).CheckClosed()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "#/components/schemas/Item/properties/meta")
		})
	}
}

func TestValidateResponseRejectsAPropertyThatAClosedSchemaDoesNotName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[{"id":"a","extra":"b"}]}`)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/items", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	err = loadClosed(t, `{"type": "string"}`).ValidateResponse(resp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "extra")
}

func TestValidateDocumentedResponseAcceptsAnUndocumentedStatus(t *testing.T) {
	resp := get(t, "/statuses", fakeResponse{status: http.StatusUnauthorized, body: `{"error":"invalid_token"}`})
	defer func() { _ = resp.Body.Close() }()

	require.NoError(t, load(t, "").ValidateDocumentedResponse(resp))
}

func TestValidateDocumentedResponseNamesTheInvalidFieldOfADocumentedStatus(t *testing.T) {
	fake := validResponse()
	fake.body = `{"data":[{"legalEntityId":"a","status":"DELETED"}]}`

	resp := get(t, "/statuses", fake)
	defer func() { _ = resp.Body.Close() }()

	err := load(t, "").ValidateDocumentedResponse(resp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "/data/0/status")
}
