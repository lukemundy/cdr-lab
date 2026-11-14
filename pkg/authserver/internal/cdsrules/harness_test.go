package cdsrules_test

import (
	"context"
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
	"maps"
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
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver/internal/cdsrules"
)

const (
	clientID      = "software-product-1"
	otherClientID = "software-product-2"
	arrangementID = "5a1bf696-ee03-408b-b315-97955415d1f0"
	pairwiseSub   = "0e0d7ac1-2a0f-4f4c-9c0c-6f2f6c7d3b11"
	redirectURI   = "https://recipient.example/callback"
	dataScope     = "bank:accounts.basic:read"
)

var signingAlgs = []jose.SignatureAlgorithm{jose.PS256}

// authServer is a data holder authorisation server with the CDS options, served over mTLS.
type authServer struct {
	op     *provider.Provider
	issuer string
	key    jose.JSONWebKey
	store  *memoryStore
	// authenticatedAt is the authentication time of each consent.
	authenticatedAt time.Time
}

// memoryStore keeps copies, as a database does, so that a change to a grant has an effect only after a save.
type memoryStore struct {
	mu       sync.Mutex
	sessions map[string]goidc.AuthnSession
	grants   map[string]goidc.Grant
}

func (m *memoryStore) SaveSession(_ context.Context, session *goidc.AuthnSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[session.ID] = *session

	return nil
}

func (m *memoryStore) session(match func(goidc.AuthnSession) bool) (*goidc.AuthnSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, session := range m.sessions {
		if match(session) {
			return &session, nil
		}
	}

	return nil, goidc.ErrNotFound
}

func (m *memoryStore) Session(_ context.Context, id string) (*goidc.AuthnSession, error) {
	return m.session(func(s goidc.AuthnSession) bool { return s.ID == id })
}

func (m *memoryStore) SessionByPushedAuthReqID(_ context.Context, id string) (*goidc.AuthnSession, error) {
	return m.session(func(s goidc.AuthnSession) bool { return s.PushedAuthReqID == id })
}

func (m *memoryStore) SaveGrant(_ context.Context, grant *goidc.Grant) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.grants[grant.ID] = *grant

	return nil
}

func (m *memoryStore) grant(match func(goidc.Grant) bool) (*goidc.Grant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, grant := range m.grants {
		if match(grant) {
			return &grant, nil
		}
	}

	return nil, goidc.ErrNotFound
}

func (m *memoryStore) Grant(_ context.Context, id string) (*goidc.Grant, error) {
	return m.grant(func(g goidc.Grant) bool { return g.ID == id })
}

func (m *memoryStore) GrantByAuthCode(_ context.Context, code string) (*goidc.Grant, error) {
	return m.grant(func(g goidc.Grant) bool { return g.AuthCode == code })
}

func (m *memoryStore) GrantByRefreshToken(_ context.Context, token string) (*goidc.Grant, error) {
	return m.grant(func(g goidc.Grant) bool { return g.RefreshToken != "" && g.RefreshToken == token })
}

// dataRecipient is a software product that calls an authServer over mTLS.
type dataRecipient struct {
	t      *testing.T
	server *authServer
	key    jose.JSONWebKey
	http   *http.Client
}

type ctxKeyClientCert struct{}

func newKey(t *testing.T, kid string) jose.JSONWebKey {
	t.Helper()

	k, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return jose.JSONWebKey{Key: k, KeyID: kid, Algorithm: string(jose.PS256), Use: "sig"}
}

func clientCertificate(t *testing.T) tls.Certificate {
	t.Helper()

	k, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: clientID},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &k.PublicKey, k)
	require.NoError(t, err)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: k}
}

