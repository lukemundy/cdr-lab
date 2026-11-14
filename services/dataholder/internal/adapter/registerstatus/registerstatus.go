// Package registerstatus is a client of the software product status API of the Register.
package registerstatus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
)

const (
	statusPath = "/cdr-register/v1/all/data-recipients/brands/software-products/status"
	version    = "3"
	maxBody    = 4 << 20
	timeout    = 10 * time.Second
)

var _ app.StatusSource = Client{}

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient gives a client that does not follow a redirect, because a redirect could send the request to a host that
// is not the Register.
func NewClient(baseURL string, transport http.RoundTripper) Client {
	return Client{
		BaseURL: baseURL,
		HTTP: &http.Client{
			Transport:     otelhttp.NewTransport(transport),
			Timeout:       timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

type statusesJSON struct {
	Data []struct {
		SoftwareProductID string `json:"softwareProductId"`
		Status            string `json:"status"`
	} `json:"data"`
}

func (c Client) SoftwareProductStatuses(ctx context.Context, etag string) (app.StatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(c.BaseURL, "/")+statusPath, nil)
	if err != nil {
		return app.StatusResponse{}, err
	}

	req.Header.Set("x-v", version)

	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return app.StatusResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotModified:
		return app.StatusResponse{NotModified: true}, nil
	case http.StatusOK:
	default:
		return app.StatusResponse{}, fmt.Errorf("the Register gave the status %d", resp.StatusCode)
	}

	var body statusesJSON
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(&body); err != nil {
		return app.StatusResponse{}, fmt.Errorf("decode the statuses: %w", err)
	}

	statuses := make(map[string]string, len(body.Data))
	for _, item := range body.Data {
		statuses[item.SoftwareProductID] = item.Status
	}

	return app.StatusResponse{Statuses: statuses, ETag: resp.Header.Get("ETag")}, nil
}
