package cds

import "net/http"

type Response[T any] struct {
	Data  T        `json:"data"`
	Links Links    `json:"links"`
	Meta  struct{} `json:"meta"`
}

type Links struct {
	Self string `json:"self"`
}

// NewResponse needs a baseURL with no trailing slash.
func NewResponse[T any](r *http.Request, baseURL string, data T) Response[T] {
	return Response[T]{Data: data, Links: Links{Self: baseURL + r.URL.EscapedPath()}}
}