// newAuthServer gives an auth server, and a data recipient that it knows as a static client.
func newAuthServer(t *testing.T) (*authServer, *dataRecipient) {
	t.Helper()

	srv := httptest.NewUnstartedServer(nil)
	srv.TLS = &tls.Config{ClientAuth: tls.RequestClientCert, MinVersion: tls.VersionTLS12}
	server := &authServer{
		issuer: "https://" + srv.Listener.Addr().String(),
		key:    newKey(t, "server"),
		store:  &memoryStore{sessions: map[string]goidc.AuthnSession{}, grants: map[string]goidc.Grant{}},
		// A time in the past shows that auth_time is not the time of the token.
		authenticatedAt: time.Now().Add(-time.Minute).Truncate(time.Second),
	}
	recipient := &dataRecipient{t: t, server: server, key: newKey(t, "recipient")}

	publicClientJWKS := goidc.JSONWebKeySet{Keys: []goidc.JSONWebKey{recipient.key.Public()}}
	staticClient := goidc.Client{
		ID:       clientID,
		ScopeIDs: strings.Join([]string{goidc.ScopeOpenID.ID, dataScope, cdsrules.RegistrationScope}, " "),
		JWKS:     &publicClientJWKS,
		GrantTypes: []goidc.GrantType{
			goidc.GrantAuthorizationCode,
			goidc.GrantRefreshToken,
			goidc.GrantClientCredentials,
		},
		ResponseTypes:    []goidc.ResponseType{goidc.ResponseTypeCode},
		RedirectURIs:     []string{redirectURI},
		TokenAuthnMethod: goidc.AuthnMethodPrivateKeyJWT,
	}
	otherClient := staticClient
	otherClient.ID = otherClientID

	op, err := provider.New(
		provider.Config{
			Issuer:  server.issuer,
			Manager: server.store,
			JWKS: func(context.Context) (goidc.JSONWebKeySet, error) {
				return goidc.JSONWebKeySet{Keys: []goidc.JSONWebKey{server.key}}, nil
			},
			IDTokenAlgs: []goidc.SignatureAlgorithm{goidc.SigAlgPS256},
		},
		provider.WithProfile(goidc.ProfileFAPI1),
		provider.WithScopes(goidc.ScopeOpenID, goidc.NewScope(dataScope), goidc.NewScope(cdsrules.RegistrationScope)),
		provider.WithMTLS(provider.MTLSConfig{Host: server.issuer, ClientCert: clientCertFromContext},
			provider.WithMTLSTokenBinding(provider.WithMTLSTokenBindingRequired())),
		provider.WithPrivateKeyJWTAuthn(goidc.SigAlgPS256),
		provider.WithAuthCodeGrant(
			provider.AuthCodeGrantConfig{
				Manager:       server.store,
				ResponseTypes: []goidc.ResponseType{goidc.ResponseTypeCode},
			},
			provider.WithJAR([]goidc.SignatureAlgorithm{goidc.SigAlgPS256}, provider.WithJARRequired()),
			provider.WithPAR(
				server.store,
				provider.WithPARRequired(),
				provider.WithPARSessionHandler(cdsrules.PushedRequest),
			),
			provider.WithPKCE([]goidc.CodeChallengeMethod{goidc.CodeChallengeMethodSHA256}),
			provider.WithJARM([]goidc.SignatureAlgorithm{goidc.SigAlgPS256}),
			provider.WithAuthPolicies(goidc.NewPolicy(
				"auto",
				func(*http.Request, *goidc.AuthnSession, *goidc.Client) bool { return true },
				func(
					_ http.ResponseWriter,
					_ *http.Request,
					as *goidc.AuthnSession,
					_ *goidc.Client,
				) (goidc.Status, error) {
					cdsrules.Approve(as, cdsrules.Consent{
						ArrangementID:   arrangementID,
						Subject:         pairwiseSub,
						AuthenticatedAt: server.authenticatedAt,
						SharingEndsAt:   time.Now().Add(cdsrules.SharingDuration(as.Store)),
					})

					return goidc.StatusSuccess, nil
				},
			)),
		),
		provider.WithRefreshTokenGrant(
			cdsrules.RefreshWithAnyCertificate(server.store, clientCertFromContext),
			provider.WithRefreshTokenShouldIssue(cdsrules.RefreshTokenForSharingDuration),
		),
		provider.WithTokenHandler(cdsrules.RecordArrangement),
		provider.WithIDTokenClaims(cdsrules.IDTokenClaims),
		provider.WithTokenIntrospection(cdsrules.IntrospectingClient),
		provider.WithClientCredentialsGrant(),
		provider.WithTokenOptions(func(context.Context, *goidc.Grant, *goidc.Client) goidc.TokenOptions {
			return goidc.NewJWTTokenOptions(goidc.SigAlgPS256, 300)
		}),
		provider.WithStaticClients(&staticClient, &otherClient),
		provider.WithJWTLeewayTime(30),
		provider.WithJTIConsumer(consumeOnce()),
	)
	require.NoError(t, err)

	server.op = op

	srv.Config.Handler = clientCertMiddleware(op.Handler(
		cdsrules.RequestObjectClaims("/par"),
		cdsrules.TokenResponse("/token"),
		cdsrules.IntrospectionResponse("/introspect", server.store),
	))
	srv.StartTLS()
	t.Cleanup(srv.Close)

	recipient.http = srv.Client()
	transport := recipient.http.Transport.(*http.Transport)
	transport.TLSClientConfig.Certificates = []tls.Certificate{clientCertificate(t)}
	recipient.http.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	return server, recipient
}

