// Package authserver gives CDR authorisation servers, with an API that has no types of the OpenID Provider library.
package authserver

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"slices"
	"strings"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
)

const (
	registerIssuerPath           = "/idp"
	registerConfigPath           = registerIssuerPath + "/.well-known/openid-configuration"
	registerJWKSPath             = registerConfigPath + "/jwks"
	registerEndpointPath         = registerIssuerPath + "/connect"
	registerTokenPath            = registerEndpointPath + "/token"
	registerScope                = "cdr-register:read"
	registerTokenLifetimeSeconds = 300
	// A JTIStore must keep a JWT ID for longer than this lifetime plus clockLeewaySeconds.
	registerAssertionLifetimeSeconds = 600
)

const clockLeewaySeconds = 30

var (
	ErrClientNotFound = errors.New("the client does not exist")
	ErrJTIUsed        = errors.New("the JWT ID is used again")
)

type Client struct {
	ID      string
	JWKSURI string
	Active  bool
	// CertificateThumbprints have the format of the RFC 8705 x5t#S256 value.
	CertificateThumbprints []string
}

type ClientStore interface {
	// Client gives ErrClientNotFound when the client does not exist.
	Client(ctx context.Context, id string) (Client, error)
}

type JTIStore interface {
	// Consume gives ErrJTIUsed when the JWT ID was consumed before.
	Consume(ctx context.Context, jti string) error
}

type SigningKey struct {
	ID  string
	Key *rsa.PrivateKey
}

type RegisterConfig struct {
	PublicBaseURL string
	SecureBaseURL string
	Clients       ClientStore
	JTIs          JTIStore
	// SigningKeys gives the keys of the access tokens. The first key signs. The JWKS publishes all keys.
	SigningKeys func(context.Context) ([]SigningKey, error)
	// JWKSTransport is optional: nil is http.DefaultTransport.
	JWKSTransport http.RoundTripper
}

type Register struct {
	cfg      RegisterConfig
	issuer   string
	provider http.Handler
}

func NewRegister(cfg RegisterConfig) (*Register, error) {
	if cfg.PublicBaseURL == "" || cfg.SecureBaseURL == "" || cfg.Clients == nil || cfg.JTIs == nil ||
		cfg.SigningKeys == nil {
		return nil, errors.New("authserver: all fields of RegisterConfig except JWKSTransport are required")
	}

	cfg.PublicBaseURL = strings.TrimSuffix(cfg.PublicBaseURL, "/")
	cfg.SecureBaseURL = strings.TrimSuffix(cfg.SecureBaseURL, "/")
	r := &Register{cfg: cfg, issuer: cfg.PublicBaseURL + registerIssuerPath}
	jwksClient := NewJWKSClient(cfg.JWKSTransport)

	op, err := provider.New(
		provider.Config{
			Issuer:      r.issuer,
			Manager:     noGrants{},
			JWKS:        privateJWKS(cfg.SigningKeys),
			IDTokenAlgs: []goidc.SignatureAlgorithm{goidc.SigAlgPS256},
		},
		provider.WithPathPrefix(registerEndpointPath),
		provider.WithScopes(goidc.NewScope(registerScope)),
		provider.WithMTLS(provider.MTLSConfig{Host: cfg.SecureBaseURL, ClientCert: clientCertFromContext},
			provider.WithMTLSTokenBinding(provider.WithMTLSTokenBindingRequired())),
		provider.WithPrivateKeyJWTAuthn(goidc.SigAlgPS256, goidc.SigAlgES256),
		provider.WithClientCredentialsGrant(),
		provider.WithTokenOptions(func(context.Context, *goidc.Grant, *goidc.Client) goidc.TokenOptions {
			return goidc.NewJWTTokenOptions(goidc.SigAlgPS256, registerTokenLifetimeSeconds)
		}),
		// The library finds clients that are not static only through its DCR
		// manager. RegisterSecure does not serve the DCR endpoint.
		provider.WithDCR(softwareProducts{cfg.Clients}),
		provider.WithGrantHandler(r.checkClient),
		provider.WithJWTLifetime(registerAssertionLifetimeSeconds),
		provider.WithJWTLeewayTime(clockLeewaySeconds),
		provider.WithJTIConsumer(func(ctx context.Context, jti string) error { return cfg.JTIs.Consume(ctx, jti) }),
		provider.WithHTTPClientFunc(func(context.Context) *http.Client { return jwksClient }),
		provider.WithErrorHandler(logRefusal),
	)
	if err != nil {
		return nil, err
	}

	r.provider = op.Handler()

	return r, nil
}

func (r *Register) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("GET "+registerConfigPath, r.configuration)
	mux.HandleFunc("GET "+registerJWKSPath, jwksHandler(r.cfg.SigningKeys))
}

