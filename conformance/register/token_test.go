//go:build conformance

package register

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// These participants of services/register/seed/example.json have https://conformance:8443/jwks as their JWKS URI.
const (
	idpTestRecipient       = "bd653a0f-cf64-40ae-b914-aa479f5fb725"
	idpTestActiveProduct   = "33bfd771-cac1-44c2-8940-a82f529b6170"
	idpTestInactiveProduct = "839ffa5f-09dd-4e00-9bbc-38dff23a9b2f"
)

const (
	openIDConfigPath = "/idp/.well-known/openid-configuration"
	assertionType    = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
	registerScope    = "cdr-register:read"
)

func recipientKeys(t *testing.T) *labtest.RecipientKeys {
	t.Helper()

	return labtest.ServeJWKS(t, "JWKS_LISTEN_ADDRESS")
}

type openIDConfig struct {
	Issuer        string `json:"issuer"`
	JWKSURI       string `json:"jwks_uri"`
	TokenEndpoint string `json:"token_endpoint"`
}

func openIDConfiguration(t *testing.T) (openIDConfig, map[string]any) {
	t.Helper()
	resp := call(t, openIDConfigPath, http.Header{})
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

	var cfg openIDConfig
	require.NoError(t, json.Unmarshal(resp.body, &cfg))

	var fields map[string]any
	require.NoError(t, json.Unmarshal(resp.body, &fields))

	return cfg, fields
}

// tokenClient registers the client certificate for the software product.
func tokenClient(t *testing.T, recipient, product, certSetting, keySetting string) tls.Certificate {
	t.Helper()
	recipientKeys(t)
	labtest.RegisterClientCertificate(t, recipient, product, certSetting)

	return labtest.ClientCertificate(t, certSetting, keySetting)
}

func activeTokenClient(t *testing.T) tls.Certificate {
	t.Helper()
	return tokenClient(t, idpTestRecipient, idpTestActiveProduct, "PKI_CLIENT_CERT", "PKI_CLIENT_KEY")
}

type (
	clientAssertion = labtest.ClientAssertion
	tokenReply      = labtest.TokenReply
)

func newAssertion(t *testing.T, product, audience string) clientAssertion {
	t.Helper()

	return labtest.NewAssertion(recipientKeys(t), product, audience)
}

func tokenForm(t *testing.T, product string, a clientAssertion) url.Values {
	t.Helper()

	return labtest.TokenForm(t, product, registerScope, a)
}

func requestToken(t *testing.T, endpoint string, cert tls.Certificate, form url.Values) tokenReply {
	t.Helper()

	return labtest.RequestToken(t, endpoint, cert, form)
}

// requireTokenError checks the RFC 6749 section 5.2 format.
func requireTokenError(t *testing.T, r tokenReply, status int, code string) {
	t.Helper()
	require.Equal(t, status, r.Status, r.Body)
	assert.Equal(t, code, r.Body["error"])
	assert.Contains(t, r.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, "no-store", r.Header.Get("Cache-Control"))
	assert.NotContains(t, r.Body, "access_token")

	for field := range r.Body {
		assert.Contains(t, []string{"error", "error_description", "error_uri"}, field)
	}
}

func TestGetOpenIDProviderConfigHasTheRequiredFieldsAndTheRegisterValues(t *testing.T) {
	cfg, fields := openIDConfiguration(t)

	assert.Equal(t, labtest.Env(t, "REGISTER_PUBLIC_URL")+"/idp", cfg.Issuer)
	assert.Equal(t, labtest.Env(t, "REGISTER_PUBLIC_URL")+openIDConfigPath+"/jwks", cfg.JWKSURI)
	assert.Equal(t, labtest.Env(t, "REGISTER_SECURE_URL")+"/idp/connect/token", cfg.TokenEndpoint)
	assert.Equal(t, []any{"client_credentials"}, fields["grant_types_supported"])
	assert.Equal(t, []any{"private_key_jwt"}, fields["token_endpoint_auth_methods_supported"])
	assert.Equal(t, []any{"PS256", "ES256"}, fields["token_endpoint_auth_signing_alg_values_supported"])
	assert.Equal(t, []any{registerScope}, fields["scopes_supported"])
	assert.Equal(t, true, fields["tls_client_certificate_bound_access_tokens"])
}

