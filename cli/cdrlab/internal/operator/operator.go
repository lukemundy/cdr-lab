// Package operator sends requests with the operator token to the operator interface of a service.
package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxBody = 1 << 20

type Client struct {
	// name is the service in an error, for example "the Register".
	name  string
	base  string
	token string
	http  *http.Client
}

func New(name, baseURL, token string, httpClient *http.Client) *Client {
	return &Client{name: name, base: strings.TrimSuffix(baseURL, "/"), token: token, http: httpClient}
}

type Request struct {
	Method, Path, ContentType, Accept string
	Body                              []byte
}

// Do gives the body of a 200 response. Each other status is an error.
func (c *Client) Do(ctx context.Context, r Request) ([]byte, error) {
	var body io.Reader = http.NoBody
	if r.Body != nil {
		body = bytes.NewReader(r.Body)
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, c.base+r.Path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	if r.ContentType != "" {
		req.Header.Set("Content-Type", r.ContentType)
	}

	if r.Accept != "" {
		req.Header.Set("Accept", r.Accept)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.refusal(resp, data)
	}

	return data, nil
}

func (c *Client) PostJSON(ctx context.Context, path string, body []byte, into any) error {
	data, err := c.Do(ctx, Request{
		Method: http.MethodPost, Path: path, ContentType: "application/json", Accept: "application/json", Body: body,
	})
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, into); err != nil {
		return fmt.Errorf("%s gave a response that is not valid JSON: %w", c.name, err)
	}

	return nil
}

// refusal gives the status and the detail of an RFC 9457 problem.
func (c *Client) refusal(resp *http.Response, data []byte) error {
	var p struct {
		Detail string `json:"detail"`
	}

	if json.Unmarshal(data, &p) != nil || p.Detail == "" {
		return fmt.Errorf("%s gave %s", c.name, resp.Status)
	}

	return fmt.Errorf("%s gave %s: %s", c.name, resp.Status, p.Detail)
}
