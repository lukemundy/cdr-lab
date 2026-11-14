package labtest

import (
	"crypto/tls"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dialPublic(t *testing.T, baseURL string, cfg *tls.Config) (*tls.Conn, error) {
	t.Helper()

	u, err := url.Parse(baseURL)
	require.NoError(t, err)
	require.Equal(t, "https", u.Scheme, "the public APIs use TLS")

	if cfg.RootCAs == nil {
		cfg.RootCAs = PublicRoot(t)
	}

	cfg.ServerName = u.Hostname()
	dialer := &tls.Dialer{Config: cfg}

	conn, err := dialer.DialContext(t.Context(), "tcp", u.Host)
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() { _ = conn.Close() })

	return conn.(*tls.Conn), nil
}

func AssertAcceptsOnlyTheTLSVersionsAndSuitesThatTheCDSPermits(t *testing.T, baseURL string) {
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
		"TLS 1.0": {MinVersion: tls.VersionTLS10, MaxVersion: tls.VersionTLS10},
		"TLS 1.1": {MinVersion: tls.VersionTLS11, MaxVersion: tls.VersionTLS11},
		"TLS 1.2 ECDHE-RSA-CHACHA20": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256},
		},
		"TLS 1.2 ECDHE-RSA-AES128-SHA": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA},
		},
		"TLS 1.2 ECDHE-RSA-AES128-SHA256": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256},
		},
		"TLS 1.2 RSA-AES128-GCM (no ECDHE)": {
			MaxVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_RSA_WITH_AES_128_GCM_SHA256},
		},
	}

	for name, cfg := range accepted {
		t.Run(name, func(t *testing.T) {
			_, err := dialPublic(t, baseURL, cfg)

			assert.NoError(t, err)
		})
	}

	for name, cfg := range refused {
		t.Run(name, func(t *testing.T) {
			_, err := dialPublic(t, baseURL, cfg)

			assert.Error(t, err)
		})
	}
}

func AssertPresentsACertificateFromThePublicCAOnly(t *testing.T, baseURL string) {
	t.Helper()

	_, err := dialPublic(t, baseURL, &tls.Config{MinVersion: tls.VersionTLS12})
	require.NoError(t, err, "a client that trusts only the public root connects")

	_, err = dialPublic(t, baseURL, &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: CDRCABundle(t)})
	assert.Error(t, err, "the certificate does not chain to the mock CDR CA")
}

func AssertDoesNotRequestAClientCertificate(t *testing.T, baseURL string) {
	t.Helper()

	requested := false

	_, err := dialPublic(t, baseURL, &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			requested = true
			return &tls.Certificate{}, nil
		},
	})

	require.NoError(t, err)
	assert.False(t, requested)
}
