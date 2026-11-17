//go:build conformance

package dataholder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// These participants of services/register/seed/example.json have https://conformance:8444/jwks as their JWKS URI.
const (
	registrationTestBrand = "b2c4e6a8-1d3f-4a5b-8c7d-9e0f1a2b3c4d"
	// minimalProduct has the scope bank:payees:read, which the data holder does not support.
	minimalProduct = "e5a7b9c1-3f5d-4d7e-8f9a-1b2c3d4e5f60"
	// dataOnlyProduct does not have the scope cdr:registration.
	dataOnlyProduct = "a7c9d1e3-5b7f-4f9a-8b1c-3d4e5f6a7b82"
	// noConsentProduct registers with no grant type for consent.
	noConsentProduct = "c9e1f3a5-7d9b-4b1c-8d3e-5f6a7b8c9da4"
)

const (
	jwtMediaType = "application/jwt"
	// The pinned specification has the registration endpoint at this path.
	specRegistrationPath = "/register"
)

var dcrSpec = sync.OnceValues(LoadDCRSpec)

// registrant is a software product that registers with the data holder.
type registrant struct {
	productID string
	cert      tls.Certificate
	// ssa is a new SSA of the software product from the Register.
	ssa      string
	endpoint string
	issuer   string
}

func newRegistrant(t *testing.T, productID string) registrant {
	t.Helper()

	cert := softwareProductOf(t, productID)
	token := labtest.RegisterAccessToken(t, recipientKeys(t), productID, cert)
	brand := labtest.RegisteredDataHolderBrand(t, token, cert)
	configuration := openIDConfiguration(t, brand.EndpointDetail.InfosecBaseURI)

	require.IsType(t, "", configuration["registration_endpoint"])
	require.IsType(t, "", configuration["issuer"])

	return registrant{
		productID: productID,
		cert:      cert,
		ssa:       labtest.SoftwareStatement(t, token, cert, registrationTestBrand, productID),
		endpoint:  configuration["registration_endpoint"].(string),
		issuer:    configuration["issuer"].(string),
	}
}

func (r registrant) claims() map[string]any {
	return labtest.RegistrationClaims(r.productID, r.issuer, r.ssa)
}

type registrationReply struct {
	status int
	header http.Header
	body   map[string]any
}

// post validates each response with a body against the pinned specification.
func (r registrant) post(t *testing.T, contentType, body string) registrationReply {
	t.Helper()

	resp := labtest.PostRegistration(t, r.endpoint, r.cert, contentType, body)

	defer func() { _ = resp.Body.Close() }()

	reply := registrationReply{status: resp.StatusCode, header: resp.Header}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusBadRequest {
		return reply
	}

	spec, err := dcrSpec()
	require.NoError(t, err)
	require.NoError(t, validateAtSpecPath(spec.ValidateResponse, resp, specRegistrationPath))

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &reply.body), string(data))

	return reply
}

// validateAtSpecPath is necessary because a data holder can have its registration endpoint at a different path.
func validateAtSpecPath(validate func(*http.Response) error, resp *http.Response, specPath string) error {
	original := resp.Request
	atSpecPath := original.Clone(original.Context())
	atSpecPath.URL.Path = specPath
	resp.Request = atSpecPath

	defer func() { resp.Request = original }()

	return validate(resp)
}

func (r registrant) register(t *testing.T, claims map[string]any) registrationReply {
	t.Helper()

	return r.post(t, jwtMediaType, labtest.SignJWT(t, jose.PS256, recipientKeys(t).RSA, "rsa", claims))
}

// registeredClient is the one registration of a software product, which all tests of the package share: a software
// product registers one time, and the data holder has no way to delete a registration.
type registeredClient struct {
	registrant

	clientID string
	// reply is the response to the registration.
	reply registrationReply
	// requested is the time before the registration request.
	requested time.Time
}

var registeredClients struct {
	mu      sync.Mutex
	clients map[string]registeredClient
}

