package authserver_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const requestURIPrefix = "urn:ietf:params:oauth:request_uri:"

// requestObject gives the claims of a request object that the data holder accepts.
func (b *bank) requestObject() map[string]any {
	challenge := sha256.Sum256([]byte(rand.Text() + rand.Text()))
	now := time.Now()

	return map[string]any{
		"iss":                   registeredClientID,
		"aud":                   b.issuer,
		"client_id":             registeredClientID,
		"response_type":         "code",
		"response_mode":         "jwt",
		"redirect_uri":          registeredRedirectURI,
		"scope":                 "openid bank:accounts.basic:read",
		"state":                 "state-1",
		"nonce":                 "nonce-1",
		"code_challenge":        base64.RawURLEncoding.EncodeToString(challenge[:]),
		"code_challenge_method": "S256",
		"nbf":                   now.Unix(),
		"exp":                   now.Add(5 * time.Minute).Unix(),
	}
}

func (b *bank) parForm(object assertion) url.Values {
	b.t.Helper()

	clientAssertion := b.assertion()
	clientAssertion.claims["aud"] = b.parURL

	form := b.form(clientAssertion)
	form.Del("grant_type")
	form.Del("scope")
	form.Set("request", b.form(object).Get("client_assertion"))

	return form
}

func (b *bank) signedObject(claims map[string]any) assertion {
	return assertion{alg: jose.PS256, key: b.rsaKey, kid: "rsa", claims: claims}
}

