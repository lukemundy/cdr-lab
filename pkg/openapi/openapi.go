// Package openapi validates HTTP responses against an OpenAPI 3.0 document.
package openapi

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

type Spec struct {
	doc    *openapi3.T
	router routers.Router
}

// Load ignores the servers of the specification and matches requests on basePath and the
// path only, so that one specification validates each deployment.
func Load(specJSON []byte, basePath string) (*Spec, error) {
	doc, err := openapi3.NewLoader().LoadFromData(specJSON)
	if err != nil {
		return nil, fmt.Errorf("load the OpenAPI specification: %w", err)
	}

	if err := doc.Validate(openapi3.NewLoader().Context); err != nil {
		return nil, fmt.Errorf("validate the OpenAPI specification: %w", err)
	}

	doc.Servers = nil
	if basePath != "" {
		doc.Servers = openapi3.Servers{{URL: basePath}}
	}

	for _, path := range doc.Paths.Map() {
		path.Servers = nil
		for _, operation := range path.Operations() {
			operation.Servers = nil
		}
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("make the OpenAPI router: %w", err)
	}

	return &Spec{doc: doc, router: router}, nil
}

// CheckClosed names each object schema that permits a property that it does not name.
func (s *Spec) CheckClosed() error {
	var open []string

	visited := map[*openapi3.Schema]bool{}

	var walk func(ref *openapi3.SchemaRef, at string)

	walk = func(ref *openapi3.SchemaRef, at string) {
		if ref == nil || ref.Value == nil || visited[ref.Value] {
			return
		}

		schema := ref.Value
		visited[schema] = true

		isObject := schema.Type.Is(openapi3.TypeObject) || len(schema.Properties) > 0
		if isObject && (schema.AdditionalProperties.Has == nil || *schema.AdditionalProperties.Has) {
			open = append(open, at)
		}

		for _, name := range slices.Sorted(maps.Keys(schema.Properties)) {
			walk(schema.Properties[name], at+"/properties/"+name)
		}

		walk(schema.Items, at+"/items")
		walk(schema.AdditionalProperties.Schema, at+"/additionalProperties")

		combined := map[string]openapi3.SchemaRefs{"allOf": schema.AllOf, "anyOf": schema.AnyOf, "oneOf": schema.OneOf}
		for kind, refs := range combined {
			for i, r := range refs {
				walk(r, fmt.Sprintf("%s/%s/%d", at, kind, i))
			}
		}
	}

	for _, name := range slices.Sorted(maps.Keys(s.doc.Components.Schemas)) {
		walk(s.doc.Components.Schemas[name], "#/components/schemas/"+name)
	}

	for _, path := range s.doc.Paths.InMatchingOrder() {
		for method, operation := range s.doc.Paths.Value(path).Operations() {
			for status, response := range operation.Responses.Map() {
				for mediaType, content := range response.Value.Content {
					walk(content.Schema, fmt.Sprintf("#/paths/%s/%s/responses/%s/content/%s/schema",
						path, strings.ToLower(method), status, mediaType))
				}
			}
		}
	}

	if len(open) > 0 {
		slices.Sort(open)

		return fmt.Errorf("these object schemas permit properties that they do not name: %s", strings.Join(open, ", "))
	}

	return nil
}

// ValidateResponse returns an error that names each invalid field, and
// replaces resp.Body so that the caller can read it again.
func (s *Spec) ValidateResponse(resp *http.Response) error {
	req := resp.Request

	route, pathParams, err := s.router.FindRoute(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", req.Method, req.URL.Path, err)
	}

	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if err != nil {
		return fmt.Errorf("read the response body: %w", err)
	}

	resp.Body = io.NopCloser(bytes.NewReader(body))

	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		Status: resp.StatusCode,
		Header: resp.Header,
		Body:   io.NopCloser(bytes.NewReader(body)),
		Options: &openapi3filter.Options{
			IncludeResponseStatus: true,
			MultiError:            true,
		},
	}
	if err := openapi3filter.ValidateResponse(req.Context(), input); err != nil {
		return fmt.Errorf("%s %s: status %d: %w", req.Method, route.Path, resp.StatusCode, err)
	}

	return nil
}