func registered(t *testing.T, productID string) registeredClient {
	t.Helper()

	registeredClients.mu.Lock()
	defer registeredClients.mu.Unlock()

	if client, ok := registeredClients.clients[productID]; ok {
		return client
	}

	r := newRegistrant(t, productID)
	client := registeredClient{registrant: r, requested: time.Now()}
	client.reply = r.register(t, registrationClaimsOf(r))
	require.Equal(t, http.StatusCreated, client.reply.status, client.reply.body)
	require.IsType(t, "", client.reply.body["client_id"])
	client.clientID = client.reply.body["client_id"].(string)

	if registeredClients.clients == nil {
		registeredClients.clients = map[string]registeredClient{}
	}

	registeredClients.clients[productID] = client

	return client
}

// registrationClaimsOf gives minimalProduct a request with only the claims that the CDS requires, and with claims that
// the data holder does not support.
func registrationClaimsOf(r registrant) map[string]any {
	claims := r.claims()

	if r.productID == noConsentProduct {
		claims["grant_types"] = []string{"client_credentials"}
		return claims
	}

	if r.productID != minimalProduct {
		return claims
	}

	delete(claims, "redirect_uris")
	delete(claims, "request_object_signing_alg")
	delete(claims, "application_type")
	claims["response_types"] = []string{"code", "code id_token"}
	claims["id_token_encrypted_response_alg"] = "RSA-OAEP"
	claims["id_token_encrypted_response_enc"] = "A256GCM"
	claims["authorization_encrypted_response_alg"] = "RSA-OAEP"
	claims["authorization_encrypted_response_enc"] = "A128CBC-HS256"
	claims["a_claim_that_the_cds_does_not_have"] = "value"

	return claims
}

func requireRegistrationError(t *testing.T, reply registrationReply, code string) {
	t.Helper()

	require.Equal(t, http.StatusBadRequest, reply.status, reply.body)
	assert.Equal(t, code, reply.body["error"], reply.body)
	assert.NotEmpty(t, reply.body["error_description"])
	assert.Contains(t, reply.header.Get("Content-Type"), "application/json")
}

