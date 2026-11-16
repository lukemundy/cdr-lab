//go:build conformance

package register

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// The brands of the software products in services/register/seed/example.json.
const (
	idpTestBrand    = "40c7892c-82c7-4452-a97d-319000cfa828"
	statusTestBrand = "c6081b2a-4335-4b6b-92bd-0e6b11f9a1e6"
	mockBudgetBrand = "e25806f4-9384-41f6-a65d-2b12e2529e3b"
)

// mockBudgetWeb is an ACTIVE software product of services/register/seed/example.json.
const mockBudgetWeb = "66b83c54-e4f4-434d-aa9b-789f64358022"

const registerJWKSPath = "/cdr-register/v1/jwks"

// The SSA body is a JWT with no JSON quotes, so the 200 response cannot be
// validated against the string schema of the OpenAPI specification.
var validatedSSAStatuses = []int{
	http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusNotAcceptable,
}

func ssaPath(industry, brand, product string) string {
	return "/cdr-register/v1/" + industry + "/data-recipients/brands/" + brand +
		"/software-products/" + product + "/ssa"
}

func bearerV4(token string) http.Header {
	return http.Header{"Authorization": {"Bearer " + token}, "X-V": {"4"}}
}

func jwksAt(t *testing.T, url string) *jose.JSONWebKeySet {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err := labtest.Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var set jose.JSONWebKeySet
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&set))

	return &set
}

// Verifies: REG-0400, REG-0413
func TestADataRecipientGetsAnSSAForItsSoftwareProductThatVerifiesWithGetJWKS(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)
	cfg, _ := openIDConfiguration(t)

	before := time.Now().Add(-2 * time.Second)
	r := secureRequest(
		t,
		cert,
		ssaPath("all", idpTestBrand, idpTestActiveProduct),
		bearerV4(token),
		validatedSSAStatuses,
	)
	second := secureRequest(
		t,
		cert,
		ssaPath("all", idpTestBrand, idpTestActiveProduct),
		bearerV4(token),
		validatedSSAStatuses,
	)

	require.Equal(t, http.StatusOK, r.status, string(r.body))
	assert.Equal(t, "application/json", r.header.Get("Content-Type"))
	assert.Equal(t, "4", r.header.Get("x-v"))
	assert.NotEmpty(t, r.header.Get("x-fapi-interaction-id"))
	parsed, err := jwt.ParseSigned(string(r.body), []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(t, err, "the body is a PS256 JWT with no JSON quotes")
	require.Len(t, parsed.Headers, 1)
	kid := parsed.Headers[0].KeyID
	assert.Equal(t, "JWT", parsed.Headers[0].ExtraHeaders[jose.HeaderType])

	ssaKeys := jwksAt(t, labtest.Env(t, "REGISTER_PUBLIC_URL")+registerJWKSPath).Key(kid)
	require.Len(t, ssaKeys, 1, "the SSA signing key is at Get JWKS")
	assert.Empty(t, jwksAt(t, cfg.JWKSURI).Key(kid), "the SSA signing key is not at the jwks_uri of the access tokens")

	var claims map[string]any
	require.NoError(t, parsed.Claims(ssaKeys[0].Key, &claims), "the signature verifies with the key from Get JWKS")

	iat, exp := int64(claims["iat"].(float64)), int64(claims["exp"].(float64))
	assert.Equal(t, int64(600), exp-iat, "the SSA expires after 10 minutes")
	assert.WithinDuration(t, time.Now(), time.Unix(iat, 0), 10*time.Second)
	assert.False(t, time.Unix(iat, 0).Before(before))

	jti := claims["jti"]
	assert.NotEmpty(t, jti)
	delete(claims, "iat")
	delete(claims, "exp")
	delete(claims, "jti")
	assert.Equal(t, map[string]any{
		"iss":                "cdr-register",
		"legal_entity_id":    idpTestRecipient,
		"legal_entity_name":  "Token Test Pty Ltd",
		"org_id":             idpTestBrand,
		"org_name":           "Token Test",
		"client_name":        "Token Test Active",
		"client_description": "Token Test Active, a software product of the conformance tests",
		"client_uri":         "https://conformance.example",
		"redirect_uris":      []any{"https://conformance.example/callback"},
		"logo_uri":           "https://conformance.example/logo.png",
		"jwks_uri":           "https://conformance:8443/jwks",
		"revocation_uri":     "https://conformance.example/revoke",
		"recipient_base_uri": "https://cdr.conformance.example",
		"software_id":        idpTestActiveProduct,
		"software_roles":     "data-recipient-software-product",
		"scope":              "openid bank:accounts.basic:read cdr:registration",
	}, claims)

	require.Equal(t, http.StatusOK, second.status)
	secondParsed, err := jwt.ParseSigned(string(second.body), []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(t, err)

	var secondClaims map[string]any
	require.NoError(t, secondParsed.Claims(ssaKeys[0].Key, &secondClaims))
	assert.NotEqual(t, jti, secondClaims["jti"], "each SSA has a different jti")
}

// Verifies: REG-0401, REG-0326, REG-0466
func TestGetJWKSHasTheRequiredFieldsAndAnETag(t *testing.T) {
	resp := call(t, registerJWKSPath, http.Header{})
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
	etag := resp.Header.Get("ETag")

	again := call(t, registerJWKSPath, http.Header{"If-None-Match": {etag}})

	assert.Regexp(t, `^"[^"]+"$`, etag, "the ETag is a quoted entity tag (RFC 9110)")
	assert.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))

	var body struct {
		Keys []map[string]any `json:"keys"`
	}
	require.NoError(t, json.Unmarshal(resp.body, &body))
	require.NotEmpty(t, body.Keys)

	for _, key := range body.Keys {
		for _, field := range []string{"alg", "e", "key_ops", "kid", "kty", "n"} {
			assert.Contains(t, key, field)
		}

		assert.Equal(t, "RSA", key["kty"])
		assert.Equal(t, "PS256", key["alg"])
		assert.NotContains(t, key, "d", "the JWKS has no private key")
		n, err := base64.RawURLEncoding.DecodeString(key["n"].(string))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, new(big.Int).SetBytes(n).BitLen(), 2048)
	}

	assert.Equal(t, http.StatusNotModified, again.StatusCode)
	assert.Empty(t, again.body)
}

