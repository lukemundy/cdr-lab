package authserver_test

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
)

const (
	productID     = "software-product-1"
	publicBaseURL = "https://api.register.example"
	issuer        = publicBaseURL + "/idp"
	assertionType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)

type clientStore map[string]authserver.Client

func (s clientStore) Client(_ context.Context, id string) (authserver.Client, error) {
	c, ok := s[id]
	if !ok {
		return authserver.Client{}, authserver.ErrClientNotFound
	}

	return c, nil
}

type jtiStore struct {
	mu   sync.Mutex
	seen map[string]bool
}

func (s *jtiStore) Consume(_ context.Context, jti string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.seen[jti] {
		return authserver.ErrJTIUsed
	}

	s.seen[jti] = true

	return nil
}

// lab is a Register authorisation server over mTLS, with one software product that has a JWKS server.
type lab struct {
	t          *testing.T
	clients    clientStore
	signingKey *rsa.PrivateKey
	// signingKeysErr makes the signing key source fail.
	signingKeysErr error
	// otherKeys follow the signing key in the key set.
	otherKeys []authserver.SigningKey
	register  *authserver.Register
	secureURL string
	tokenURL  string
	public    http.Handler
	cert      tls.Certificate
	rsaKey    *rsa.PrivateKey
	ecKey     *ecdsa.PrivateKey
	http      *http.Client
}

func clientCertificate(t *testing.T, commonName string) tls.Certificate {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: serial, Subject: pkix.Name{CommonName: commonName},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	leaf, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}
}

func thumbprint(cert tls.Certificate) string {
	sum := sha256.Sum256(cert.Certificate[0])
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func newLab(t *testing.T) *lab {
	t.Helper()
	l := &lab{t: t, cert: clientCertificate(t, "Recipient App Web")}

	var err error

	l.signingKey, err = rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	l.rsaKey, err = rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	l.ecKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
			{Key: &l.rsaKey.PublicKey, KeyID: "rsa", Algorithm: "PS256", Use: "sig"},
			{Key: &l.ecKey.PublicKey, KeyID: "ec", Algorithm: "ES256", Use: "sig"},
		}})
	}))
	t.Cleanup(jwks.Close)
	l.clients = clientStore{productID: {
		ID: productID, JWKSURI: jwks.URL, Active: true, CertificateThumbprints: []string{thumbprint(l.cert)},
	}}

	srv := httptest.NewUnstartedServer(nil)
	srv.TLS = &tls.Config{ClientAuth: tls.RequireAnyClientCert, MinVersion: tls.VersionTLS12}
	l.secureURL = "https://" + srv.Listener.Addr().String()
	l.tokenURL = l.secureURL + "/idp/connect/token"
	register, err := authserver.NewRegister(authserver.RegisterConfig{
		PublicBaseURL: publicBaseURL,
		SecureBaseURL: l.secureURL,
		Clients:       l.clients,
		JTIs:          &jtiStore{seen: map[string]bool{}},
		SigningKeys: func(context.Context) ([]authserver.SigningKey, error) {
			return append(
				[]authserver.SigningKey{{ID: "register-1", Key: l.signingKey}},
				l.otherKeys...), l.signingKeysErr
		},
		JWKSTransport: jwks.Client().Transport,
	})
	require.NoError(t, err)

	l.register = register
	secure, public := http.NewServeMux(), http.NewServeMux()
	register.RegisterSecure(secure)
	register.RegisterPublic(public)
	l.public = public
	srv.Config.Handler = secure
	srv.StartTLS()
	t.Cleanup(srv.Close)
	l.http = srv.Client()
	l.useCertificate(l.cert)

	return l
}

func (l *lab) useCertificate(cert tls.Certificate) {
	l.http.Transport.(*http.Transport).TLSClientConfig.Certificates = []tls.Certificate{cert}
	l.http.CloseIdleConnections()
}

type assertion struct {
	alg    jose.SignatureAlgorithm
	key    crypto.Signer
	kid    string
	claims map[string]any
}