func ssaClaims(t *testing.T, ssa string) map[string]any {
	t.Helper()

	parsed, err := jwt.ParseSigned(ssa, []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(t, err)

	var claims map[string]any
	require.NoError(t, parsed.UnsafeClaimsWithoutVerification(&claims))

	return claims
}

func TestASoftwareProductRegistersOneTimeWithAnSSAFromTheRegister(t *testing.T) {
	client := registered(t, registrationTestProduct)
	r, reply := client.registrant, client.reply

	assert.Equal(t, "no-store", reply.header.Get("Cache-Control"))

	t.Run("the data holder issues a new client ID", func(t *testing.T) {
		require.IsType(t, "", reply.body["client_id"])
		_, err := uuid.Parse(reply.body["client_id"].(string))
		require.NoError(t, err)
		assert.NotEqual(t, registrationTestProduct, reply.body["client_id"])

		require.IsType(t, float64(0), reply.body["client_id_issued_at"])
		issuedAt := time.Unix(int64(reply.body["client_id_issued_at"].(float64)), 0)
		assert.WithinRange(t, issuedAt, client.requested.Add(-2*time.Second), time.Now().Add(2*time.Second))
	})

	t.Run("the response has the values of the SSA as top-level members", func(t *testing.T) {
		ssa := ssaClaims(t, r.ssa)
		for _, claim := range []string{
			"client_name", "client_description", "client_uri", "legal_entity_id", "legal_entity_name", "org_id",
			"org_name", "redirect_uris", "logo_uri", "jwks_uri", "revocation_uri", "recipient_base_uri",
			"software_id", "software_roles", "scope",
		} {
			assert.Equal(t, ssa[claim], reply.body[claim], claim)
		}

		assert.Equal(t, registrationTestProduct, reply.body["software_id"])
		assert.Equal(t, r.ssa, reply.body["software_statement"], "the SSA is not changed")
	})

	t.Run("the response has the client metadata of the request", func(t *testing.T) {
		//nolint:gosec // G101 reads the protocol field token_endpoint_auth_method as a credential
		for claim, value := range map[string]any{
			"token_endpoint_auth_method":        "private_key_jwt",
			"token_endpoint_auth_signing_alg":   "PS256",
			"grant_types":                       []any{"client_credentials", "authorization_code", "refresh_token"},
			"response_types":                    []any{"code"},
			"application_type":                  "web",
			"id_token_signed_response_alg":      "PS256",
			"authorization_signed_response_alg": "PS256",
			"request_object_signing_alg":        "PS256",
		} {
			assert.Equal(t, value, reply.body[claim], claim)
		}
	})

	t.Run("the response has no member of RFC 7592", func(t *testing.T) {
		assert.NotContains(t, reply.body, "registration_access_token")
		assert.NotContains(t, reply.body, "registration_client_uri")
	})

	// The next requests have a new SSA. The SSA of the registration can be from a key that the Register removed,
	// because the Register tests change the SSA key two times while these tests run.
	t.Run("a second registration of the software product is refused", func(t *testing.T) {
		again := newRegistrant(t, registrationTestProduct)

		requireRegistrationError(t, again.register(t, again.claims()), "invalid_software_statement")
	})

	t.Run("a request JWT that the data holder accepted before is refused", func(t *testing.T) {
		again := newRegistrant(t, registrationTestProduct)
		request := labtest.SignJWT(t, jose.PS256, recipientKeys(t).RSA, "rsa", again.claims())
		// The data holder accepts this request JWT, and then refuses the second registration.
		requireRegistrationError(t, again.post(t, jwtMediaType, request), "invalid_software_statement")

		requireRegistrationError(t, again.post(t, jwtMediaType, request), "invalid_client_metadata")
	})
}

func TestARegistrationGetsTheDefaultsAndOnlyWhatTheDataHolderSupports(t *testing.T) {
	client := registered(t, minimalProduct)
	r, reply := client.registrant, client.reply

	ssa := ssaClaims(t, r.ssa)
	assert.Contains(t, ssa["scope"], "bank:payees:read", "the SSA has a scope that the data holder does not support")
	assert.Equal(t, "openid profile bank:accounts.basic:read cdr:registration", reply.body["scope"])

	assert.Equal(t, ssa["redirect_uris"], reply.body["redirect_uris"])
	assert.Equal(t, "PS256", reply.body["request_object_signing_alg"])
	assert.Equal(t, "web", reply.body["application_type"])
	assert.Equal(t, []any{"code"}, reply.body["response_types"])

	for _, claim := range []string{
		"id_token_encrypted_response_alg", "id_token_encrypted_response_enc",
		"authorization_encrypted_response_alg", "authorization_encrypted_response_enc",
		"a_claim_that_the_cds_does_not_have",
	} {
		assert.NotContains(t, reply.body, claim)
	}
}

func TestARegistrationWithAnSSAThatTheRegisterDidNotSignIsRefused(t *testing.T) {
	r := newRegistrant(t, registrationTestProduct)
	keys := recipientKeys(t)

	header, rest, _ := strings.Cut(r.ssa, ".")
	payload, _, _ := strings.Cut(rest, ".")

	tests := map[string]string{
		"a signature that is not valid": header + "." + payload + "." + strings.Repeat("A", 342),
		"a key ID that is not in the Register JWKS": labtest.SignJWT(
			t, jose.PS256, keys.RSA, "not-a-register-key", ssaClaims(t, r.ssa),
		),
		"not a JWT":             "not-an-ssa",
		"no software statement": "",
	}
	for name, ssa := range tests {
		t.Run(name, func(t *testing.T) {
			claims := r.claims()
			claims["software_statement"] = ssa

			requireRegistrationError(t, r.register(t, claims), "invalid_software_statement")
		})
	}
}

func TestARegistrationWithARedirectURIThatTheSSADoesNotHaveIsRefused(t *testing.T) {
	r := newRegistrant(t, registrationTestProduct)
	claims := r.claims()
	claims["redirect_uris"] = []string{"https://conformance.example/callback", "https://other.example/callback"}

	requireRegistrationError(t, r.register(t, claims), "invalid_redirect_uri")
}

func TestARegistrationRequestJWTThatIsNotValidIsRefused(t *testing.T) {
	r := newRegistrant(t, registrationTestProduct)

	tests := map[string]func(claims map[string]any){
		"an aud that is not the issuer":      func(c map[string]any) { c["aud"] = "https://other.example" },
		"an iss that is not the software_id": func(c map[string]any) { c["iss"] = minimalProduct },
		"an exp in the past":                 func(c map[string]any) { c["exp"] = time.Now().Add(-time.Minute).Unix() },
		"an exp of more than 10 minutes": func(c map[string]any) {
			c["exp"] = time.Now().Add(11 * time.Minute).Unix()
		},
		"no jti": func(c map[string]any) { delete(c, "jti") },
		"an authentication method that is not private_key_jwt": func(c map[string]any) {
			c["token_endpoint_auth_method"] = "client_secret_basic"
		},
		"a token algorithm that is not PS256 or ES256": func(c map[string]any) {
			c["token_endpoint_auth_signing_alg"] = "RS256"
		},
		"an ID token algorithm that the data holder does not sign with": func(c map[string]any) {
			c["id_token_signed_response_alg"] = "ES256"
		},
		"no authorisation response algorithm": func(c map[string]any) {
			delete(c, "authorization_signed_response_alg")
		},
		"a grant type that is not supported": func(c map[string]any) { c["grant_types"] = []string{"implicit"} },
		"no response type code": func(c map[string]any) {
			c["response_types"] = []string{"code id_token"}
		},
		"an application type that is not web": func(c map[string]any) { c["application_type"] = "native" },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			claims := r.claims()
			change(claims)

			requireRegistrationError(t, r.register(t, claims), "invalid_client_metadata")
		})
	}

	t.Run("a signature from a key that is not at the jwks_uri of the SSA", func(t *testing.T) {
		otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		reply := r.post(t, jwtMediaType, labtest.SignJWT(t, jose.PS256, otherKey, "rsa", r.claims()))

		requireRegistrationError(t, reply, "invalid_client_metadata")
	})

	t.Run("a body that is not a JWT", func(t *testing.T) {
		requireRegistrationError(
			t,
			r.post(t, jwtMediaType, `{"software_statement": "a.b.c"}`),
			"invalid_client_metadata",
		)
	})
}

