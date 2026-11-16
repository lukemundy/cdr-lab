package authserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

const (
	dataHolderIssuer = "https://id.bank.example"
	dataHolderSecure = "https://mtls.bank.example"
)

type dataHolderClients map[string]authserver.DataHolderClient

func (c dataHolderClients) Client(_ context.Context, id string) (authserver.DataHolderClient, error) {
	client, ok := c[id]
	if !ok {
		return authserver.DataHolderClient{}, authserver.ErrClientNotFound
	}

	return client, nil
}

func dataHolderConfig(t *testing.T) authserver.DataHolderConfig {
	t.Helper()

	key, err := authserver.NewSigningKey()
	require.NoError(t, err)

	return authserver.DataHolderConfig{
		Issuer:        dataHolderIssuer,
		SecureBaseURL: dataHolderSecure,
		Scopes:        []string{"openid", "profile", "cdr:registration", "bank:accounts.basic:read"},
		Clients:       dataHolderClients{},
		JTIs:          &jtiStore{seen: map[string]bool{}},
		SigningKeys: func(context.Context) ([]authserver.SigningKey, error) {
			return []authserver.SigningKey{key}, nil
		},
		Pages:          &pages{},
		Grants:         &grantStore{grants: map[string]authserver.Grant{}},
		RevokedTokens:  &revokedTokenStore{expiries: map[string]time.Time{}},
		Arrangements:   arrangementGrants{},
		ClientStatuses: &clientStatuses{},
	}
}

func getDataHolder(t *testing.T, cfg authserver.DataHolderConfig, target string) *httptest.ResponseRecorder {
	t.Helper()

	dataHolder, err := authserver.NewDataHolder(cfg)
	require.NoError(t, err)

	mux := http.NewServeMux()
	dataHolder.RegisterPublic(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))

	return rec
}

// Verifies: AUTH-0001, AUTH-0004, AUTH-0005
func TestTheDataHolderConfigurationHasEachFieldThatTheCDSRequires(t *testing.T) {
	rec := getDataHolder(t, dataHolderConfig(t), "/.well-known/openid-configuration")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{
		"issuer": "https://id.bank.example",
		"jwks_uri": "https://id.bank.example/jwks",
		"authorization_endpoint": "https://id.bank.example/authorize",
		"registration_endpoint": "https://mtls.bank.example/register",
		"token_endpoint": "https://mtls.bank.example/token",
		"userinfo_endpoint": "https://mtls.bank.example/userinfo",
		"introspection_endpoint": "https://mtls.bank.example/introspect",
		"revocation_endpoint": "https://mtls.bank.example/revoke",
		"pushed_authorization_request_endpoint": "https://mtls.bank.example/par",
		"cdr_arrangement_revocation_endpoint": "https://mtls.bank.example/arrangements/revoke",
		"scopes_supported": ["openid", "profile", "cdr:registration", "bank:accounts.basic:read"],
		"claims_supported": [
			"sub", "acr", "auth_time", "name", "given_name", "family_name", "updated_at", "sharing_duration"
		],
		"claims_parameter_supported": true,
		"acr_values_supported": ["urn:cds.au:cdr:2"],
		"subject_types_supported": ["pairwise"],
		"grant_types_supported": ["authorization_code", "client_credentials", "refresh_token"],
		"response_types_supported": ["code"],
		"response_modes_supported": ["jwt"],
		"code_challenge_methods_supported": ["S256"],
		"token_endpoint_auth_methods_supported": ["private_key_jwt"],
		"token_endpoint_auth_signing_alg_values_supported": ["PS256", "ES256"],
		"request_object_signing_alg_values_supported": ["PS256", "ES256"],
		"id_token_signing_alg_values_supported": ["PS256"],
		"authorization_signing_alg_values_supported": ["PS256"],
		"tls_client_certificate_bound_access_tokens": true,
		"require_pushed_authorization_requests": true
	}`, rec.Body.String())
}

func TestTheDataHolderEndpointURLsHaveNoDoubleSlashWhenABaseURLEndsWithASlash(t *testing.T) {
	cfg := dataHolderConfig(t)
	cfg.Issuer += "/"
	cfg.SecureBaseURL += "/"

	rec := getDataHolder(t, cfg, "/.well-known/openid-configuration")

	var configuration map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &configuration))
	assert.Equal(t, "https://id.bank.example", configuration["issuer"])
	assert.Equal(t, "https://id.bank.example/jwks", configuration["jwks_uri"])
	assert.Equal(t, "https://mtls.bank.example/token", configuration["token_endpoint"])
}

// Verifies: AUTH-0006
func TestTheDataHolderJWKSHasThePublicSigningKeyOnly(t *testing.T) {
	cfg := dataHolderConfig(t)
	keys, err := cfg.SigningKeys(t.Context())
	require.NoError(t, err)

	rec := getDataHolder(t, cfg, "/jwks")

	require.Equal(t, http.StatusOK, rec.Code)

	var set struct {
		Keys []map[string]any `json:"keys"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &set))
	require.Len(t, set.Keys, 1)
	assert.Equal(t, keys[0].ID, set.Keys[0]["kid"])
	assert.Equal(t, "RSA", set.Keys[0]["kty"])
	assert.Equal(t, "PS256", set.Keys[0]["alg"])

	for _, private := range []string{"d", "p", "q", "dp", "dq", "qi"} {
		assert.NotContains(t, set.Keys[0], private)
	}
}

