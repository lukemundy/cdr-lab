//go:build conformance

package dataholder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

const (
	requestURIPrefix = "urn:ietf:params:oauth:request_uri:"
	// The redirect URI of the Registration Test software products in services/register/seed/example.json.
	registeredRedirectURI = "https://conformance.example/callback"
)

func parEndpoint(t *testing.T, client registeredClient) string {
	t.Helper()

	endpoint := openIDConfiguration(t, client.issuer)["pushed_authorization_request_endpoint"]
	require.IsType(t, "", endpoint)

	return endpoint.(string)
}

// requestObjectClaims gives the claims of a request object that the data holder accepts.
func (c registeredClient) requestObjectClaims() map[string]any {
	challenge := sha256.Sum256([]byte(rand.Text() + rand.Text()))
	now := time.Now()

	return map[string]any{
		"iss":                   c.clientID,
		"aud":                   c.issuer,
		"client_id":             c.clientID,
		"response_type":         "code",
		"response_mode":         "jwt",
		"redirect_uri":          registeredRedirectURI,
		"scope":                 "openid bank:accounts.basic:read",
		"state":                 rand.Text(),
		"nonce":                 rand.Text(),
		"code_challenge":        base64.RawURLEncoding.EncodeToString(challenge[:]),
		"code_challenge_method": "S256",
		"nbf":                   now.Unix(),
		"exp":                   now.Add(5 * time.Minute).Unix(),
	}
}

// parForm gives the form of a PAR request: a client assertion for the PAR endpoint and the signed request object.
func (c registeredClient) parForm(t *testing.T, requestObject string) url.Values {
	t.Helper()

	form := labtest.TokenForm(t, c.clientID, "", c.assertion(t, parEndpoint(t, c)))
	form.Del("grant_type")
	form.Del("scope")
	form.Set("request", requestObject)

	return form
}

func (c registeredClient) signRequestObject(t *testing.T, claims map[string]any) string {
	t.Helper()

	return labtest.SignJWT(t, jose.PS256, recipientKeys(t).RSA, "rsa", claims)
}

// push sends the claims as a request object that the software product signs.
func (c registeredClient) push(t *testing.T, claims map[string]any) labtest.TokenReply {
	t.Helper()

	return labtest.RequestToken(t, parEndpoint(t, c), c.cert, c.parForm(t, c.signRequestObject(t, claims)))
}

// requirePushRefused checks the RFC 9126 section 2.3 format, and gives the description.
func requirePushRefused(t *testing.T, r labtest.TokenReply, status int, code string) string {
	t.Helper()
	require.Equal(t, status, r.Status, r.Body)
	assert.Equal(t, code, r.Body["error"], r.Body)
	assert.Contains(t, r.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, "no-store", r.Header.Get("Cache-Control"))
	assert.NotContains(t, r.Body, "request_uri")

	description, _ := r.Body["error_description"].(string)
	assert.NotEmpty(t, description)

	return description
}

func TestARegisteredSoftwareProductPushesARequestObjectAndGetsARequestURIFor60Seconds(t *testing.T) {
	client := registered(t, registrationTestProduct)

	first := client.push(t, client.requestObjectClaims())
	second := client.push(t, client.requestObjectClaims())

	require.Equal(t, http.StatusCreated, first.Status, first.Body)
	require.Equal(t, http.StatusCreated, second.Status, second.Body)
	assert.Contains(t, first.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, "no-store", first.Header.Get("Cache-Control"))
	assert.InDelta(t, 60, first.Body["expires_in"], 0)

	requestURI, _ := first.Body["request_uri"].(string)
	assert.True(t, strings.HasPrefix(requestURI, requestURIPrefix), requestURI)
	assert.NotEqual(t, first.Body["request_uri"], second.Body["request_uri"])
}

func TestAPushedRequestWithASharingDurationAndAnACRThatTheDataHolderReachesIsAccepted(t *testing.T) {
	client := registered(t, registrationTestProduct)
	claims := client.requestObjectClaims()
	claims["claims"] = map[string]any{
		"sharing_duration": 7776000,
		"id_token": map[string]any{
			"acr": map[string]any{"essential": true, "values": []string{"urn:cds.au:cdr:2"}},
		},
	}

	r := client.push(t, claims)

	require.Equal(t, http.StatusCreated, r.Status, r.Body)
}

func TestAnAuthorisationParameterOutsideTheRequestObjectHasNoEffect(t *testing.T) {
	client := registered(t, registrationTestProduct)
	form := client.parForm(t, client.signRequestObject(t, client.requestObjectClaims()))
	form.Set("redirect_uri", "https://attacker.example/callback")
	form.Set("scope", "openid bank:transactions:read")
	form.Set("response_type", "token")

	r := labtest.RequestToken(t, parEndpoint(t, client), client.cert, form)

	require.Equal(t, http.StatusCreated, r.Status, r.Body)
}

func TestAPushedRequestWithARequestURIParameterIsRefused(t *testing.T) {
	client := registered(t, registrationTestProduct)
	form := client.parForm(t, client.signRequestObject(t, client.requestObjectClaims()))
	form.Set("request_uri", requestURIPrefix+"abc")

	r := labtest.RequestToken(t, parEndpoint(t, client), client.cert, form)

	requirePushRefused(t, r, http.StatusBadRequest, "invalid_request")
}