func TestARegistrationWithAnES256RequestJWTIsNotRefusedForItsSignature(t *testing.T) {
	r := newRegistrant(t, registrationTestProduct)
	claims := r.claims()
	claims["redirect_uris"] = []string{"https://other.example/callback"}

	reply := r.post(t, jwtMediaType, labtest.SignJWT(t, jose.ES256, recipientKeys(t).EC, "ec", claims))

	requireRegistrationError(t, reply, "invalid_redirect_uri")
}

func TestARegistrationWithAContentTypeThatIsNotAJWTGives415(t *testing.T) {
	r := newRegistrant(t, registrationTestProduct)
	request := labtest.SignJWT(t, jose.PS256, recipientKeys(t).RSA, "rsa", r.claims())

	for _, contentType := range []string{"application/json", "text/plain"} {
		t.Run(contentType, func(t *testing.T) {
			reply := r.post(t, contentType, request)

			assert.Equal(t, http.StatusUnsupportedMediaType, reply.status)
		})
	}
}

func TestTheRegistrationEndpointHasAnInteractionIDAndNoCORSHeaders(t *testing.T) {
	r := newRegistrant(t, registrationTestProduct)

	reply := r.register(t, r.claims())

	assert.Empty(t, reply.header.Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, reply.header.Get("x-fapi-interaction-id"))
}