func TestAnActiveSoftwareProductGetsATokenThatIsBoundToItsCertificate(t *testing.T) {
	cert := activeTokenClient(t)
	cfg, _ := openIDConfiguration(t)
	sum := sha256.Sum256(cert.Certificate[0])
	thumbprint := base64.RawURLEncoding.EncodeToString(sum[:])

	jwksResp, err := labtest.Client(t).Get(cfg.JWKSURI) //nolint:noctx // the client has a timeout
	require.NoError(t, err)

	defer func() { _ = jwksResp.Body.Close() }()

	var registerKeys jose.JSONWebKeySet
	require.NoError(t, json.NewDecoder(jwksResp.Body).Decode(&registerKeys))
	require.NotEmpty(t, registerKeys.Keys)

	tests := map[string]struct {
		alg jose.SignatureAlgorithm
		key crypto.Signer
		kid string
		aud string
	}{
		"PS256, token endpoint as the audience": {jose.PS256, recipientKeys(t).RSA, "rsa", cfg.TokenEndpoint},
		"ES256, token endpoint as the audience": {jose.ES256, recipientKeys(t).EC, "ec", cfg.TokenEndpoint},
		"PS256, issuer as the audience":         {jose.PS256, recipientKeys(t).RSA, "rsa", cfg.Issuer},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := newAssertion(t, idpTestActiveProduct, tt.aud)
			a.Alg, a.Key, a.KID = tt.alg, tt.key, tt.kid

			r := requestToken(t, cfg.TokenEndpoint, cert, tokenForm(t, idpTestActiveProduct, a))

			require.Equal(t, http.StatusOK, r.Status, r.Body)
			assert.Equal(t, "Bearer", r.Body["token_type"])
			assert.EqualValues(t, 300, r.Body["expires_in"])
			assert.Equal(t, registerScope, r.Body["scope"])
			assert.NotContains(t, r.Body, "refresh_token")
			assert.Equal(t, "no-store", r.Header.Get("Cache-Control"))
			assert.NotEmpty(t, r.Header.Get("x-fapi-interaction-id"))

			parsed, err := jwt.ParseSigned(r.Body["access_token"].(string), []jose.SignatureAlgorithm{jose.PS256})
			require.NoError(t, err, "the access token is a PS256 JWT")

			keys := registerKeys.Key(parsed.Headers[0].KeyID)
			require.Len(t, keys, 1, "the signing key is at the jwks_uri")

			var claims map[string]any
			require.NoError(t, parsed.Claims(keys[0].Key, &claims))
			assert.Equal(t, cfg.Issuer, claims["iss"])
			assert.Equal(t, idpTestActiveProduct, claims["sub"])
			assert.Equal(t, idpTestActiveProduct, claims["client_id"])
			assert.Equal(t, registerScope, claims["scope"])
			assert.Equal(t, map[string]any{"x5t#S256": thumbprint}, claims["cnf"])
			assert.InDelta(
				t,
				300,
				claims["exp"].(float64)-claims["iat"].(float64),
				0,
				"the token expires after 5 minutes",
			)
		})
	}
}

func TestTheTokenEndpointRefusesAClientThatItCannotTrust(t *testing.T) {
	registered := activeTokenClient(t)
	inactive := tokenClient(
		t,
		idpTestRecipient,
		idpTestInactiveProduct,
		"PKI_INACTIVE_CLIENT_CERT",
		"PKI_INACTIVE_CLIENT_KEY",
	)
	cfg, _ := openIDConfiguration(t)
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	usedBefore := newAssertion(t, idpTestActiveProduct, cfg.TokenEndpoint)
	first := requestToken(t, cfg.TokenEndpoint, registered, tokenForm(t, idpTestActiveProduct, usedBefore))
	require.Equal(t, http.StatusOK, first.Status, first.Body)

	tests := map[string]struct {
		cert    tls.Certificate
		product string
		change  func(a *clientAssertion)
	}{
		"a certificate that is not registered, with the same common name": {
			cert: labtest.ClientCertificate(
				t,
				"PKI_UNREGISTERED_CLIENT_CERT",
				"PKI_UNREGISTERED_CLIENT_KEY",
			),
			product: idpTestActiveProduct,
		},
		"a software product that is not ACTIVE": {cert: inactive, product: idpTestInactiveProduct},
		"a JWT ID that was used before": {
			cert:    registered,
			product: idpTestActiveProduct,
			change: func(a *clientAssertion) {
				a.Claims["jti"] = usedBefore.Claims["jti"]
			},
		},
		"a key that is not at the JWKS URI": {
			cert:    registered,
			product: idpTestActiveProduct,
			change:  func(a *clientAssertion) { a.Key = otherKey },
		},
		"an RS256 assertion": {
			cert:    registered,
			product: idpTestActiveProduct,
			change:  func(a *clientAssertion) { a.Alg = jose.RS256 },
		},
		"an audience that is not the Register": {
			cert:    registered,
			product: idpTestActiveProduct,
			change: func(a *clientAssertion) {
				a.Claims["aud"] = "https://holder.example/token"
			},
		},
		"a software product that does not exist": {cert: registered, product: "00000000-0000-0000-0000-000000000000"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := newAssertion(t, tt.product, cfg.TokenEndpoint)
			if tt.change != nil {
				tt.change(&a)
			}

			r := requestToken(t, cfg.TokenEndpoint, tt.cert, tokenForm(t, tt.product, a))

			requireTokenError(t, r, http.StatusUnauthorized, "invalid_client")
		})
	}
}

func TestTheTokenEndpointAcceptsOnlyClientCredentialsWithAJWTAssertion(t *testing.T) {
	cert := activeTokenClient(t)
	cfg, _ := openIDConfiguration(t)

	tests := map[string]struct {
		change func(form url.Values)
		status int
		code   string
	}{
		"authorization_code grant": {
			func(f url.Values) { f.Set("grant_type", "authorization_code") },
			http.StatusBadRequest,
			"unsupported_grant_type",
		},
		"refresh_token grant": {
			func(f url.Values) { f.Set("grant_type", "refresh_token") },
			http.StatusBadRequest,
			"unsupported_grant_type",
		},
		"a scope of the data APIs": {
			func(f url.Values) { f.Set("scope", "bank:accounts.basic:read") },
			http.StatusBadRequest,
			"invalid_scope",
		},
		"no client assertion": {func(f url.Values) {
			f.Del("client_assertion")
			f.Del("client_assertion_type")
		}, http.StatusUnauthorized, "invalid_client"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			form := tokenForm(t, idpTestActiveProduct, newAssertion(t, idpTestActiveProduct, cfg.TokenEndpoint))
			tt.change(form)

			r := requestToken(t, cfg.TokenEndpoint, cert, form)

			requireTokenError(t, r, tt.status, tt.code)
		})
	}
}

func accessToken(t *testing.T, product string, cert tls.Certificate) string {
	t.Helper()

	return labtest.RegisterAccessToken(t, recipientKeys(t), product, cert)
}
