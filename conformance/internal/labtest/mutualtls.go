package labtest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func selfSignedClientCertificate(t *testing.T) tls.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Token Test Active"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

type secureResponse struct {
	status     int
	header     http.Header
	body       string
	serverCert *x509.Certificate
}

// secureGet makes one request, because a TLS 1.3 client learns that the
// server refused its certificate only when it reads.
func secureGet(t *testing.T, cfg *tls.Config, target string) (secureResponse, error) {
	t.Helper()

	if cfg.RootCAs == nil {
		cfg.RootCAs = CDRCABundle(t)
	}

	httpClient := &http.Client{Timeout: RequestTimeout, Transport: &http.Transport{TLSClientConfig: cfg}}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	if err != nil {
		return secureResponse{}, err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return secureResponse{resp.StatusCode, resp.Header, string(body), resp.TLS.PeerCertificates[0]}, nil
}

// AssertAcceptsAClientCertificateFromTheCDRCA also checks the response of the mTLS listener for an unknown path.
func AssertAcceptsAClientCertificateFromTheCDRCA(t *testing.T, baseURL string, client tls.Certificate) {
	t.Helper()

	u, err := url.Parse(baseURL)
	require.NoError(t, err)

	resp, err := secureGet(
		t,
		&tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{client}},
		baseURL+"/unknown",
	)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.status)
	assert.Contains(t, resp.body, "urn:au-cds:error:cds-all:Resource/NotFound")
	assert.NotEmpty(t, resp.header.Get("x-fapi-interaction-id"))
	assert.Empty(t, resp.header.Get("Access-Control-Allow-Origin"), "the mTLS endpoints do not permit other origins")
	assert.Contains(t, resp.serverCert.DNSNames, u.Hostname())
}

func AssertRefusesAClientWithoutAValidCDRCertificate(t *testing.T, baseURL string) {
	t.Helper()

	selfSigned := selfSignedClientCertificate(t)
	revoked := RevokedClientCertificate(t)

	tests := map[string][]tls.Certificate{
		"no client certificate":     nil,
		"a self-signed certificate": {selfSigned},
		"a revoked certificate":     {revoked},
	}
	for name, certs := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := secureGet(
				t,
				&tls.Config{MinVersion: tls.VersionTLS12, Certificates: certs},
				baseURL+"/unknown",
			)

			assert.Error(t, err)
		})
	}
}

func AssertPresentsACertificateFromTheCDRCAOnly(t *testing.T, baseURL string, client tls.Certificate) {
	t.Helper()

	_, err := secureGet(
		t,
		&tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{client}, RootCAs: PublicRoot(t)},
		baseURL+"/unknown",
	)

	assert.Error(t, err, "the certificate does not chain to the mock public CA")
}

func AssertMutualTLSAcceptsOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(
	t *testing.T,
	baseURL string,
	client tls.Certificate,
) {
	t.Helper()

	accepted := map[string]*tls.Config{
		"TLS 1.3": {MinVersion: tls.VersionTLS13},
		"TLS 1.2 ECDHE-RSA-AES128-GCM": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		},
		"TLS 1.2 ECDHE-RSA-AES256-GCM": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384},
		},
	}
	refused := map[string]*tls.Config{
		"TLS 1.1": {MinVersion: tls.VersionTLS11, MaxVersion: tls.VersionTLS11},
		"TLS 1.2 ECDHE-RSA-CHACHA20": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256},
		},
		"TLS 1.2 ECDHE-RSA-AES128-SHA": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA},
		},
	}

	connect := func(t *testing.T, cfg *tls.Config) error {
		t.Helper()

		cfg.Certificates = []tls.Certificate{client}

		_, err := secureGet(t, cfg, baseURL+"/unknown")

		return err
	}

	for name, cfg := range accepted {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, connect(t, cfg))
		})
	}

	for name, cfg := range refused {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, connect(t, cfg))
		})
	}
}

func AssertRefusesTheClientCertificate(t *testing.T, baseURL string, client tls.Certificate) {
	t.Helper()

	_, err := secureGet(
		t,
		&tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{client}},
		baseURL+"/unknown",
	)

	assert.Error(t, err)
}
