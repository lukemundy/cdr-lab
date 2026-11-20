// Package sectoridentifier gets the document at the sector_identifier_uri of a software product.
package sectoridentifier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

const maxDocumentBytes = 1 << 20

type Documents struct {
	client *http.Client
}

var _ app.SectorIdentifiers = (*Documents)(nil)

// NewDocuments uses the JWKS client, because the data holder connects to the host of a data recipient with the same
// rules for both.
func NewDocuments(transport http.RoundTripper) *Documents {
	return &Documents{client: authserver.NewJWKSClient(transport)}
}

func (d *Documents) RedirectURIs(ctx context.Context, uri string) ([]string, error) {
	uris, err := d.get(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", app.ErrSectorDocumentNotAvailable, err)
	}

	return uris, nil
}

func (d *Documents) get(ctx context.Context, uri string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the server gave status %d", resp.StatusCode)
	}

	var uris []string
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxDocumentBytes)).Decode(&uris); err != nil {
		return nil, fmt.Errorf("read the document: %w", err)
	}

	return uris, nil
}
