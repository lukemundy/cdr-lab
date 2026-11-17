package authserver_test

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

const (
	registeredClientID    = "client-1"
	secondClientID        = "second-client"
	registeredRedirectURI = "https://recipient.example/callback"
)

// bank is a data holder authorisation server over mTLS, with two registered clients that share a JWKS server. The
// requests of the bank are those of the first client.
type bank struct {
	t          *testing.T
	clients    dataHolderClients
	dataHolder *authserver.DataHolder
	cfg        authserver.DataHolderConfig
	grants     *grantStore
	revoked    *revokedTokenStore
	// secure is the handler of the mTLS listener.
	secure           http.Handler
	tokenURL         string
	parURL           string
	introspectionURL string
	revocationURL    string
	arrangementURL   string
	// publicURL is the base URL of the listener with the authorisation endpoint.
	publicURL   string
	pages       *pages
	statuses    *clientStatuses
	issuer      string
	signingKeys func(context.Context) ([]authserver.SigningKey, error)
	http        *http.Client
	// browser calls the public listener and does not follow a redirect.
	browser *http.Client
	cert    tls.Certificate
	rsaKey  *rsa.PrivateKey
	ecKey   *ecdsa.PrivateKey
}

func newBank(t *testing.T) *bank {
	t.Helper()

	b := &bank{t: t, cert: clientCertificate(t, "Recipient App Web")}

	var err error

	b.rsaKey, err = rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	b.ecKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
			{Key: &b.rsaKey.PublicKey, KeyID: "rsa", Algorithm: "PS256", Use: "sig"},
			{Key: &b.ecKey.PublicKey, KeyID: "ec", Algorithm: "ES256", Use: "sig"},
		}})
	}))
	t.Cleanup(jwks.Close)

	b.clients = dataHolderClients{}
	for _, id := range []string{registeredClientID, secondClientID} {
		b.clients[id] = authserver.DataHolderClient{
			ID:                          id,
			JWKSURI:                     jwks.URL,
			Scopes:                      []string{"openid", "bank:accounts.basic:read", "cdr:registration"},
			GrantTypes:                  []string{"client_credentials", "authorization_code", "refresh_token"},
			RedirectURIs:                []string{registeredRedirectURI},
			RequestObjectSigningAlg:     "PS256",
			TokenEndpointAuthSigningAlg: "PS256",
		}
	}

	srv := httptest.NewUnstartedServer(nil)
	srv.TLS = &tls.Config{ClientAuth: tls.RequireAnyClientCert, MinVersion: tls.VersionTLS12}
	secureURL := "https://" + srv.Listener.Addr().String()
	b.tokenURL = secureURL + "/token"
	b.parURL = secureURL + "/par"
	b.introspectionURL = secureURL + "/introspect"
	b.revocationURL = secureURL + "/revoke"
	b.arrangementURL = secureURL + "/arrangements/revoke"

	cfg := dataHolderConfig(t)
	b.issuer = cfg.Issuer
	b.signingKeys = cfg.SigningKeys
	cfg.SecureBaseURL = secureURL
	cfg.Clients = b.clients
	cfg.JWKSTransport = jwks.Client().Transport
	b.pages = &pages{}
	cfg.Pages = b.pages
	b.grants = &grantStore{grants: map[string]authserver.Grant{}}
	cfg.Grants = b.grants
	b.revoked = cfg.RevokedTokens.(*revokedTokenStore)
	cfg.Arrangements = arrangementGrants{b.grants}
	b.statuses = &clientStatuses{}
	cfg.ClientStatuses = b.statuses
	b.cfg = cfg
	b.dataHolder, err = authserver.NewDataHolder(cfg)
	require.NoError(t, err)

	secure := http.NewServeMux()
	b.dataHolder.RegisterSecure(secure)
	b.secure = secure
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { b.secure.ServeHTTP(w, r) })
	srv.StartTLS()
	t.Cleanup(srv.Close)

	public := http.NewServeMux()
	b.dataHolder.RegisterAuthorisation(public)
	publicServer := httptest.NewTLSServer(public)
	t.Cleanup(publicServer.Close)
	b.publicURL = publicServer.URL
	b.browser = publicServer.Client()
	b.browser.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	b.http = srv.Client()
	b.http.Transport.(*http.Transport).TLSClientConfig.Certificates = []tls.Certificate{b.cert}

	return b
}

func (b *bank) assertion() assertion {
	now := time.Now()

	return assertion{alg: jose.PS256, key: b.rsaKey, kid: "rsa", claims: map[string]any{
		"iss": registeredClientID, "sub": registeredClientID, "aud": b.tokenURL, "jti": rand.Text(),
		"iat": now.Unix(), "exp": now.Add(time.Minute).Unix(),
	}}
}

func (b *bank) form(a assertion) url.Values {
	b.t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: a.alg, Key: jose.JSONWebKey{Key: a.key, KeyID: a.kid}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	require.NoError(b.t, err)
	signed, err := jwt.Signed(signer).Claims(a.claims).Serialize()
	require.NoError(b.t, err)

	return url.Values{
		"grant_type":            {"client_credentials"},
		"client_assertion_type": {assertionType},
		"client_assertion":      {signed},
		"scope":                 {"cdr:registration"},
	}
}

