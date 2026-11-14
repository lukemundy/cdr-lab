//go:build conformance

package dataholder

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// This participant of services/register/seed/example.json has https://conformance:8444/jwks as its JWKS URI.
const (
	registrationTestRecipient = "6f1d2a57-0c5e-4b0e-9d0a-3f6c1b7e8a41"
	registrationTestProduct   = "d4f6a8c0-2e4b-4c6d-9e8f-0a1b2c3d4e5f"
)

const (
	openIDConfigPath = "/.well-known/openid-configuration"
	minimumRSABits   = 2048
)

// softwareProductOf registers the client certificate of the tests for the software product, and gives the certificate.
func softwareProductOf(t *testing.T, softwareProductID string) tls.Certificate {
	t.Helper()

	recipientKeys(t)
	labtest.RegisterClientCertificate(t, registrationTestRecipient, softwareProductID, "PKI_REGISTRATION_CLIENT_CERT")

	return labtest.ClientCertificate(t, "PKI_REGISTRATION_CLIENT_CERT", "PKI_REGISTRATION_CLIENT_KEY")
}

func softwareProduct(t *testing.T) tls.Certificate {
	t.Helper()

	return softwareProductOf(t, registrationTestProduct)
}

func recipientKeys(t *testing.T) *labtest.RecipientKeys {
	t.Helper()

	return labtest.ServeJWKS(t, "DATAHOLDER_JWKS_LISTEN_ADDRESS")
}

// brandOnTheRegister gets the brand with Get Data Holder Brands, as a software product does.
func brandOnTheRegister(t *testing.T) labtest.DataHolderBrand {
	t.Helper()

	cert := softwareProduct(t)
	token := labtest.RegisterAccessToken(t, recipientKeys(t), registrationTestProduct, cert)

	return labtest.RegisteredDataHolderBrand(t, token, cert)
}

// bankName is the brand name of the data holder on the Register, which its pages and its products give.
func bankName(t *testing.T) string {
	t.Helper()

	name := brandOnTheRegister(t).BrandName
	require.NotEmpty(t, name)

	return name
}

func openIDConfiguration(t *testing.T, issuer string) map[string]any {
	t.Helper()

	return labtest.OpenIDConfiguration(t, issuer)
}

func TestTheRegisterGivesTheIssuerTheSecureBaseURIAndTheJWKSOfTheBrand(t *testing.T) {
	brand := brandOnTheRegister(t)

	configuration := openIDConfiguration(t, brand.EndpointDetail.InfosecBaseURI)

	assert.Equal(t, brand.EndpointDetail.InfosecBaseURI, configuration["issuer"])
	require.Len(t, brand.AuthDetails, 1)
	assert.Equal(t, brand.AuthDetails[0].JWKSEndpoint, configuration["jwks_uri"])
	require.IsType(t, "", configuration["token_endpoint"])
	assert.True(
		t,
		strings.HasPrefix(configuration["token_endpoint"].(string), brand.EndpointDetail.ResourceBaseURI+"/"),
		"the token endpoint is below the secure base URI",
	)
}