func TestAPushedRequestWithAClientAssertionThatIsNotValidIsRefused(t *testing.T) {
	client := registered(t, registrationTestProduct)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	wrongKey := client.assertion(t, parEndpoint(t, client))
	wrongKey.Key = otherKey

	form := client.parForm(t, client.signRequestObject(t, client.requestObjectClaims()))
	form.Set("client_assertion", labtest.TokenForm(t, client.clientID, "", wrongKey).Get("client_assertion"))

	r := labtest.RequestToken(t, parEndpoint(t, client), client.cert, form)

	requirePushRefused(t, r, http.StatusUnauthorized, "invalid_client")
}

func TestASoftwareProductWithNoAuthorisationCodeGrantCannotPushARequest(t *testing.T) {
	client := registered(t, noConsentProduct)

	r := client.push(t, client.requestObjectClaims())

	requirePushRefused(t, r, http.StatusBadRequest, "invalid_grant")
}

func TestARequestObjectThatBreaksARuleIsRefusedWithADescriptionThatNamesTheRule(t *testing.T) {
	client := registered(t, registrationTestProduct)
	now := time.Now()

	tests := map[string]struct {
		change map[string]any
		remove string
		code   string
		// names is a part of the error_description.
		names string
	}{
		"wrong aud": {
			change: map[string]any{"aud": "https://conformance.example"},
			code:   "invalid_request_object", names: "aud",
		},
		"wrong iss": {
			change: map[string]any{"iss": registrationTestProduct},
			code:   "invalid_request_object", names: "iss",
		},
		"no nbf": {remove: "nbf", code: "invalid_request_object", names: "nbf"},
		"no exp": {remove: "exp", code: "invalid_request_object", names: "exp"},
		"nbf of more than 60 minutes before the request": {
			change: map[string]any{"nbf": now.Add(-61 * time.Minute).Unix()},
			code:   "invalid_request_object", names: "nbf",
		},
		"exp of more than 60 minutes after the nbf": {
			change: map[string]any{"nbf": now.Add(-30 * time.Minute).Unix(), "exp": now.Add(31 * time.Minute).Unix()},
			code:   "invalid_request_object", names: "60 minutes after nbf",
		},
		"response_type that is not code": {
			change: map[string]any{"response_type": "code id_token"},
			code:   "invalid_request", names: "response_type",
		},
		"response_mode that is not jwt": {
			change: map[string]any{"response_mode": "query"},
			code:   "invalid_request", names: "response_mode",
		},
		"redirect_uri that the registration does not have": {
			change: map[string]any{"redirect_uri": "https://attacker.example/callback"},
			code:   "invalid_request", names: "redirect_uri",
		},
		"no nonce":          {remove: "nonce", code: "invalid_request", names: "nonce"},
		"no code_challenge": {remove: "code_challenge", code: "invalid_request", names: "code_challenge"},
		"code_challenge_method that is not S256": {
			change: map[string]any{"code_challenge_method": "plain"},
			code:   "invalid_request", names: "code_challenge_method",
		},
		"no openid scope": {
			change: map[string]any{"scope": "bank:accounts.basic:read"},
			code:   "invalid_request", names: "openid",
		},
		"scope that the registration does not have": {
			change: map[string]any{"scope": "openid profile"},
			code:   "invalid_scope", names: "profile",
		},
		"registration scope": {
			change: map[string]any{"scope": "openid cdr:registration"},
			code:   "invalid_scope", names: "cdr:registration",
		},
		"negative sharing_duration": {
			change: map[string]any{"claims": map[string]any{"sharing_duration": -1}},
			code:   "invalid_request_object", names: "sharing_duration",
		},
		"cdr_arrangement_id": {
			change: map[string]any{"claims": map[string]any{
				"cdr_arrangement_id": "5a1bf696-ee03-408b-b315-97955415d1f0",
			}},
			code: "invalid_request_object", names: "cdr_arrangement_id",
		},
		"essential acr that the data holder cannot reach": {
			change: map[string]any{"claims": map[string]any{"id_token": map[string]any{
				"acr": map[string]any{"essential": true, "value": "urn:cds.au:cdr:3"},
			}}},
			code: "invalid_request_object", names: "acr",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			claims := client.requestObjectClaims()
			maps.Copy(claims, tt.change)
			delete(claims, tt.remove)

			description := requirePushRefused(t, client.push(t, claims), http.StatusBadRequest, tt.code)

			assert.Contains(t, description, tt.names)
		})
	}
}

func TestARequestObjectWithASignatureFromAKeyThatTheSoftwareProductDoesNotHaveIsRefused(t *testing.T) {
	client := registered(t, registrationTestProduct)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	requestObject := labtest.SignJWT(t, jose.PS256, otherKey, "rsa", client.requestObjectClaims())

	r := labtest.RequestToken(t, parEndpoint(t, client), client.cert, client.parForm(t, requestObject))

	requirePushRefused(t, r, http.StatusBadRequest, "invalid_request_object")
}
