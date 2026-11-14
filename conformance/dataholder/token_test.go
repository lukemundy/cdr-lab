//go:build conformance

package dataholder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

const registrationScope = "cdr:registration"

func tokenEndpoint(t *testing.T, client registeredClient) string {
	t.Helper()

	endpoint := openIDConfiguration(t, client.issuer)["token_endpoint"]
	require.IsType(t, "", endpoint)

	return endpoint.(string)
}

func (c registeredClient) assertion(t *testing.T, audience string) labtest.ClientAssertion {
	t.Helper()

	return labtest.NewAssertion(recipientKeys(t), c.clientID, audience)
}

func (c registeredClient) requestToken(t *testing.T, a labtest.ClientAssertion) labtest.TokenReply {
	t.Helper()

	form := labtest.TokenForm(t, c.clientID, registrationScope, a)

	return labtest.RequestToken(t, tokenEndpoint(t, c), c.cert, form)
}

func (c registeredClient) accessToken(t *testing.T) string {
	t.Helper()

	r := c.requestToken(t, c.assertion(t, tokenEndpoint(t, c)))
	require.Equal(t, http.StatusOK, r.Status, r.Body)

	return r.Body["access_token"].(string)
}

// requireTokenError checks the RFC 6749 section 5.2 format.
func requireTokenError(t *testing.T, r labtest.TokenReply, status int, code string) {
	t.Helper()

	require.Equal(t, status, r.Status, r.Body)
	assert.Equal(t, code, r.Body["error"])
	assert.Contains(t, r.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, "no-store", r.Header.Get("Cache-Control"))
	assert.NotContains(t, r.Body, "access_token")
}

type registrationRead struct {
	status int
	header http.Header
	body   []byte
}

func (c registeredClient) readRegistration(
	t *testing.T,
	cert tls.Certificate,
	clientID, token string,
) registrationRead {
	t.Helper()

	return c.manageRegistration(t, http.MethodGet, cert, clientID, token, "")
}

// manageRegistration sends requestJWT as the body when it is not empty, and validates a 200 response against the
// pinned specification.
func (c registeredClient) manageRegistration(
	t *testing.T,
	method string,
	cert tls.Certificate,
	clientID, token, requestJWT string,
) registrationRead {
	t.Helper()

	httpClient := &http.Client{Timeout: labtest.RequestTimeout, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      labtest.CDRCABundle(t),
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}}
	target := c.endpoint + "/" + url.PathEscape(clientID)
	req, err := http.NewRequestWithContext(t.Context(), method, target, strings.NewReader(requestJWT))
	require.NoError(t, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	if requestJWT != "" {
		req.Header.Set("Content-Type", jwtMediaType)
	}

	resp, err := httpClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		spec, err := dcrSpec()
		require.NoError(t, err)
		require.NoError(t, validateAtSpecPath(spec, resp, specRegistrationPath+"/"+clientID))
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return registrationRead{status: resp.StatusCode, header: resp.Header, body: body}
}

func TestARegisteredSoftwareProductGetsATokenThatIsBoundToItsCertificate(t *testing.T) {
	client := registered(t, registrationTestProduct)

	r := client.requestToken(t, client.assertion(t, tokenEndpoint(t, client)))

	require.Equal(t, http.StatusOK, r.Status, r.Body)
	assert.Equal(t, "Bearer", r.Body["token_type"])
	assert.EqualValues(t, 300, r.Body["expires_in"])
	assert.Equal(t, registrationScope, r.Body["scope"])
	assert.NotContains(t, r.Body, "refresh_token")
	assert.Equal(t, "no-store", r.Header.Get("Cache-Control"))
	assert.NotEmpty(t, r.Header.Get("x-fapi-interaction-id"))
	assert.Empty(t, r.Header.Get("Access-Control-Allow-Origin"))
}

func TestATokenRequestWithNoScopeGetsTheRegistrationScope(t *testing.T) {
	client := registered(t, registrationTestProduct)
	endpoint := tokenEndpoint(t, client)
	form := labtest.TokenForm(t, client.clientID, registrationScope, client.assertion(t, endpoint))
	form.Del("scope")

	r := labtest.RequestToken(t, endpoint, client.cert, form)

	require.Equal(t, http.StatusOK, r.Status, r.Body)
	assert.Equal(t, registrationScope, r.Body["scope"])
}

func TestATokenRequestForADifferentScopeIsRefused(t *testing.T) {
	client := registered(t, registrationTestProduct)
	endpoint := tokenEndpoint(t, client)

	for _, scope := range []string{"bank:accounts.basic:read", "cdr:registration bank:accounts.basic:read"} {
		t.Run(scope, func(t *testing.T) {
			form := labtest.TokenForm(t, client.clientID, scope, client.assertion(t, endpoint))

			requireTokenError(
				t,
				labtest.RequestToken(t, endpoint, client.cert, form),
				http.StatusBadRequest,
				"invalid_scope",
			)
		})
	}
}

func TestASoftwareProductWithNoRegistrationScopeGetsNoToken(t *testing.T) {
	client := registered(t, dataOnlyProduct)
	assert.NotContains(t, client.reply.body["scope"], registrationScope)

	r := client.requestToken(t, client.assertion(t, tokenEndpoint(t, client)))

	requireTokenError(t, r, http.StatusBadRequest, "invalid_scope")
}