func (l *lab) assertion() assertion {
	now := time.Now()

	return assertion{alg: jose.PS256, key: l.rsaKey, kid: "rsa", claims: map[string]any{
		"iss": productID, "sub": productID, "aud": l.tokenURL, "jti": rand.Text(),
		"iat": now.Unix(), "exp": now.Add(time.Minute).Unix(),
	}}
}

func (l *lab) sign(a assertion) string {
	l.t.Helper()

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: a.alg, Key: jose.JSONWebKey{Key: a.key, KeyID: a.kid}},
		(&jose.SignerOptions{}).WithType("JWT"))
	require.NoError(l.t, err)
	token, err := jwt.Signed(signer).Claims(a.claims).Serialize()
	require.NoError(l.t, err)

	return token
}

func (l *lab) form(a assertion) url.Values {
	return url.Values{
		"grant_type":            {"client_credentials"},
		"client_id":             {productID},
		"client_assertion_type": {assertionType},
		"client_assertion":      {l.sign(a)},
		"scope":                 {"cdr-register:read"},
	}
}

type reply struct {
	status int
	header http.Header
	body   map[string]any
}

func (l *lab) token(form url.Values) reply {
	l.t.Helper()
	req, err := http.NewRequestWithContext(l.t.Context(), http.MethodPost, l.tokenURL, strings.NewReader(form.Encode()))
	require.NoError(l.t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := l.http.Do(req)
	require.NoError(l.t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(l.t, err)

	var body map[string]any
	require.NoError(l.t, json.Unmarshal(data, &body), string(data))

	return reply{status: resp.StatusCode, header: resp.Header, body: body}
}

func (l *lab) accessTokenClaims(r reply) map[string]any {
	l.t.Helper()
	require.Equal(l.t, http.StatusOK, r.status, r.body)
	parsed, err := jwt.ParseSigned(r.body["access_token"].(string), []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(l.t, err)
	require.Equal(l.t, "register-1", parsed.Headers[0].KeyID)

	var claims map[string]any
	require.NoError(
		l.t,
		parsed.Claims(&l.signingKey.PublicKey, &claims),
		"the access token is signed with the Register signing key",
	)

	return claims
}

// requireError checks the RFC 6749 section 5.2 format.
func requireError(t *testing.T, r reply, status int, code string) {
	t.Helper()
	require.Equal(t, status, r.status, r.body)
	assert.Equal(t, code, r.body["error"], r.body)
	assert.Contains(t, r.header.Get("Content-Type"), "application/json")
	assert.Equal(t, "no-store", r.header.Get("Cache-Control"))
	assert.NotContains(t, r.body, "access_token")

	for field := range r.body {
		assert.Contains(t, []string{"error", "error_description", "error_uri"}, field)
	}
}

func TestAnActiveSoftwareProductWithARegisteredCertificateGetsABoundToken(t *testing.T) {
	l := newLab(t)

	r := l.token(l.form(l.assertion()))

	claims := l.accessTokenClaims(r)
	assert.Equal(t, "Bearer", r.body["token_type"])
	assert.EqualValues(t, 300, r.body["expires_in"])
	assert.Equal(t, "cdr-register:read", r.body["scope"])
	assert.NotContains(t, r.body, "refresh_token")
	assert.Equal(t, "no-store", r.header.Get("Cache-Control"))
	assert.Equal(t, issuer, claims["iss"])
	assert.Equal(t, productID, claims["sub"])
	assert.Equal(t, productID, claims["client_id"])
	assert.Equal(t, "cdr-register:read", claims["scope"])
	assert.Equal(t, map[string]any{"x5t#S256": thumbprint(l.cert)}, claims["cnf"])
	assert.InDelta(t, time.Now().Add(5*time.Minute).Unix(), claims["exp"], 5)
}

func TestTheClientAssertionAlgorithms(t *testing.T) {
	l := newLab(t)

	tests := map[string]struct {
		alg    jose.SignatureAlgorithm
		key    crypto.Signer
		kid    string
		accept bool
	}{
		"PS256": {alg: jose.PS256, key: l.rsaKey, kid: "rsa", accept: true},
		"ES256": {alg: jose.ES256, key: l.ecKey, kid: "ec", accept: true},
		"RS256": {alg: jose.RS256, key: l.rsaKey, kid: "rsa", accept: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := l.assertion()
			a.alg, a.key, a.kid = tt.alg, tt.key, tt.kid

			r := l.token(l.form(a))

			if tt.accept {
				assert.Equal(t, http.StatusOK, r.status, r.body)
			} else {
				requireError(t, r, http.StatusUnauthorized, "invalid_client")
			}
		})
	}
}

func TestTheAcceptedAudiences(t *testing.T) {
	l := newLab(t)

	tests := map[string]struct {
		aud    string
		accept bool
	}{
		"the issuer":         {issuer, true},
		"the token endpoint": {l.tokenURL, true},
		"a different server": {"https://holder.example/token", false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := l.assertion()
			a.claims["aud"] = tt.aud

			r := l.token(l.form(a))

			if tt.accept {
				assert.Equal(t, http.StatusOK, r.status, r.body)
			} else {
				requireError(t, r, http.StatusUnauthorized, "invalid_client")
			}
		})
	}
}

func TestTheTokenEndpointRefusesAClientThatItCannotTrust(t *testing.T) {
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tests := map[string]func(l *lab) url.Values{
		"a certificate that is not registered, with the same common name": func(l *lab) url.Values {
			l.useCertificate(clientCertificate(l.t, "Recipient App Web"))
			return l.form(l.assertion())
		},
		"a software product that is not ACTIVE": func(l *lab) url.Values {
			c := l.clients[productID]
			c.Active = false
			l.clients[productID] = c

			return l.form(l.assertion())
		},
		"a software product that does not exist": func(l *lab) url.Values {
			a := l.assertion()
			a.claims["iss"], a.claims["sub"] = "unknown", "unknown"
			form := l.form(a)
			form.Set("client_id", "unknown")

			return form
		},
		"a JWT ID that was used before": func(l *lab) url.Values {
			a := l.assertion()
			require.Equal(l.t, http.StatusOK, l.token(l.form(a)).status)

			return l.form(a)
		},
		"a key that is not at the JWKS URI": func(l *lab) url.Values {
			a := l.assertion()
			a.key = otherKey

			return l.form(a)
		},
		"a JWKS URI that does not answer": func(l *lab) url.Values {
			c := l.clients[productID]
			c.JWKSURI = "https://127.0.0.1:1/jwks"
			l.clients[productID] = c

			return l.form(l.assertion())
		},
		"an assertion that is expired": func(l *lab) url.Values {
			a := l.assertion()
			a.claims["exp"] = time.Now().Add(-time.Hour).Unix()

			return l.form(a)
		},
		"an assertion that is valid for more than 10 minutes": func(l *lab) url.Values {
			a := l.assertion()
			a.claims["exp"] = time.Now().Add(11 * time.Minute).Unix()

			return l.form(a)
		},
		"an assertion of a different software product": func(l *lab) url.Values {
			a := l.assertion()
			a.claims["iss"], a.claims["sub"] = "software-product-2", "software-product-2"

			return l.form(a)
		},
		"no assertion": func(l *lab) url.Values {
			form := l.form(l.assertion())
			form.Del("client_assertion")
			form.Del("client_assertion_type")

			return form
		},
	}
	for name, form := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLab(t)

			r := l.token(form(l))

			requireError(t, r, http.StatusUnauthorized, "invalid_client")
		})
	}
}

