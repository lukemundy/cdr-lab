package registrationjwt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

// MinRefreshInterval limits the requests to the Register that SSAs with invented key IDs can cause.
const MinRefreshInterval = time.Second

const maxJWKSBytes = 1 << 20

var errKeyNotFound = errors.New("the JWKS has no key with the key ID")

// RegisterKeys keeps a copy of the Register JWKS, and gets it again when an SSA has a key ID that the copy does not
// have.
type RegisterKeys struct {
	Now func() time.Time

	url    string
	client *http.Client
	mu     sync.Mutex
	keys   jose.JSONWebKeySet
	loaded time.Time
}

func NewRegisterKeys(url string, transport http.RoundTripper) *RegisterKeys {
	return &RegisterKeys{Now: time.Now, url: url, client: authserver.NewJWKSClient(transport)}
}

// key holds the lock during the request, so that concurrent registrations load the JWKS one time.
func (r *RegisterKeys) key(ctx context.Context, keyID string) (jose.JSONWebKey, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if found := r.keys.Key(keyID); len(found) == 1 {
		return found[0], nil
	}

	if now := r.Now(); r.loaded.IsZero() || now.Sub(r.loaded) >= MinRefreshInterval {
		keys, err := getJWKS(ctx, r.client, r.url)
		if err != nil {
			return jose.JSONWebKey{}, fmt.Errorf("get the Register JWKS: %w", err)
		}

		r.keys, r.loaded = keys, now
	}

	if found := r.keys.Key(keyID); len(found) == 1 {
		return found[0], nil
	}

	return jose.JSONWebKey{}, errKeyNotFound
}

func getJWKS(ctx context.Context, client *http.Client, url string) (jose.JSONWebKeySet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return jose.JSONWebKeySet{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return jose.JSONWebKeySet{}, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return jose.JSONWebKeySet{}, fmt.Errorf("the JWKS server gave status %d", resp.StatusCode)
	}

	var keys jose.JSONWebKeySet
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxJWKSBytes)).Decode(&keys); err != nil {
		return jose.JSONWebKeySet{}, fmt.Errorf("read the JWKS: %w", err)
	}

	return keys, nil
}