// restart replaces the data holder with a new one that has only the stores of the first.
func (b *bank) restart() {
	b.t.Helper()

	dataHolder, err := authserver.NewDataHolder(b.cfg)
	require.NoError(b.t, err)

	secure := http.NewServeMux()
	dataHolder.RegisterSecure(secure)
	b.dataHolder, b.secure = dataHolder, secure
}

func (b *bank) token(form url.Values) reply {
	b.t.Helper()
	return b.post(b.tokenURL, form)
}

func (b *bank) post(target string, form url.Values) reply {
	b.t.Helper()

	req, err := http.NewRequestWithContext(b.t.Context(), http.MethodPost, target, strings.NewReader(form.Encode()))
	require.NoError(b.t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := b.http.Do(req)
	require.NoError(b.t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(b.t, err)

	var body map[string]any
	if len(data) > 0 {
		require.NoError(b.t, json.Unmarshal(data, &body), string(data))
	}

	return reply{status: resp.StatusCode, header: resp.Header, body: body}
}

func (b *bank) accessToken() string {
	b.t.Helper()

	r := b.token(b.form(b.assertion()))
	require.Equal(b.t, http.StatusOK, r.status, r.body)

	return r.body["access_token"].(string)
}

func leaf(t *testing.T, cert tls.Certificate) *x509.Certificate {
	t.Helper()

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	require.NoError(t, err)

	return parsed
}

// Verifies: AUTH-0038, AUTH-0042, AUTH-0043, AUTH-0046, AUTH-0047
func TestARegisteredClientGetsARegistrationTokenThatIsBoundToItsCertificate(t *testing.T) {
	b := newBank(t)

	r := b.token(b.form(b.assertion()))

	require.Equal(t, http.StatusOK, r.status, r.body)
	assert.Equal(t, "Bearer", r.body["token_type"])
	assert.EqualValues(t, 300, r.body["expires_in"])
	assert.Equal(t, "cdr:registration", r.body["scope"])
	assert.NotContains(t, r.body, "refresh_token")
	assert.Equal(t, "no-store", r.header.Get("Cache-Control"))

	verified, err := b.dataHolder.VerifyAccessToken(t.Context(), r.body["access_token"].(string), leaf(t, b.cert))

	require.NoError(t, err)
	assert.Equal(
		t,
		authserver.AccessToken{ClientID: registeredClientID, Scopes: []string{"cdr:registration"}},
		verified,
	)
}

// Verifies: AUTH-0039
func TestAClientCredentialsRequestWithNoScopeGetsTheRegistrationScope(t *testing.T) {
	b := newBank(t)
	form := b.form(b.assertion())
	form.Del("scope")

	r := b.token(form)

	require.Equal(t, http.StatusOK, r.status, r.body)
	assert.Equal(t, "cdr:registration", r.body["scope"])
}

// Verifies: AUTH-0040
func TestAClientCredentialsTokenHasOnlyTheRegistrationScope(t *testing.T) {
	for _, scope := range []string{"bank:accounts.basic:read", "cdr:registration bank:accounts.basic:read", "openid"} {
		t.Run(scope, func(t *testing.T) {
			b := newBank(t)
			form := b.form(b.assertion())
			form.Set("scope", scope)

			requireError(t, b.token(form), http.StatusBadRequest, "invalid_scope")
		})
	}
}

// Verifies: AUTH-0037, AUTH-0041
func TestAClientWhoseRegistrationDoesNotPermitTheRequestIsRefused(t *testing.T) {
	tests := map[string]struct {
		change func(*authserver.DataHolderClient)
		code   string
	}{
		"no registration scope": {
			func(c *authserver.DataHolderClient) { c.Scopes = []string{"openid", "bank:accounts.basic:read"} },
			"invalid_scope",
		},
		"no client credentials grant": {
			func(c *authserver.DataHolderClient) { c.GrantTypes = []string{"authorization_code"} },
			"unauthorized_client",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			b := newBank(t)
			client := b.clients[registeredClientID]
			tc.change(&client)
			b.clients[registeredClientID] = client

			requireError(t, b.token(b.form(b.assertion())), http.StatusBadRequest, tc.code)
		})
	}
}

// registerSigningAlg changes the token_endpoint_auth_signing_alg of the registered client.
func (b *bank) registerSigningAlg(alg string) {
	client := b.clients[registeredClientID]
	client.TokenEndpointAuthSigningAlg = alg
	b.clients[registeredClientID] = client
}

// Verifies: AUTH-0026, AUTH-0027, AUTH-0240
func TestTheDataHolderTokenEndpointAcceptsItsThreeAudiencesAndTwoAlgorithms(t *testing.T) {
	b := newBank(t)

	tests := map[string]struct {
		alg jose.SignatureAlgorithm
		key crypto.Signer
		kid string
		aud string
	}{
		"PS256, the token endpoint as the audience": {jose.PS256, b.rsaKey, "rsa", b.tokenURL},
		"ES256, the token endpoint as the audience": {jose.ES256, b.ecKey, "ec", b.tokenURL},
		"PS256, the issuer as the audience":         {jose.PS256, b.rsaKey, "rsa", dataHolderIssuer},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			b.registerSigningAlg(string(tc.alg))

			a := b.assertion()
			a.alg, a.key, a.kid = tc.alg, tc.key, tc.kid
			a.claims["aud"] = tc.aud

			r := b.token(b.form(a))

			assert.Equal(t, http.StatusOK, r.status, r.body)
		})
	}
}