func consumeOnce() goidc.ConsumeJTIFunc {
	var mu sync.Mutex

	seen := map[string]bool{}

	return func(_ context.Context, jti string) error {
		mu.Lock()
		defer mu.Unlock()

		if seen[jti] {
			return errors.New("the JWT ID is used again")
		}

		seen[jti] = true

		return nil
	}
}

func clientCertMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TLS.PeerCertificates) > 0 {
			r = r.WithContext(context.WithValue(r.Context(), ctxKeyClientCert{}, r.TLS.PeerCertificates[0]))
		}

		next.ServeHTTP(w, r)
	})
}

func clientCertFromContext(ctx context.Context) (*x509.Certificate, error) {
	cert, ok := ctx.Value(ctxKeyClientCert{}).(*x509.Certificate)
	if !ok {
		return nil, errors.New("the request has no client certificate")
	}

	return cert, nil
}

func (dr *dataRecipient) sign(key jose.JSONWebKey, claims map[string]any) string {
	dr.t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.PS256, Key: key},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	require.NoError(dr.t, err)
	token, err := jwt.Signed(signer).Claims(claims).Serialize()
	require.NoError(dr.t, err)

	return token
}

func (dr *dataRecipient) clientAssertion(id string) url.Values {
	now := time.Now()

	return url.Values{
		"client_id":             {id},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion": {dr.sign(dr.key, map[string]any{
			"iss": id,
			"sub": id,
			"aud": dr.server.issuer,
			"jti": rand.Text(),
			"iat": now.Unix(),
			"exp": now.Add(time.Minute).Unix(),
		})},
	}
}

// reply is an HTTP response with its body read and closed.
type reply struct {
	status int
	header http.Header
	body   map[string]any
}

func (dr *dataRecipient) post(path string, form url.Values) reply {
	dr.t.Helper()
	return dr.do(http.MethodPost, path, "application/x-www-form-urlencoded", "", strings.NewReader(form.Encode()))
}