func TestTheDataHolderJWKSGivesAServerErrorWhenTheKeysAreNotAvailable(t *testing.T) {
	cfg := dataHolderConfig(t)
	cfg.SigningKeys = func(context.Context) ([]authserver.SigningKey, error) {
		return nil, errors.New("database failure")
	}

	rec := getDataHolder(t, cfg, "/jwks")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "database failure")
}

func TestNewDataHolderRefusesAConfigurationThatIsNotComplete(t *testing.T) {
	tests := map[string]func(*authserver.DataHolderConfig){
		"no issuer":          func(c *authserver.DataHolderConfig) { c.Issuer = "" },
		"no secure base URL": func(c *authserver.DataHolderConfig) { c.SecureBaseURL = "" },
		"no scopes":          func(c *authserver.DataHolderConfig) { c.Scopes = nil },
		"no clients":         func(c *authserver.DataHolderConfig) { c.Clients = nil },
		"no JTI store":       func(c *authserver.DataHolderConfig) { c.JTIs = nil },
		"no signing keys":    func(c *authserver.DataHolderConfig) { c.SigningKeys = nil },
		"no consumer pages":  func(c *authserver.DataHolderConfig) { c.Pages = nil },
		"no grant store":     func(c *authserver.DataHolderConfig) { c.Grants = nil },
		"no revoked tokens":  func(c *authserver.DataHolderConfig) { c.RevokedTokens = nil },
		"no arrangements":    func(c *authserver.DataHolderConfig) { c.Arrangements = nil },
		"no client statuses": func(c *authserver.DataHolderConfig) { c.ClientStatuses = nil },
	}

	for name, remove := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := dataHolderConfig(t)
			remove(&cfg)

			_, err := authserver.NewDataHolder(cfg)

			assert.Error(t, err)
		})
	}
}

func TestTheRegistrationPathIsThePathOfTheRegistrationEndpointInTheConfiguration(t *testing.T) {
	cfg := dataHolderConfig(t)
	dataHolder, err := authserver.NewDataHolder(cfg)
	require.NoError(t, err)

	rec := getDataHolder(t, cfg, "/.well-known/openid-configuration")

	var configuration map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &configuration))
	assert.Equal(t, dataHolderSecure+dataHolder.RegistrationPath(), configuration["registration_endpoint"])
}
