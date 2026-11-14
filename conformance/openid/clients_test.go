//go:build openid

package openid

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

// These participants of services/register/seed/example.json have https://openid-runner:8443 as their JWKS host.
const (
	openIDTestEntity = "1c664017-a83c-4086-96f8-d61026d2df41"
	openIDTestBrand  = "37b32ae1-eb99-4774-ab0a-6fccf3a85725"
	// The suite requires a query in the redirect URI of the second client.
	redirectURI       = "https://openid:8443/test/a/" + suiteAlias + "/callback"
	secondRedirectURI = redirectURI + "?dummy1=lorem&dummy2=ipsum"
	// keyID is the key ID that labtest.NewAssertion gives.
	keyID = "rsa"
)

// client is one of the two clients of the suite: a software product of the Register, registered with the data holder.
type client struct {
	productID string
	// jwksPath is the path of the JWKS URI of the software product in the seed file.
	jwksPath     string
	redirectURI  string
	certVariable string
	keyVariable  string
	key          *rsa.PrivateKey
	clientID     string
}

func newClients(t *testing.T) []*client {
	t.Helper()

	clients := []*client{
		{
			productID: "aa29d742-9413-4d10-850c-70b310427da2", jwksPath: "/client-1/jwks", redirectURI: redirectURI,
			certVariable: "PKI_CLIENT_1_CERT", keyVariable: "PKI_CLIENT_1_KEY",
		},
		{
			productID:    "d07c4109-dc08-4319-9856-274afa9fa7e2",
			jwksPath:     "/client-2/jwks",
			redirectURI:  secondRedirectURI,
			certVariable: "PKI_CLIENT_2_CERT",
			keyVariable:  "PKI_CLIENT_2_KEY",
		},
	}

	for _, c := range clients {
		var err error

		c.key, err = rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
	}

	return clients
}

func (c *client) publicJWK() jose.JSONWebKey {
	return jose.JSONWebKey{Key: &c.key.PublicKey, KeyID: keyID, Algorithm: string(jose.PS256), Use: "sig"}
}

func (c *client) privateJWK() jose.JSONWebKey {
	return jose.JSONWebKey{Key: c.key, KeyID: keyID, Algorithm: string(jose.PS256), Use: "sig"}
}

func serveJWKS(t *testing.T, clients []*client) {
	t.Helper()

	cert, err := tls.LoadX509KeyPair(labtest.Env(t, "JWKS_TLS_CERT"), labtest.Env(t, "JWKS_TLS_KEY"))
	require.NoError(t, err)

	listener, err := tls.Listen(
		"tcp",
		labtest.Env(t, "JWKS_LISTEN_ADDRESS"),
		&tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
	)
	require.NoError(t, err)

	mux := http.NewServeMux()

	for _, c := range clients {
		set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{c.publicJWK()}}
		mux.HandleFunc("GET "+c.jwksPath, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(set)
		})
	}

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = srv.Serve(listener) }()

	t.Cleanup(func() { _ = srv.Close() })
}

type dataHolder struct {
	issuer          string
	resourceBaseURI string
}

// register registers the client certificate on the Register, and registers the client with the data holder with a
// new SSA, as a software product does.
func (c *client) register(t *testing.T) dataHolder {
	t.Helper()

	cert := labtest.ClientCertificate(t, c.certVariable, c.keyVariable)
	labtest.RegisterClientCertificate(t, openIDTestEntity, c.productID, c.certVariable)

	token := labtest.RegisterAccessToken(t, &labtest.RecipientKeys{RSA: c.key}, c.productID, cert)
	brand := labtest.RegisteredDataHolderBrand(t, token, cert)
	configuration := labtest.OpenIDConfiguration(t, brand.EndpointDetail.InfosecBaseURI)

	issuer, ok := configuration["issuer"].(string)
	require.True(t, ok, "the configuration has an issuer")
	endpoint, ok := configuration["registration_endpoint"].(string)
	require.True(t, ok, "the configuration has a registration endpoint")

	claims := labtest.RegistrationClaims(c.productID, issuer, labtest.SoftwareStatement(
		t, token, cert, openIDTestBrand, c.productID,
	))
	claims["redirect_uris"] = []string{c.redirectURI}

	resp := labtest.PostRegistration(
		t, endpoint, cert, "application/jwt", labtest.SignJWT(t, jose.PS256, c.key, keyID, claims),
	)

	defer func() { _ = resp.Body.Close() }()

	var registration struct {
		ClientID string `json:"client_id"`
	}

	require.Equal(t, http.StatusCreated, resp.StatusCode,
		"the registration of %s succeeds; the run needs a new lab: mise run reset && mise run up", c.productID)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&registration))
	c.clientID = registration.ClientID

	return dataHolder{issuer: issuer, resourceBaseURI: brand.EndpointDetail.ResourceBaseURI}
}

func (c *client) mtls(t *testing.T) map[string]string {
	t.Helper()

	return map[string]string{
		"cert": readFile(t, labtest.Env(t, c.certVariable)),
		"key":  readFile(t, labtest.Env(t, c.keyVariable)),
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // G304: the path is from the compose configuration of the lab
	require.NoError(t, err)

	return string(data)
}

// ready waits until the data holder has the statuses of the Register, because before that it gives no data.
func ready(t *testing.T) {
	t.Helper()

	target := labtest.Env(t, "DATAHOLDER_PROBE_URL") + "/readyz"

	require.Eventually(t, func() bool {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
		if err != nil {
			return false
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}

		_ = resp.Body.Close()

		return resp.StatusCode == http.StatusOK
	}, time.Minute, time.Second, "the data holder is ready")
}
