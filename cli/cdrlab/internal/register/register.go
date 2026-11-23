// Package register calls the certificate paths of the Register operator interface.
package register

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

const maxBody = 1 << 20

type Client struct {
	base  string
	token string
	http  *http.Client
}

func New(baseURL, token string, httpClient *http.Client) *Client {
	return &Client{base: strings.TrimSuffix(baseURL, "/"), token: token, http: httpClient}
}

func (c *Client) CABundle(ctx context.Context) ([]*x509.Certificate, error) {
	certs, err := c.certificates(ctx, http.MethodGet, "/v1/cdr-ca-bundle", nil)
	if err != nil {
		return nil, fmt.Errorf("get the CDR CA bundle from the Register: %w", err)
	}

	return certs, nil
}

// IssueServer gives the certificate and then the intermediate.
func (c *Client) IssueServer(
	ctx context.Context,
	legalEntityID, brandID string,
	csr []byte,
) ([]*x509.Certificate, error) {
	return c.issue(ctx, "/v1/data-holders/"+url.PathEscape(legalEntityID)+"/brands/"+url.PathEscape(brandID), csr)
}

// IssueClient gives the certificate and then the intermediate.
func (c *Client) IssueClient(
	ctx context.Context, legalEntityID, softwareProductID string, csr []byte,
) ([]*x509.Certificate, error) {
	return c.issue(
		ctx,
		"/v1/data-recipients/"+url.PathEscape(
			legalEntityID,
		)+"/software-products/"+url.PathEscape(
			softwareProductID,
		),
		csr,
	)
}

func (c *Client) issue(ctx context.Context, owner string, csr []byte) ([]*x509.Certificate, error) {
	certs, err := c.certificates(ctx, http.MethodPost, owner+"/certificates/issue", csr)
	if err != nil {
		return nil, fmt.Errorf("issue the certificate through the Register: %w", err)
	}

	return certs, nil
}

func (c *Client) certificates(ctx context.Context, method, path string, body []byte) ([]*x509.Certificate, error) {
	var reader io.Reader = http.NoBody
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/pem-certificate-chain")

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
		return nil, refusal(resp, data)
	}

	return ca.ParseCertificates(data)
}

// refusal gives the status and the detail of an RFC 9457 problem.
func refusal(resp *http.Response, data []byte) error {
	var p struct {
		Detail string `json:"detail"`
	}

	if json.Unmarshal(data, &p) != nil || p.Detail == "" {
		return fmt.Errorf("the Register gave %s", resp.Status)
	}

	return fmt.Errorf("the Register gave %s: %s", resp.Status, p.Detail)
}
