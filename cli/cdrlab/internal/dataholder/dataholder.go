// Package dataholder calls the customer load of the data holder operator interface.
package dataholder

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/operator"
)

type Client struct {
	operator *operator.Client
}

func New(baseURL, token string, httpClient *http.Client) *Client {
	return &Client{operator: operator.New("the data holder", baseURL, token, httpClient)}
}

// CustomerLoad has only counts, because a customer number goes to no output.
type CustomerLoad struct {
	Created, Existed int
}

// LoadCustomers sends the customer load file with no change.
func (c *Client) LoadCustomers(ctx context.Context, file []byte) (CustomerLoad, error) {
	var response struct {
		Customers []struct {
			Created bool `json:"created"`
		} `json:"customers"`
	}

	if err := c.operator.PostJSON(ctx, "/v1/customer-loads", file, &response); err != nil {
		return CustomerLoad{}, fmt.Errorf("load the customers into the data holder: %w", err)
	}

	var load CustomerLoad

	for _, customer := range response.Customers {
		if customer.Created {
			load.Created++
		} else {
			load.Existed++
		}
	}

	return load, nil
}