func TestTheTokenEndpointAcceptsItsThreeAudiencesAndTwoAlgorithms(t *testing.T) {
	client := registered(t, registrationTestProduct)
	endpoint := tokenEndpoint(t, client)
	keys := recipientKeys(t)

	t.Run("PS256", func(t *testing.T) {
		for name, audience := range map[string]string{"the issuer": client.issuer, "the token endpoint": endpoint} {
			t.Run(name, func(t *testing.T) {
				r := client.requestToken(t, client.assertion(t, audience))

				assert.Equal(t, http.StatusOK, r.Status, r.Body)
			})
		}
	})

	t.Run("ES256", func(t *testing.T) {
		a := client.assertion(t, endpoint)
		a.Alg, a.Key, a.KID = jose.ES256, keys.EC, "ec"

		r := client.requestToken(t, a)

		assert.Equal(t, http.StatusOK, r.Status, r.Body)
	})
}

func TestTheTokenEndpointRefusesAClientAssertionThatIsNotValid(t *testing.T) {
	client := registered(t, registrationTestProduct)
	endpoint := tokenEndpoint(t, client)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	usedBefore := client.assertion(t, endpoint)
	require.Equal(t, http.StatusOK, client.requestToken(t, usedBefore).Status)

	tests := map[string]func(a *labtest.ClientAssertion){
		"a signature from a different key": func(a *labtest.ClientAssertion) { a.Key = otherKey },
		"a client ID that does not exist": func(a *labtest.ClientAssertion) {
			a.Claims["iss"], a.Claims["sub"] = registrationTestProduct, registrationTestProduct
		},
		"an exp in the past": func(a *labtest.ClientAssertion) {
			a.Claims["exp"] = time.Now().Add(-time.Minute).Unix()
		},
		"an exp of more than 10 minutes": func(a *labtest.ClientAssertion) {
			a.Claims["exp"] = time.Now().Add(11 * time.Minute).Unix()
		},
		"a jti that was used before": func(a *labtest.ClientAssertion) { a.Claims["jti"] = usedBefore.Claims["jti"] },
		"an audience of a different server": func(a *labtest.ClientAssertion) {
			a.Claims["aud"] = "https://other.example/token"
		},
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			a := client.assertion(t, endpoint)
			change(&a)

			requireTokenError(t, client.requestToken(t, a), http.StatusUnauthorized, "invalid_client")
		})
	}
}

func TestTheTokenEndpointRefusesOtherGrantTypes(t *testing.T) {
	client := registered(t, registrationTestProduct)
	endpoint := tokenEndpoint(t, client)

	for _, grant := range []string{"urn:ietf:params:oauth:grant-type:jwt-bearer", "password"} {
		t.Run(grant, func(t *testing.T) {
			form := labtest.TokenForm(t, client.clientID, registrationScope, client.assertion(t, endpoint))
			form.Set("grant_type", grant)

			r := labtest.RequestToken(t, endpoint, client.cert, form)

			requireTokenError(t, r, http.StatusBadRequest, "unsupported_grant_type")
		})
	}
}

func TestASoftwareProductReadsItsRegistrationWithItsToken(t *testing.T) {
	client := registered(t, registrationTestProduct)

	read := client.readRegistration(t, client.cert, client.clientID, client.accessToken(t))

	require.Equal(t, http.StatusOK, read.status, string(read.body))
	assert.Equal(t, "no-store", read.header.Get("Cache-Control"))

	var registration map[string]any
	require.NoError(t, json.Unmarshal(read.body, &registration))
	assert.Equal(t, client.reply.body, registration, "the registration is the one that the registration response gave")
}

func TestAReadOfARegistrationIsRefusedWithNoBody(t *testing.T) {
	client := registered(t, registrationTestProduct)
	other := registered(t, minimalProduct)
	token := client.accessToken(t)
	otherCert := labtest.ClientCertificate(t, "PKI_CLIENT_CERT", "PKI_CLIENT_KEY")

	tests := map[string]struct {
		cert            tls.Certificate
		clientID, token string
		status          int
		challenge       string
	}{
		"no token": {client.cert, client.clientID, "", http.StatusUnauthorized, "Bearer"},
		"a token that the data holder did not issue": {
			client.cert, client.clientID, "not-a-token", http.StatusUnauthorized, `Bearer error="invalid_token"`,
		},
		"a different client certificate": {
			otherCert, client.clientID, token, http.StatusUnauthorized, `Bearer error="invalid_token"`,
		},
		"a client ID that does not exist": {
			client.cert, uuid.NewString(), token, http.StatusUnauthorized, `Bearer error="invalid_token"`,
		},
		"the client ID of a different registration": {
			client.cert, other.clientID, token, http.StatusForbidden, "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			read := client.readRegistration(t, tc.cert, tc.clientID, tc.token)

			assert.Equal(t, tc.status, read.status, string(read.body))
			assert.Equal(t, tc.challenge, read.header.Get("WWW-Authenticate"))
			assert.Empty(t, read.body)
		})
	}
}