func (b *bank) par(form url.Values) reply {
	b.t.Helper()

	req, err := http.NewRequestWithContext(b.t.Context(), http.MethodPost, b.parURL, strings.NewReader(form.Encode()))
	require.NoError(b.t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := b.http.Do(req)
	require.NoError(b.t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(b.t, err)

	// A response with no JSON body, for example 405, has an empty body here.
	var body map[string]any

	_ = json.Unmarshal(data, &body)

	return reply{status: resp.StatusCode, header: resp.Header, body: body}
}

// push sends the claims as a request object that the registered client signs.
func (b *bank) push(claims map[string]any) reply {
	b.t.Helper()
	return b.par(b.parForm(b.signedObject(claims)))
}

// requirePushRefused checks the RFC 9126 section 2.3 format, and gives the description.
func requirePushRefused(t *testing.T, r reply, status int, code string) string {
	t.Helper()
	require.Equal(t, status, r.status, r.body)
	assert.Equal(t, code, r.body["error"], r.body)
	assert.Equal(t, "application/json", r.header.Get("Content-Type"))
	assert.Equal(t, "no-store", r.header.Get("Cache-Control"))
	assert.NotContains(t, r.body, "request_uri")

	description, _ := r.body["error_description"].(string)

	return description
}

func TestARegisteredClientPushesARequestObjectAndGetsARequestURIFor60Seconds(t *testing.T) {
	b := newBank(t)

	r := b.push(b.requestObject())

	require.Equal(t, http.StatusCreated, r.status, r.body)
	assert.Equal(t, "application/json", r.header.Get("Content-Type"))
	assert.Equal(t, "no-store", r.header.Get("Cache-Control"))
	assert.InDelta(t, 60, r.body["expires_in"], 0)

	requestURI, _ := r.body["request_uri"].(string)
	assert.True(t, strings.HasPrefix(requestURI, requestURIPrefix), requestURI)
	assert.Greater(t, len(strings.TrimPrefix(requestURI, requestURIPrefix)), 20, "the reference is not short")
}

func TestTwoPushedRequestsGetDifferentRequestURIs(t *testing.T) {
	b := newBank(t)

	first, second := b.push(b.requestObject()), b.push(b.requestObject())

	require.Equal(t, http.StatusCreated, second.status, second.body)
	assert.NotEqual(t, first.body["request_uri"], second.body["request_uri"])
}

func TestARequestObjectSignedWithES256IsAcceptedWhenTheClientRegisteredES256(t *testing.T) {
	b := newBank(t)
	client := b.clients[registeredClientID]
	client.RequestObjectSigningAlg = "ES256"
	b.clients[registeredClientID] = client

	object := assertion{alg: jose.ES256, key: b.ecKey, kid: "ec", claims: b.requestObject()}

	r := b.par(b.parForm(object))

	require.Equal(t, http.StatusCreated, r.status, r.body)
}

func TestARequestObjectSignedWithAnAlgorithmThatTheClientDidNotRegisterIsRefused(t *testing.T) {
	b := newBank(t)
	object := assertion{alg: jose.ES256, key: b.ecKey, kid: "ec", claims: b.requestObject()}

	r := b.par(b.parForm(object))

	requirePushRefused(t, r, http.StatusBadRequest, "invalid_request_object")
}

func TestAnAuthorisationParameterOutsideTheRequestObjectHasNoEffect(t *testing.T) {
	b := newBank(t)
	form := b.parForm(b.signedObject(b.requestObject()))
	form.Set("redirect_uri", "https://attacker.example/callback")
	form.Set("scope", "openid bank:transactions:read")
	form.Set("response_type", "token")

	r := b.par(form)

	require.Equal(t, http.StatusCreated, r.status, r.body)
}

func TestAPushedRequestWithNoRequestObjectIsRefused(t *testing.T) {
	b := newBank(t)
	form := b.parForm(b.signedObject(b.requestObject()))
	form.Del("request")

	for name, value := range b.requestObject() {
		if text, ok := value.(string); ok {
			form.Set(name, text)
		}
	}

	requirePushRefused(t, b.par(form), http.StatusBadRequest, "invalid_request")
}

func TestAPushedRequestWithARequestURIParameterIsRefused(t *testing.T) {
	b := newBank(t)
	form := b.parForm(b.signedObject(b.requestObject()))
	form.Set("request_uri", requestURIPrefix+"abc")

	requirePushRefused(t, b.par(form), http.StatusBadRequest, "invalid_request")
}

func TestAPushedRequestWithAClientAssertionThatIsNotValidIsRefused(t *testing.T) {
	b := newBank(t)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	form := b.parForm(b.signedObject(b.requestObject()))
	wrongKey := b.assertion()
	wrongKey.key = otherKey
	form.Set("client_assertion", b.form(wrongKey).Get("client_assertion"))

	description := requirePushRefused(t, b.par(form), http.StatusUnauthorized, "invalid_client")

	assert.Equal(t, "invalid client", description, "the response does not give the cause of a failed authentication")
}

func TestAClientWithNoAuthorisationCodeGrantCannotPushARequest(t *testing.T) {
	b := newBank(t)
	client := b.clients[registeredClientID]
	client.GrantTypes = []string{"client_credentials"}
	b.clients[registeredClientID] = client

	requirePushRefused(t, b.push(b.requestObject()), http.StatusBadRequest, "invalid_grant")
}

func TestARequestObjectThatBreaksARuleIsRefusedWithACauseThatNamesTheRule(t *testing.T) {
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	now := time.Now()

	tests := map[string]struct {
		change func(b *bank, claims map[string]any)
		code   string
		// cause is a part of the error_description.
		cause string
	}{
		"wrong aud": {
			change: func(_ *bank, c map[string]any) { c["aud"] = "https://recipient.example" },
			code:   "invalid_request_object", cause: "aud",
		},
		"wrong iss": {
			change: func(_ *bank, c map[string]any) { c["iss"] = "client-2" },
			code:   "invalid_request_object", cause: "iss",
		},
		"client_id of a different client": {
			change: func(_ *bank, c map[string]any) { c["client_id"] = "client-2" },
			code:   "invalid_request_object", cause: "client_id",
		},
		"no nbf": {
			change: func(_ *bank, c map[string]any) { delete(c, "nbf") },
			code:   "invalid_request_object", cause: "nbf",
		},
		"no exp": {
			change: func(_ *bank, c map[string]any) { delete(c, "exp") },
			code:   "invalid_request_object", cause: "exp",
		},
		"nbf of more than 60 minutes before the request": {
			change: func(_ *bank, c map[string]any) { c["nbf"] = now.Add(-61 * time.Minute).Unix() },
			code:   "invalid_request_object", cause: "nbf",
		},
		"exp of more than 60 minutes after the nbf": {
			change: func(_ *bank, c map[string]any) {
				c["nbf"], c["exp"] = now.Add(-30*time.Minute).Unix(), now.Add(31*time.Minute).Unix()
			},
			code: "invalid_request_object", cause: "60 minutes after nbf",
		},
		"expired": {
			change: func(_ *bank, c map[string]any) {
				c["nbf"], c["exp"] = now.Add(-10*time.Minute).Unix(), now.Add(-5*time.Minute).Unix()
			},
			code: "invalid_request_object", cause: "exp",
		},
		"response_type that is not code": {
			change: func(_ *bank, c map[string]any) { c["response_type"] = "code id_token" },
			code:   "invalid_request",
		},
		"response_mode that is not jwt": {
			change: func(_ *bank, c map[string]any) { c["response_mode"] = "query" },
			code:   "invalid_request",
		},
		"no response_mode": {
			change: func(_ *bank, c map[string]any) { delete(c, "response_mode") },
			code:   "invalid_request", cause: "response_mode",
		},
		"redirect_uri that is not registered": {
			change: func(_ *bank, c map[string]any) { c["redirect_uri"] = "https://attacker.example/callback" },
			code:   "invalid_request", cause: "redirect_uri",
		},
		"no redirect_uri": {
			change: func(_ *bank, c map[string]any) { delete(c, "redirect_uri") },
			code:   "invalid_request", cause: "redirect_uri",
		},
		"no nonce": {
			change: func(_ *bank, c map[string]any) { delete(c, "nonce") },
			code:   "invalid_request", cause: "nonce",
		},
		"no code_challenge": {
			change: func(_ *bank, c map[string]any) { delete(c, "code_challenge") },
			code:   "invalid_request", cause: "code_challenge",
		},
		"code_challenge_method that is not S256": {
			change: func(_ *bank, c map[string]any) { c["code_challenge_method"] = "plain" },
			code:   "invalid_request", cause: "code_challenge_method",
		},
		"no openid scope": {
			change: func(_ *bank, c map[string]any) { c["scope"] = "bank:accounts.basic:read" },
			code:   "invalid_request", cause: "openid",
		},
		"scope that the registration does not have": {
			change: func(_ *bank, c map[string]any) { c["scope"] = "openid profile" },
			code:   "invalid_scope", cause: "profile",
		},
		"scope that the data holder does not support": {
			change: func(_ *bank, c map[string]any) { c["scope"] = "openid bank:payees:read" },
			code:   "invalid_scope", cause: "bank:payees:read",
		},
		"registration scope": {
			change: func(_ *bank, c map[string]any) { c["scope"] = "openid cdr:registration" },
			code:   "invalid_scope", cause: "cdr:registration",
		},
		"negative sharing_duration": {
			change: func(_ *bank, c map[string]any) { c["claims"] = map[string]any{"sharing_duration": -1} },
			code:   "invalid_request_object", cause: "sharing_duration",
		},
		"cdr_arrangement_id": {
			change: func(_ *bank, c map[string]any) {
				c["claims"] = map[string]any{"cdr_arrangement_id": "5a1bf696-ee03-408b-b315-97955415d1f0"}
			},
			code: "invalid_request_object", cause: "cdr_arrangement_id",
		},
		"essential acr that the data holder cannot reach": {
			change: func(_ *bank, c map[string]any) {
				c["claims"] = map[string]any{"id_token": map[string]any{
					"acr": map[string]any{"essential": true, "value": "urn:cds.au:cdr:3"},
				}}
			},
			code: "invalid_request_object", cause: "acr",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := newBank(t)
			claims := b.requestObject()
			tt.change(b, claims)

			description := requirePushRefused(t, b.push(claims), http.StatusBadRequest, tt.code)

			assert.Contains(t, description, tt.cause)
		})
	}

	t.Run("signature of a key that the client does not have", func(t *testing.T) {
		b := newBank(t)
		object := b.signedObject(b.requestObject())
		object.key = otherKey

		requirePushRefused(t, b.par(b.parForm(object)), http.StatusBadRequest, "invalid_request_object")
	})
}

func TestARequestObjectWithASharingDurationAndAnACRThatTheDataHolderReachesIsAccepted(t *testing.T) {
	b := newBank(t)
	claims := b.requestObject()
	claims["claims"] = map[string]any{
		"sharing_duration": 7776000,
		"id_token": map[string]any{
			"acr":       map[string]any{"essential": true, "values": []string{"urn:cds.au:cdr:2"}},
			"auth_time": map[string]any{"essential": true},
		},
		"userinfo": map[string]any{"given_name": nil, "family_name": nil},
	}

	r := b.push(claims)

	require.Equal(t, http.StatusCreated, r.status, r.body)
}

func TestTheCauseOfARefusalDoesNotRepeatTheRequestObject(t *testing.T) {
	b := newBank(t)
	claims := b.requestObject()
	claims["state"] = "state-that-must-not-come-back"
	delete(claims, "nonce")

	form := b.parForm(b.signedObject(claims))
	r := b.par(form)

	description := requirePushRefused(t, r, http.StatusBadRequest, "invalid_request")
	assert.NotContains(t, description, "state-that-must-not-come-back")
	assert.NotContains(t, description, form.Get("request"))
}

func TestAPushedRequestUsesTheJTIOfItsClientAssertionOneTime(t *testing.T) {
	b := newBank(t)
	form := b.parForm(b.signedObject(b.requestObject()))

	first, second := b.par(form), b.par(form)

	require.Equal(t, http.StatusCreated, first.status, first.body)
	requirePushRefused(t, second, http.StatusUnauthorized, "invalid_client")
}

func TestThePAREndpointAcceptsOnlyPOST(t *testing.T) {
	b := newBank(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, b.parURL, nil)
	require.NoError(t, err)

	resp, err := b.http.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}
