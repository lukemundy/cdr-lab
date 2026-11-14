package authserver

import (
	"net/http"
	"time"
)

const jwksRequestTimeout = 5 * time.Second

// NewJWKSClient gives a client that does not follow a redirect, because a redirect could send the request to a host
// that the Register, an SSA or a registration does not name. A nil transport is http.DefaultTransport.
func NewJWKSClient(transport http.RoundTripper) *http.Client {
	return &http.Client{
		Transport:     transport,
		Timeout:       jwksRequestTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}
