package cds_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

func serverCertificate(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "register"}, DNSNames: []string{"register"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, BasicConstraintsValid: true, IsCA: true,
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	pool := x509.NewCertPool()
	pool.AddCert(cert)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, pool
}

func publicTLSServer(t *testing.T) (*httptest.Server, *x509.CertPool) {
	t.Helper()
	cert, roots := serverCertificate(t)
	srv := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }),
	)
	srv.TLS = cds.TLSConfig(cert)
	srv.StartTLS()
	t.Cleanup(srv.Close)

	return srv, roots
}

func handshake(t *testing.T, srv *httptest.Server, cfg *tls.Config) (*tls.Conn, error) {
	t.Helper()

	conn, err := (&tls.Dialer{Config: cfg}).DialContext(t.Context(), "tcp", srv.Listener.Addr().String())
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() { _ = conn.Close() })

	return conn.(*tls.Conn), nil
}

func TestThePublicAPIsAcceptTLS13AndTheTLS12SuitesThatTheCDSPermits(t *testing.T) {
	srv, roots := publicTLSServer(t)

	tests := map[string]*tls.Config{
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
	for name, cfg := range tests {
		t.Run(name, func(t *testing.T) {
			cfg.RootCAs, cfg.ServerName = roots, "register"

			_, err := handshake(t, srv, cfg)

			assert.NoError(t, err)
		})
	}
}

func TestThePublicAPIsRefuseOtherTLSVersionsAndSuites(t *testing.T) {
	srv, roots := publicTLSServer(t)

	tests := map[string]*tls.Config{
		"TLS 1.0": {MinVersion: tls.VersionTLS10, MaxVersion: tls.VersionTLS10},
		"TLS 1.1": {MinVersion: tls.VersionTLS11, MaxVersion: tls.VersionTLS11},
		"TLS 1.2 ECDHE-RSA-CHACHA20": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256},
		},
		"TLS 1.2 ECDHE-RSA-AES128-CBC": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA},
		},
		"TLS 1.2 ECDHE-RSA-AES128-SHA256": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256},
		},
		"TLS 1.2 RSA-AES128-GCM": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_RSA_WITH_AES_128_GCM_SHA256},
		},
	}
	for name, cfg := range tests {
		t.Run(name, func(t *testing.T) {
			cfg.RootCAs, cfg.ServerName = roots, "register"

			_, err := handshake(t, srv, cfg)

			assert.Error(t, err)
		})
	}
}

func TestThePublicAPIsDoNotRequestAClientCertificate(t *testing.T) {
	srv, roots := publicTLSServer(t)
	requested := false

	_, err := handshake(t, srv, &tls.Config{
		RootCAs: roots, ServerName: "register", MinVersion: tls.VersionTLS12,
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			requested = true
			return &tls.Certificate{}, nil
		},
	})

	require.NoError(t, err)
	assert.False(t, requested)
}
