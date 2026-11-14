package labtest

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"
)

const (
	assertionType              = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
	registerScope              = "cdr-register:read"
	registerOpenIDConfig       = "/idp/.well-known/openid-configuration"
	registerDataRecipientsPath = "/v1/data-recipients/"
)

// RecipientKeys are the signing keys of the software products that a test package acts as.
type RecipientKeys struct {
	RSA *rsa.PrivateKey
	EC  *ecdsa.PrivateKey
}

var jwksServer struct {
	once sync.Once
	err  error
	keys RecipientKeys
}

// ServeJWKS starts the JWKS server of the test process one time, and the
// server runs until the process ends. Each test package has its own address,
// because the packages run at the same time. The server is also the CDR
// arrangement revocation endpoint of the software products whose recipient
// base URI is its address.
func ServeJWKS(t *testing.T, addressVariable string) *RecipientKeys {
	t.Helper()

	s := &jwksServer
	s.once.Do(func() { s.err = serveJWKS(t, addressVariable, &s.keys) })
	require.NoError(t, s.err, "start the JWKS server")

	return &s.keys
}

func serveJWKS(t *testing.T, addressVariable string, keys *RecipientKeys) error {
	t.Helper()

	var err error
	if keys.RSA, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return err
	}

	if keys.EC, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		return err
	}

	cert, err := tls.LoadX509KeyPair(Env(t, "JWKS_TLS_CERT"), Env(t, "JWKS_TLS_KEY"))
	if err != nil {
		return err
	}

	var listener net.Listener

	listener, err = tls.Listen(
		"tcp",
		Env(t, addressVariable),
		&tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
	)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /jwks", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
			{Key: &keys.RSA.PublicKey, KeyID: "rsa", Algorithm: "PS256", Use: "sig"},
			{Key: &keys.EC.PublicKey, KeyID: "ec", Algorithm: "ES256", Use: "sig"},
		}})
	})

	mux.HandleFunc("POST /arrangements/revoke", receiveNotice)

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = srv.Serve(listener) }()

	return nil
}

type ClientAssertion struct {
	Alg    jose.SignatureAlgorithm
	Key    crypto.Signer
	KID    string
	Claims map[string]any
}

// NewAssertion gives a valid PS256 assertion, which a test can change before it signs.
func NewAssertion(keys *RecipientKeys, clientID, audience string) ClientAssertion {
	now := time.Now()

	return ClientAssertion{Alg: jose.PS256, Key: keys.RSA, KID: "rsa", Claims: map[string]any{
		"iss": clientID,
		"sub": clientID,
		"aud": audience,
		"jti": rand.Text(),
		"iat": now.Unix(),
		"exp": now.Add(time.Minute).Unix(),
	}}
}

func TokenForm(t *testing.T, clientID, scope string, a ClientAssertion) url.Values {
	t.Helper()

	signed := SignJWT(t, a.Alg, a.Key, a.KID, a.Claims)

	return url.Values{
		"grant_type": {"client_credentials"}, "client_id": {clientID}, "scope": {scope},
		"client_assertion_type": {assertionType}, "client_assertion": {signed},
	}
}

type TokenReply struct {
	Status int
	Header http.Header
	Body   map[string]any
}

func RequestToken(t *testing.T, endpoint string, cert tls.Certificate, form url.Values) TokenReply {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := secureClient(t, cert).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// A token revocation that succeeds has no body (RFC 7009, section 2.2).
	var body map[string]any
	if len(data) > 0 {
		require.NoError(t, json.Unmarshal(data, &body), string(data))
	}

	return TokenReply{Status: resp.StatusCode, Header: resp.Header, Body: body}
}

// RegisterClientCertificate uses the operator interface of the Register, as
// the CDR Registrar registers the certificate of a software product in production.
func RegisterClientCertificate(t *testing.T, legalEntityID, softwareProductID, certVariable string) {
	t.Helper()

	pemData, err := os.ReadFile(Env(t, certVariable))
	require.NoError(t, err)

	target := Env(t, "REGISTER_OPERATOR_URL") + registerDataRecipientsPath + legalEntityID +
		"/software-products/" + softwareProductID + "/certificates"

	// The Register gives 409 when a different request changes the same data recipient at the same time, and the test
	// packages register certificates for software products of one data recipient at the same time.
	var (
		status int
		body   []byte
	)

	for range maxOperatorAttempts {
		status, body = postOperator(t, target, pemData)
		if status != http.StatusConflict {
			break
		}
	}

	require.Equal(t, http.StatusNoContent, status, string(body))
}

const maxOperatorAttempts = 5

// ChangeSoftwareProductStatus takes the action activate, deactivate or remove.
func ChangeSoftwareProductStatus(t *testing.T, legalEntityID, softwareProductID, action string) {
	t.Helper()

	target := Env(t, "REGISTER_OPERATOR_URL") + registerDataRecipientsPath + legalEntityID +
		"/software-products/" + softwareProductID + "/" + action

	var (
		status int
		body   []byte
	)

	for range maxOperatorAttempts {
		status, body = postOperator(t, target, nil)
		if status != http.StatusConflict {
			break
		}
	}

	require.Equal(t, http.StatusNoContent, status, string(body))
}

func postOperator(t *testing.T, target string, request []byte) (status int, body []byte) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, target, bytes.NewReader(request))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+Env(t, "REGISTER_OPERATOR_TOKEN"))

	resp, err := Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, body
}

func RegisterTokenEndpoint(t *testing.T) string {
	t.Helper()

	target := Env(t, "REGISTER_PUBLIC_URL") + registerOpenIDConfig
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	require.NoError(t, err)

	resp, err := Client(t).Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var cfg struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&cfg))

	return cfg.TokenEndpoint
}

// RegisterAccessToken needs a certificate that is registered for the software product.
func RegisterAccessToken(t *testing.T, keys *RecipientKeys, softwareProductID string, cert tls.Certificate) string {
	t.Helper()

	endpoint := RegisterTokenEndpoint(t)
	form := TokenForm(t, softwareProductID, registerScope, NewAssertion(keys, softwareProductID, endpoint))
	r := RequestToken(t, endpoint, cert, form)
	require.Equal(t, http.StatusOK, r.Status, r.Body)

	return r.Body["access_token"].(string)
}