func TestTheTokenEndpointRefusesOtherGrantsAndScopes(t *testing.T) {
	tests := map[string]struct {
		field, value string
		code         string
	}{
		"authorization_code grant": {"grant_type", "authorization_code", "unsupported_grant_type"},
		"refresh_token grant":      {"grant_type", "refresh_token", "unsupported_grant_type"},
		"a scope of the data APIs": {"scope", "bank:accounts.basic:read", "invalid_scope"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLab(t)
			form := l.form(l.assertion())
			form.Set(tt.field, tt.value)

			r := l.token(form)

			requireError(t, r, http.StatusBadRequest, tt.code)
		})
	}
}

func TestAStoreFailureIsAServerErrorWithNoDetail(t *testing.T) {
	l := newLab(t)
	l.clients[productID] = authserver.Client{}
	failing := failingStore{}
	register, err := authserver.NewRegister(authserver.RegisterConfig{
		PublicBaseURL: publicBaseURL,
		SecureBaseURL: l.secureURL,
		Clients:       failing,
		JTIs:          &jtiStore{seen: map[string]bool{}},
		SigningKeys:   func(context.Context) ([]authserver.SigningKey, error) { return nil, errors.New("no keys") },
	})
	require.NoError(t, err)

	mux := http.NewServeMux()
	register.RegisterSecure(mux)

	form := l.form(l.assertion())
	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"/idp/connect/token",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{l.cert.Leaf}}
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":"server_error"}`, rec.Body.String())
}

