// Package register calls the certificate paths and the participant load of the Register operator interface.
package register

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/operator"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

type Client struct {
	operator *operator.Client
}

func New(baseURL, token string, httpClient *http.Client) *Client {
	return &Client{operator: operator.New("the Register", baseURL, token, httpClient)}
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

// ParticipantLoad gives the identifiers of the file in the sequence of the file.
type ParticipantLoad struct {
	Added    []string `json:"added"`
	Existing []string `json:"existing"`
}

// LoadParticipants sends the load file with no change.
func (c *Client) LoadParticipants(ctx context.Context, file []byte) (ParticipantLoad, error) {
	var load ParticipantLoad
	if err := c.operator.PostJSON(ctx, "/v1/participant-loads", file, &load); err != nil {
		return ParticipantLoad{}, fmt.Errorf("load the participants into the Register: %w", err)
	}

	return load, nil
}

func (c *Client) issue(ctx context.Context, owner string, csr []byte) ([]*x509.Certificate, error) {
	certs, err := c.certificates(ctx, http.MethodPost, owner+"/certificates/issue", csr)
	if err != nil {
		return nil, fmt.Errorf("issue the certificate through the Register: %w", err)
	}

	return certs, nil
}

func (c *Client) certificates(ctx context.Context, method, path string, body []byte) ([]*x509.Certificate, error) {
	data, err := c.operator.Do(ctx, operator.Request{
		Method: method, Path: path, Accept: "application/pem-certificate-chain", Body: body,
	})
	if err != nil {
		return nil, err
	}

	return ca.ParseCertificates(data)
}
