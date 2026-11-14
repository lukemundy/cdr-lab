package labtest

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/require"
)

const (
	registerBrandsPath = "/cdr-register/v1/banking/data-holders/brands?page-size=1000"
	openIDConfigPath   = "/.well-known/openid-configuration"
)

func secureClient(t *testing.T, cert tls.Certificate) *http.Client {
	t.Helper()

	return &http.Client{Timeout: RequestTimeout, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      CDRCABundle(t),
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}}
}

// registerSecureGet calls a secure API of the Register and requires 200.
func registerSecureGet(t *testing.T, cert tls.Certificate, path, token, version string) []byte {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, Env(t, "REGISTER_SECURE_URL")+path, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-v", version)

	resp, err := secureClient(t, cert).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))

	return body
}

// DataHolderBrand has the values of a brand that Get Data Holder Brands gives.
type DataHolderBrand struct {
	DataHolderBrandID string `json:"dataHolderBrandId"`
	BrandName         string `json:"brandName"`
	LogoURI           string `json:"logoUri"`
	EndpointDetail    struct {
		ResourceBaseURI string `json:"resourceBaseUri"`
		InfosecBaseURI  string `json:"infosecBaseUri"`
	} `json:"endpointDetail"`
	AuthDetails []struct {
		JWKSEndpoint string `json:"jwksEndpoint"`
	} `json:"authDetails"`
}

// RegisteredDataHolderBrand gets the brand of DATAHOLDER_BRAND_ID from the Register, as a software product does.
func RegisteredDataHolderBrand(t *testing.T, token string, cert tls.Certificate) DataHolderBrand {
	t.Helper()

	var page struct {
		Data []DataHolderBrand `json:"data"`
	}
	require.NoError(t, json.Unmarshal(registerSecureGet(t, cert, registerBrandsPath, token, "3"), &page))

	brandID := Env(t, "DATAHOLDER_BRAND_ID")
	for _, brand := range page.Data {
		if brand.DataHolderBrandID == brandID {
			return brand
		}
	}

	t.Fatalf("the Register has no banking brand %s", brandID)

	return DataHolderBrand{}
}

func OpenIDConfiguration(t *testing.T, issuer string) map[string]any {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, issuer+openIDConfigPath, nil)
	require.NoError(t, err)

	resp, err := Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var configuration map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&configuration))

	return configuration
}

// SoftwareStatement gets a new SSA of the software product from the Register.
func SoftwareStatement(
	t *testing.T,
	token string,
	cert tls.Certificate,
	dataRecipientBrandID, softwareProductID string,
) string {
	t.Helper()

	path := "/cdr-register/v1/all/data-recipients/brands/" + dataRecipientBrandID +
		"/software-products/" + softwareProductID + "/ssa"

	return string(registerSecureGet(t, cert, path, token, "4"))
}

func SignJWT(t *testing.T, alg jose.SignatureAlgorithm, key crypto.Signer, keyID string, claims map[string]any) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: alg, Key: jose.JSONWebKey{Key: key, KeyID: keyID}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	require.NoError(t, err)
	signed, err := jwt.Signed(signer).Claims(claims).Serialize()
	require.NoError(t, err)

	return signed
}

// RegistrationClaims gives the claims of a valid registration request, which a test can change before it signs.
func RegistrationClaims(softwareProductID, issuer, ssa string) map[string]any {
	now := time.Now()
	//nolint:gosec // G101 reads the protocol field token_endpoint_auth_method as a credential
	return map[string]any{
		"iss":                               softwareProductID,
		"aud":                               issuer,
		"iat":                               now.Unix(),
		"exp":                               now.Add(time.Minute).Unix(),
		"jti":                               rand.Text(),
		"redirect_uris":                     []string{"https://conformance.example/callback"},
		"token_endpoint_auth_method":        "private_key_jwt",
		"token_endpoint_auth_signing_alg":   "PS256",
		"grant_types":                       []string{"client_credentials", "authorization_code", "refresh_token"},
		"response_types":                    []string{"code"},
		"application_type":                  "web",
		"id_token_signed_response_alg":      "PS256",
		"authorization_signed_response_alg": "PS256",
		"request_object_signing_alg":        "PS256",
		"software_statement":                ssa,
	}
}

// PostRegistration gives a response whose body the caller must close.
func PostRegistration(t *testing.T, endpoint string, cert tls.Certificate, contentType, body string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, endpoint, strings.NewReader(body))
	require.NoError(t, err)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := secureClient(t, cert).Do(req)
	require.NoError(t, err)

	return resp
}