type failingStore struct{}

func (failingStore) Client(context.Context, string) (authserver.Client, error) {
	return authserver.Client{}, errors.New("connection to 10.0.0.1 failed")
}

func TestTheOnlySecureRouteIsTheTokenEndpoint(t *testing.T) {
	l := newLab(t)
	for _, path := range []string{
		"/idp/connect/register",
		"/idp/connect/userinfo",
		"/idp/connect/authorize",
		"/.well-known/openid-configuration/idp",
	} {
		t.Run(path, func(t *testing.T) {
			req, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				l.secureURL+path,
				strings.NewReader("{}"),
			)
			require.NoError(t, err)
			resp, err := l.http.Do(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	}
}

func (l *lab) getPublic(path string) (*httptest.ResponseRecorder, map[string]any) {
	l.t.Helper()

	rec := httptest.NewRecorder()
	l.public.ServeHTTP(rec, httptest.NewRequestWithContext(l.t.Context(), http.MethodGet, path, nil))
	require.Equal(l.t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(l.t, json.Unmarshal(rec.Body.Bytes(), &body))

	return rec, body
}

func TestTheOpenIDProviderConfiguration(t *testing.T) {
	l := newLab(t)

	rec, body := l.getPublic("/idp/.well-known/openid-configuration")

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Equal(t, map[string]any{
		"issuer":         issuer,
		"jwks_uri":       issuer + "/.well-known/openid-configuration/jwks",
		"token_endpoint": l.tokenURL,
		"claims_supported": []any{
			"iss",
			"sub",
			"aud",
			"exp",
			"iat",
			"jti",
			"client_id",
			"scope",
			"cnf",
		},
		"id_token_signing_alg_values_supported":            []any{"PS256"},
		"subject_types_supported":                          []any{"public"},
		"code_challenge_methods_supported":                 []any{},
		"scopes_supported":                                 []any{"cdr-register:read"},
		"response_types_supported":                         []any{"token"},
		"grant_types_supported":                            []any{"client_credentials"},
		"token_endpoint_auth_methods_supported":            []any{"private_key_jwt"},
		"tls_client_certificate_bound_access_tokens":       true,
		"token_endpoint_auth_signing_alg_values_supported": []any{"PS256", "ES256"},
	}, body)
}

func TestTheJWKSHasThePublicSigningKeyOnly(t *testing.T) {
	l := newLab(t)

	_, body := l.getPublic("/idp/.well-known/openid-configuration/jwks")

	keys := body["keys"].([]any)
	require.Len(t, keys, 1)
	key := keys[0].(map[string]any)
	assert.Equal(t, map[string]any{
		"kty": "RSA", "kid": "register-1", "alg": "PS256", "key_ops": []any{"verify"},
		"n": base64.RawURLEncoding.EncodeToString(l.signingKey.N.Bytes()),
		"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(l.signingKey.E)).Bytes()),
	}, key)
}

func TestNewRegisterRefusesAConfigurationThatIsNotComplete(t *testing.T) {
	_, err := authserver.NewRegister(authserver.RegisterConfig{PublicBaseURL: publicBaseURL})

	assert.Error(t, err)
}