// RegisterSecure needs a listener that verifies the client certificate.
func (r *Register) RegisterSecure(mux *http.ServeMux) {
	mux.HandleFunc("POST "+registerTokenPath, tokenHandler(r.provider))
}

type ctxKeyClientCert struct{}

func clientCertFromContext(ctx context.Context) (*x509.Certificate, error) {
	cert, ok := ctx.Value(ctxKeyClientCert{}).(*x509.Certificate)
	if !ok {
		return nil, errors.New("the request has no client certificate")
	}

	return cert, nil
}

// The RFC 6749 section 5.2 error codes.
var tokenErrorCodes = []string{
	"invalid_request",
	"invalid_client",
	"invalid_grant",
	"unauthorized_client",
	"unsupported_grant_type",
	"invalid_scope",
}

type tokenError struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// tokenHandler gives the token endpoint of provider the client certificate, and makes each response obey RFC 6749.
func tokenHandler(provider http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		rec := &recorder{header: http.Header{}, status: http.StatusOK}
		provider.ServeHTTP(rec, withClientCert(req))

		writeTokenResponse(w, rec)
	}
}

func withClientCert(req *http.Request) *http.Request {
	if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
		return req
	}

	return req.WithContext(context.WithValue(req.Context(), ctxKeyClientCert{}, req.TLS.PeerCertificates[0]))
}

func writeTokenResponse(w http.ResponseWriter, rec *recorder) {
	// RFC 6749 section 5.1 requires these headers, also on errors (section 5.2).
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	if rec.status == http.StatusOK {
		_, _ = w.Write(rec.body.Bytes())
		return
	}
	// The library has error codes that RFC 6749 section 5.2 does not have.
	var e tokenError

	_ = json.Unmarshal(rec.body.Bytes(), &e)

	status := rec.status
	switch {
	case status >= http.StatusInternalServerError:
		status, e = http.StatusInternalServerError, tokenError{Code: "server_error"}
	case !slices.Contains(tokenErrorCodes, e.Code):
		status, e = http.StatusUnauthorized, tokenError{Code: "invalid_client", Description: "invalid client"}
	// RFC 6749 section 5.2 gives 401 only for invalid_client, and the library also gives it for unauthorized_client.
	case e.Code != "invalid_client":
		status = http.StatusBadRequest
	}

	data, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(status)
	_, _ = w.Write(data)
}

type recorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *recorder) Header() http.Header         { return r.header }
func (r *recorder) WriteHeader(status int)      { r.status = status }
func (r *recorder) Write(b []byte) (int, error) { return r.body.Write(b) }

// checkClient runs after the library authenticates the client assertion.
func (r *Register) checkClient(ctx context.Context, _ goidc.GrantType, grant *goidc.Grant) error {
	client, err := r.cfg.Clients.Client(ctx, grant.ClientID)
	if err != nil {
		return err
	}

	cert, err := clientCertFromContext(ctx)
	if err != nil {
		return goidc.WrapError(goidc.ErrorCodeInvalidClient, "invalid client", err)
	}

	if !slices.Contains(client.CertificateThumbprints, thumbprint(cert)) {
		return goidc.NewError(goidc.ErrorCodeInvalidClient, "the client certificate is not registered for the client")
	}

	if !client.Active {
		return goidc.NewError(goidc.ErrorCodeInvalidClient, "the client is not active")
	}

	return nil
}

// logRefusal logs only the code and the description, because the cause can contain client data.
func logRefusal(ctx context.Context, err error) {
	if e, ok := errors.AsType[goidc.Error](err); ok {
		slog.InfoContext(ctx, "request refused", "error.type", string(e.Code), "error.message", e.Description)
		return
	}

	slog.ErrorContext(ctx, "request failed", "error.type", "server_error")
}

type softwareProducts struct{ store ClientStore }