func (dr *dataRecipient) do(method, path, contentType, bearer string, body io.Reader) reply {
	dr.t.Helper()
	req, err := http.NewRequestWithContext(dr.t.Context(), method, dr.server.issuer+path, body)
	require.NoError(dr.t, err)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	resp, err := dr.http.Do(req)
	require.NoError(dr.t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(dr.t, err)

	var out map[string]any

	_ = json.Unmarshal(data, &out)

	return reply{status: resp.StatusCode, header: resp.Header, body: out}
}

type authorisation struct {
	par      reply
	jarm     map[string]any
	verifier string
}

// authorise pushes a signed request object with the extra claims, then follows
// the authorisation redirect and verifies the JARM response.
func (dr *dataRecipient) authorise(extra map[string]any) authorisation {
	dr.t.Helper()

	verifier := rand.Text() + rand.Text()
	challenge := sha256.Sum256([]byte(verifier))
	now := time.Now()
	claims := map[string]any{
		"iss":                   clientID,
		"aud":                   dr.server.issuer,
		"client_id":             clientID,
		"response_type":         "code",
		"response_mode":         "jwt",
		"redirect_uri":          redirectURI,
		"scope":                 "openid " + dataScope,
		"state":                 "state-1",
		"nonce":                 "nonce-1",
		"code_challenge":        base64.RawURLEncoding.EncodeToString(challenge[:]),
		"code_challenge_method": "S256",
		"nbf":                   now.Unix(),
		"exp":                   now.Add(5 * time.Minute).Unix(),
	}
	maps.Copy(claims, extra)

	form := dr.clientAssertion(clientID)
	form.Set("request", dr.sign(dr.key, claims))

	par := dr.post("/par", form)
	if par.status != http.StatusCreated {
		return authorisation{par: par}
	}

	query := url.Values{"client_id": {clientID}, "request_uri": {par.body["request_uri"].(string)}}
	redirect := dr.do(http.MethodGet, "/authorize?"+query.Encode(), "", "", nil)
	location, err := url.Parse(redirect.header.Get("Location"))
	require.NoError(dr.t, err, "the authorisation response is a redirect")

	response := location.Query().Get("response")
	require.NotEmpty(dr.t, response, "the redirect has a JARM response: %s", location)

	parsed, err := jwt.ParseSigned(response, signingAlgs)
	require.NoError(dr.t, err)

	var jarm map[string]any
	require.NoError(
		dr.t,
		parsed.Claims(dr.server.key.Public().Key, &jarm),
		"the JARM response is signed with the server key",
	)

	return authorisation{par: par, jarm: jarm, verifier: verifier}
}

func (dr *dataRecipient) exchange(a authorisation) map[string]any {
	dr.t.Helper()

	resp := dr.exchangeAs(clientID, a)
	require.Equal(dr.t, http.StatusOK, resp.status, resp.body)

	return resp.body
}

func (dr *dataRecipient) exchangeAs(id string, a authorisation) reply {
	dr.t.Helper()
	form := dr.clientAssertion(id)
	form.Set("grant_type", "authorization_code")
	form.Set("code", a.jarm["code"].(string))
	form.Set("redirect_uri", redirectURI)
	form.Set("code_verifier", a.verifier)

	return dr.post("/token", form)
}

func (dr *dataRecipient) refreshAs(id, refreshToken string) reply {
	dr.t.Helper()
	form := dr.clientAssertion(id)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	return dr.post("/token", form)
}

func (dr *dataRecipient) introspectAs(id, token string) reply {
	dr.t.Helper()
	form := dr.clientAssertion(id)
	form.Set("token", token)

	return dr.post("/introspect", form)
}

// withNewCertificate gives the same software product after it replaced its client certificate.
func (dr *dataRecipient) withNewCertificate() *dataRecipient {
	dr.t.Helper()

	transport := dr.http.Transport.(*http.Transport).Clone()
	transport.TLSClientConfig.Certificates = []tls.Certificate{clientCertificate(dr.t)}
	renewed := *dr
	renewed.http = &http.Client{Transport: transport, CheckRedirect: dr.http.CheckRedirect}

	return &renewed
}

// claims verifies a JWT of the auth server and gives its claims.
func (s *authServer) claims(t *testing.T, token string) map[string]any {
	t.Helper()

	parsed, err := jwt.ParseSigned(token, signingAlgs)
	require.NoError(t, err)

	var claims map[string]any
	require.NoError(t, parsed.Claims(s.key.Public().Key, &claims))

	return claims
}