// Verifies: AUTH-0025, AUTH-0028, AUTH-0029, AUTH-0030, AUTH-0031, AUTH-0032
func TestTheDataHolderTokenEndpointRefusesAnAssertionThatIsNotValid(t *testing.T) {
	b := newBank(t)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	usedBefore := b.assertion()
	require.Equal(t, http.StatusOK, b.token(b.form(usedBefore)).status)

	tests := map[string]func(a *assertion){
		"a signature from a different key": func(a *assertion) { a.key = otherKey },
		"a client ID that does not exist": func(a *assertion) {
			a.claims["iss"], a.claims["sub"] = "client-2", "client-2"
		},
		"an exp in the past": func(a *assertion) { a.claims["exp"] = time.Now().Add(-time.Minute).Unix() },
		"an exp of more than 10 minutes": func(a *assertion) {
			a.claims["exp"] = time.Now().Add(11 * time.Minute).Unix()
		},
		"a jti that was used before": func(a *assertion) { a.claims["jti"] = usedBefore.claims["jti"] },
		"no jti":                     func(a *assertion) { delete(a.claims, "jti") },
		"an audience of a different server": func(a *assertion) {
			a.claims["aud"] = "https://other.example/token"
		},
		"an algorithm that is not permitted": func(a *assertion) { a.alg = jose.RS256 },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			a := b.assertion()
			change(&a)

			requireError(t, b.token(b.form(a)), http.StatusUnauthorized, "invalid_client")
		})
	}
}

// Verifies: AUTH-0240
func TestAnAssertionWithAnAlgorithmThatTheClientDidNotRegisterIsRefusedAtEachEndpoint(t *testing.T) {
	b := newBank(t)
	accessToken := b.accessToken()
	arrangementID := b.authorise(hour).consent.ArrangementID
	// The requests of the bank sign the assertion with PS256.
	b.registerSigningAlg("ES256")

	tests := map[string]func() reply{
		"token":            func() reply { return b.token(b.form(b.assertion())) },
		"PAR":              func() reply { return b.push(b.requestObject()) },
		"introspection":    func() reply { return b.introspect(accessToken) },
		"token revocation": func() reply { return b.revoke(registeredClientID, accessToken) },
		"CDR arrangement revocation": func() reply {
			r := b.revokeArrangement(registeredClientID, url.Values{"cdr_arrangement_id": {arrangementID}})

			var body map[string]any
			require.NoError(t, json.Unmarshal([]byte(r.body), &body), r.body)

			return reply{status: r.status, header: r.header, body: body}
		},
	}
	for name, request := range tests {
		t.Run(name, func(t *testing.T) {
			r := request()

			require.Equal(t, http.StatusUnauthorized, r.status, r.body)
			assert.Equal(t, "invalid_client", r.body["error"], r.body)
		})
	}

	assert.True(t, b.grants.only(t).RevokedAt.IsZero(), "the arrangement does not end")
}

// Verifies: AUTH-0036
func TestTheDataHolderTokenEndpointRefusesOtherGrants(t *testing.T) {
	for _, grant := range []string{"urn:ietf:params:oauth:grant-type:jwt-bearer", "password"} {
		t.Run(grant, func(t *testing.T) {
			b := newBank(t)
			form := b.form(b.assertion())
			form.Set("grant_type", grant)

			requireError(t, b.token(form), http.StatusBadRequest, "unsupported_grant_type")
		})
	}
}

// Verifies: AUTH-0055
func TestVerifyAccessTokenRefusesATokenThatTheClientCannotUse(t *testing.T) {
	b := newBank(t)
	token := b.accessToken()
	other := clientCertificate(t, "Other App")

	t.Run("a different client certificate", func(t *testing.T) {
		_, err := b.dataHolder.VerifyAccessToken(t.Context(), token, leaf(t, other))

		assert.ErrorIs(t, err, authserver.ErrInvalidToken)
	})

	t.Run("no client certificate", func(t *testing.T) {
		_, err := b.dataHolder.VerifyAccessToken(t.Context(), token, nil)

		assert.ErrorIs(t, err, authserver.ErrInvalidToken)
	})

	t.Run("not a token of the data holder", func(t *testing.T) {
		_, err := b.dataHolder.VerifyAccessToken(t.Context(), "not-a-token", leaf(t, b.cert))

		assert.ErrorIs(t, err, authserver.ErrInvalidToken)
	})

	t.Run("a client that is not registered now", func(t *testing.T) {
		delete(b.clients, registeredClientID)

		_, err := b.dataHolder.VerifyAccessToken(context.Background(), token, leaf(t, b.cert))

		assert.ErrorIs(t, err, authserver.ErrInvalidToken)
	})
}