func (s softwareProducts) Client(ctx context.Context, id string) (*goidc.Client, error) {
	c, err := s.store.Client(ctx, id)
	if errors.Is(err, ErrClientNotFound) {
		return nil, goidc.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &goidc.Client{
		ID:               c.ID,
		JWKSURI:          c.JWKSURI,
		ScopeIDs:         registerScope,
		GrantTypes:       []goidc.GrantType{goidc.GrantClientCredentials},
		TokenAuthnMethod: goidc.AuthnMethodPrivateKeyJWT,
	}, nil
}

func (softwareProducts) SaveClient(context.Context, *goidc.Client) error {
	return errors.New("authserver: the Register does not register clients")
}

func (softwareProducts) DeleteClient(context.Context, string) error {
	return errors.New("authserver: the Register does not register clients")
}

// noGrants keeps nothing: the access tokens are JWTs, and the Register has no endpoint that reads a grant.
type noGrants struct{}

func (noGrants) SaveGrant(context.Context, *goidc.Grant) error { return nil }

func (noGrants) Grant(context.Context, string) (*goidc.Grant, error) { return nil, goidc.ErrNotFound }

// privateJWKS gives the library the keys with their private parts: the first key signs.
func privateJWKS(
	signingKeys func(context.Context) ([]SigningKey, error),
) func(context.Context) (goidc.JSONWebKeySet, error) {
	return func(ctx context.Context) (goidc.JSONWebKeySet, error) {
		keys, err := signingKeys(ctx)
		if err != nil {
			return goidc.JSONWebKeySet{}, err
		}

		if len(keys) == 0 {
			return goidc.JSONWebKeySet{}, errors.New("authserver: the authorisation server has no signing key")
		}

		var set goidc.JSONWebKeySet
		for _, k := range keys {
			set.Keys = append(
				set.Keys,
				goidc.JSONWebKey{Key: k.Key, KeyID: k.ID, Algorithm: string(goidc.SigAlgPS256), Use: "sig"},
			)
		}

		return set, nil
	}
}

type configurationJSON struct {
	Issuer                       string   `json:"issuer"`
	JWKSURI                      string   `json:"jwks_uri"`
	TokenEndpoint                string   `json:"token_endpoint"`
	ClaimsSupported              []string `json:"claims_supported"`
	IDTokenSigningAlgs           []string `json:"id_token_signing_alg_values_supported"`
	SubjectTypes                 []string `json:"subject_types_supported"`
	CodeChallengeMethods         []string `json:"code_challenge_methods_supported"`
	Scopes                       []string `json:"scopes_supported"`
	ResponseTypes                []string `json:"response_types_supported"`
	GrantTypes                   []string `json:"grant_types_supported"`
	TokenEndpointAuthMethods     []string `json:"token_endpoint_auth_methods_supported"`
	CertificateBoundTokens       bool     `json:"tls_client_certificate_bound_access_tokens"`
	TokenEndpointAuthSigningAlgs []string `json:"token_endpoint_auth_signing_alg_values_supported"`
}

// configuration writes the document here, because the library uses a
// different path and omits fields that the CDS requires.
func (r *Register) configuration(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, configurationJSON{
		Issuer:                       r.issuer,
		JWKSURI:                      r.cfg.PublicBaseURL + registerJWKSPath,
		TokenEndpoint:                r.cfg.SecureBaseURL + registerTokenPath,
		ClaimsSupported:              []string{"iss", "sub", "aud", "exp", "iat", "jti", "client_id", "scope", "cnf"},
		IDTokenSigningAlgs:           []string{string(goidc.SigAlgPS256)},
		SubjectTypes:                 []string{"public"},
		CodeChallengeMethods:         []string{},
		Scopes:                       []string{registerScope},
		ResponseTypes:                []string{"token"},
		GrantTypes:                   []string{string(goidc.GrantClientCredentials)},
		TokenEndpointAuthMethods:     []string{string(goidc.AuthnMethodPrivateKeyJWT)},
		CertificateBoundTokens:       true,
		TokenEndpointAuthSigningAlgs: []string{string(goidc.SigAlgPS256), string(goidc.SigAlgES256)},
	})
}

// JWK has the fields that the CDS JWK schema requires.
type JWK struct {
	KeyType string   `json:"kty"`
	KeyID   string   `json:"kid"`
	Alg     string   `json:"alg"`
	KeyOps  []string `json:"key_ops"`
	N       string   `json:"n"`
	E       string   `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func PublicJWKS(keys []SigningKey) JWKS {
	set := JWKS{Keys: []JWK{}}
	for _, k := range keys {
		set.Keys = append(set.Keys, JWK{
			KeyType: "RSA", KeyID: k.ID, Alg: string(goidc.SigAlgPS256), KeyOps: []string{"verify"},
			N: base64.RawURLEncoding.EncodeToString(k.Key.N.Bytes()),
			E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(k.Key.E)).Bytes()),
		})
	}

	return set
}

func jwksHandler(signingKeys func(context.Context) ([]SigningKey, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		keys, err := signingKeys(req.Context())
		if err != nil {
			slog.ErrorContext(req.Context(), "signing keys not available", "error.type", "server_error")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		writeJSON(w, PublicJWKS(keys))
	}
}

func writeJSON(w http.ResponseWriter, body any) {
	// Clients must check for a new document on each use, for example after a key rotation.
	w.Header().Set("Cache-Control", "no-cache")
	jsonhttp.Write(w, http.StatusOK, body)
}