// Verifies: REG-0403, REG-0404
func TestGetSSARefusesABrandOrSoftwareProductThatIsNotOfTheCaller(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)

	tests := map[string]struct {
		brand, product string
		code, title    string
	}{
		"another software product": {
			idpTestBrand,
			mockBudgetWeb,
			"urn:au-cds:error:cds-register:Field/InvalidSoftwareProduct",
			"Invalid Software Product",
		},
		"another brand": {
			mockBudgetBrand,
			idpTestActiveProduct,
			"urn:au-cds:error:cds-register:Field/InvalidBrand",
			"Invalid Brand",
		},
		"a software product and a brand of another data recipient": {
			mockBudgetBrand,
			mockBudgetWeb,
			"urn:au-cds:error:cds-register:Field/InvalidSoftwareProduct",
			"Invalid Software Product",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := secureRequest(t, cert, ssaPath("all", tt.brand, tt.product), bearerV4(token), validatedSSAStatuses)

			requireSecureCDSError(t, r, http.StatusNotFound, tt.code, tt.title)
		})
	}
}

func TestGetSSAChecksTheTokenTheVersionAndTheIndustry(t *testing.T) {
	cert := activeTokenClient(t)
	token := accessToken(t, idpTestActiveProduct, cert)
	path := ssaPath("all", idpTestBrand, idpTestActiveProduct)

	noToken := secureRequest(t, cert, path, http.Header{"X-V": {"4"}}, validatedSSAStatuses)
	oldVersion := secureRequest(
		t,
		cert,
		path,
		http.Header{"Authorization": {"Bearer " + token}, "X-V": {"3"}},
		validatedSSAStatuses,
	)
	industry := secureRequest(
		t,
		cert,
		ssaPath("banking", idpTestBrand, idpTestActiveProduct),
		bearerV4(token),
		validatedSSAStatuses,
	)

	assert.Equal(t, http.StatusUnauthorized, noToken.status)
	assert.Equal(t, "Bearer", noToken.header.Get("WWW-Authenticate"))
	requireSecureCDSError(
		t,
		oldVersion,
		http.StatusNotAcceptable,
		"urn:au-cds:error:cds-all:Header/UnsupportedVersion",
		"Unsupported Version",
	)
	requireSecureCDSError(t, industry, http.StatusBadRequest, "urn:au-cds:error:cds-all:Field/Invalid", "Invalid Field")
}

// Verifies: REG-0402
func TestGetSSARefusesACallerThatIsNotActive(t *testing.T) {
	cert := tokenClient(t, statusTestRecipient, statusTestProduct, "PKI_STATUS_CLIENT_CERT", "PKI_STATUS_CLIENT_KEY")
	token := accessToken(t, statusTestProduct, cert)
	get := func() secureReply {
		return secureRequest(
			t,
			cert,
			ssaPath("all", statusTestBrand, statusTestProduct),
			bearerV4(token),
			validatedSSAStatuses,
		)
	}
	operator := func(path string) {
		t.Helper()
		code, p := operate(t, path)
		require.Equal(t, http.StatusNoContent, code, p.Detail)
	}

	require.Equal(t, http.StatusOK, get().status)

	operator(productAction(statusTestRecipient, statusTestProduct, "deactivate"))

	var refused secureReply

	require.Eventually(
		t,
		func() bool { refused = get(); return refused.status == http.StatusForbidden },
		propagation,
		50*time.Millisecond,
	)
	// The test leaves the software product ACTIVE, so that it can run again.
	operator(productAction(statusTestRecipient, statusTestProduct, "activate"))

	requireSecureCDSError(
		t,
		refused,
		http.StatusForbidden,
		"urn:au-cds:error:cds-all:Authorisation/AdrStatusNotActive",
		"ADR Status Is Not Active",
	)
	require.Eventually(t, func() bool { return get().status == http.StatusOK }, propagation, 50*time.Millisecond)
}