func TestTheOpenIDProviderConfigurationHasEachFieldAndValueOfTheDesign(t *testing.T) {
	brand := brandOnTheRegister(t)

	configuration := openIDConfiguration(t, brand.EndpointDetail.InfosecBaseURI)

	endpoints := map[string]string{
		"jwks_uri":                              brand.EndpointDetail.InfosecBaseURI,
		"authorization_endpoint":                brand.EndpointDetail.InfosecBaseURI,
		"registration_endpoint":                 brand.EndpointDetail.ResourceBaseURI,
		"token_endpoint":                        brand.EndpointDetail.ResourceBaseURI,
		"userinfo_endpoint":                     brand.EndpointDetail.ResourceBaseURI,
		"introspection_endpoint":                brand.EndpointDetail.ResourceBaseURI,
		"revocation_endpoint":                   brand.EndpointDetail.ResourceBaseURI,
		"pushed_authorization_request_endpoint": brand.EndpointDetail.ResourceBaseURI,
		"cdr_arrangement_revocation_endpoint":   brand.EndpointDetail.ResourceBaseURI,
	}
	for field, baseURI := range endpoints {
		t.Run(field, func(t *testing.T) {
			base, err := url.Parse(baseURI)
			require.NoError(t, err)

			require.IsType(t, "", configuration[field])
			endpoint, err := url.Parse(configuration[field].(string))
			require.NoError(t, err)

			assert.Equal(t, "https", endpoint.Scheme)
			assert.Equal(t, base.Host, endpoint.Host)
			assert.NotEmpty(t, endpoint.Path)

			delete(configuration, field)
		})
	}

	assert.Equal(t, map[string]any{
		"issuer": brand.EndpointDetail.InfosecBaseURI,
		"scopes_supported": []any{
			"openid", "profile", "cdr:registration", "common:customer.basic:read",
			"bank:accounts.basic:read", "bank:accounts.detail:read", "bank:transactions:read",
		},
		"claims_supported": []any{
			"sub", "acr", "auth_time", "name", "given_name", "family_name", "updated_at", "sharing_duration",
		},
		"claims_parameter_supported": true,
		"acr_values_supported":       []any{"urn:cds.au:cdr:2"},
		"subject_types_supported":    []any{"pairwise"},
		"grant_types_supported": []any{
			"authorization_code",
			"client_credentials",
			"refresh_token",
		},
		"response_types_supported":                         []any{"code"},
		"response_modes_supported":                         []any{"jwt"},
		"code_challenge_methods_supported":                 []any{"S256"},
		"token_endpoint_auth_methods_supported":            []any{"private_key_jwt"},
		"token_endpoint_auth_signing_alg_values_supported": []any{"PS256", "ES256"},
		"request_object_signing_alg_values_supported":      []any{"PS256", "ES256"},
		"id_token_signing_alg_values_supported":            []any{"PS256"},
		"authorization_signing_alg_values_supported":       []any{"PS256"},
		"tls_client_certificate_bound_access_tokens":       true,
		"require_pushed_authorization_requests":            true,
	}, configuration, "the configuration has no other field, and thus no encryption field")
}

func TestTheJWKSHasOneRSAPublicKeyOf2048BitsOrMore(t *testing.T) {
	brand := brandOnTheRegister(t)

	resp := get(t, brand.AuthDetails[0].JWKSEndpoint, http.Header{})

	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))

	var set struct {
		Keys []map[string]any `json:"keys"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &set))
	require.Len(t, set.Keys, 1)

	key := set.Keys[0]
	assert.Equal(t, "RSA", key["kty"])
	assert.NotEmpty(t, key["kid"])
	require.IsType(t, "", key["n"])
	modulus, err := base64.RawURLEncoding.DecodeString(key["n"].(string))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(modulus)*8, minimumRSABits)

	for _, private := range []string{"d", "p", "q", "dp", "dq", "qi"} {
		assert.NotContains(t, key, private)
	}
}

func TestTheConfigurationAndTheJWKSAllowAnyOriginAndHaveAnInteractionID(t *testing.T) {
	brand := brandOnTheRegister(t)

	for name, target := range map[string]string{
		"configuration": brand.EndpointDetail.InfosecBaseURI + openIDConfigPath,
		"JWKS":          brand.AuthDetails[0].JWKSEndpoint,
	} {
		t.Run(name, func(t *testing.T) {
			resp := get(t, target, http.Header{"Origin": {"https://app.example"}})

			assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))

			_, err := uuid.Parse(resp.Header.Get("x-fapi-interaction-id"))
			assert.NoError(t, err)
		})
	}
}

func secureBaseURI(t *testing.T) string {
	t.Helper()

	return brandOnTheRegister(t).EndpointDetail.ResourceBaseURI
}

func TestTheSecureAPIsAcceptAClientCertificateFromTheCDRCA(t *testing.T) {
	labtest.AssertAcceptsAClientCertificateFromTheCDRCA(t, secureBaseURI(t), softwareProduct(t))
}

func TestTheSecureAPIsRefuseAClientWithoutAValidCDRCertificate(t *testing.T) {
	labtest.AssertRefusesAClientWithoutAValidCDRCertificate(t, secureBaseURI(t))
}

func TestTheSecureAPIsPresentACertificateFromTheCDRCAOnly(t *testing.T) {
	labtest.AssertPresentsACertificateFromTheCDRCAOnly(t, secureBaseURI(t), softwareProduct(t))
}

func TestTheSecureAPIsAcceptOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t *testing.T) {
	labtest.AssertMutualTLSAcceptsOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t, secureBaseURI(t), softwareProduct(t))
}
