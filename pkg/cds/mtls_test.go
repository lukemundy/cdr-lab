package cds_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

type testCA struct {
	cert *x509.Certificate
	key  *rsa.PrivateKey
}

func newTestCA(t *testing.T) testCA {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Test CDR CA"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return testCA{cert: cert, key: key}
}

func (ca testCA) pool() *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(ca.cert)

	return pool
}

func (ca testCA) issue(t *testing.T, usage x509.ExtKeyUsage, names ...string) tls.Certificate {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "leaf"}, DNSNames: names,
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{usage},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	require.NoError(t, err)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

// mutualTLSRequest gives the error of one request, because a TLS 1.3 client learns
// that the server refused its certificate only when it reads.
func mutualTLSRequest(
	t *testing.T,
	ca testCA,
	check func(context.Context, *x509.Certificate) error,
	client *tls.Certificate,
) error {
	t.Helper()

	srv := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }),
	)
	srv.TLS = cds.MutualTLSConfig(
		t.Context(),
		ca.issue(t, x509.ExtKeyUsageServerAuth, "register"),
		ca.pool(),
		check,
	)
	srv.StartTLS()
	t.Cleanup(srv.Close)

	cfg := &tls.Config{RootCAs: ca.pool(), ServerName: "register", MinVersion: tls.VersionTLS12}
	if client != nil {
		cfg.Certificates = []tls.Certificate{*client}
	}

	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: cfg}}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	return nil
}

func notRevoked(context.Context, *x509.Certificate) error { return nil }

// Verifies: CDSL-0044
func TestTheSecureAPIsAcceptAClientCertificateFromTheCDRCA(t *testing.T) {
	ca := newTestCA(t)
	client := ca.issue(t, x509.ExtKeyUsageClientAuth)

	var checked *x509.Certificate

	err := mutualTLSRequest(t, ca, func(_ context.Context, cert *x509.Certificate) error {
		checked = cert
		return nil
	}, &client)

	require.NoError(t, err)
	require.NotNil(t, checked, "the revocation status is checked")
	assert.Equal(t, client.Certificate[0], checked.Raw)
}

// Verifies: CDSL-0044, CDSL-0047
func TestTheSecureAPIsRefuseOtherClients(t *testing.T) {
	ca := newTestCA(t)
	fromCA, fromOtherCA := ca.issue(t, x509.ExtKeyUsageClientAuth), newTestCA(t).issue(t, x509.ExtKeyUsageClientAuth)
	serverCert := ca.issue(t, x509.ExtKeyUsageServerAuth)

	tests := map[string]struct {
		client *tls.Certificate
		check  func(context.Context, *x509.Certificate) error
	}{
		"no client certificate":             {nil, notRevoked},
		"a certificate from a different CA": {&fromOtherCA, notRevoked},
		"a server certificate":              {&serverCert, notRevoked},
		"a certificate that fails the revocation check": {
			&fromCA,
			func(context.Context, *x509.Certificate) error { return errors.New("revoked") },
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := mutualTLSRequest(t, ca, tt.check, tt.client)

			assert.Error(t, err)
		})
	}
}

// Verifies: CDSL-0041, CDSL-0042
func TestTheSecureAPIsUseTheTLSVersionsAndSuitesOfThePublicAPIs(t *testing.T) {
	ca := newTestCA(t)
	server := ca.issue(t, x509.ExtKeyUsageServerAuth, "register")

	secure, public := cds.MutualTLSConfig(t.Context(), server, ca.pool(), notRevoked), cds.TLSConfig(server)

	assert.Equal(t, public.MinVersion, secure.MinVersion)
	assert.Equal(t, public.CipherSuites, secure.CipherSuites)
}

// Verifies: CDSL-0084
func TestAResumedSessionGetsTheRevocationCheckAgain(t *testing.T) {
	ca := newTestCA(t)
	client := ca.issue(t, x509.ExtKeyUsageClientAuth)

	var (
		checks  atomic.Int32
		revoked atomic.Bool
	)

	srv := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }),
	)
	srv.TLS = cds.MutualTLSConfig(
		t.Context(),
		ca.issue(t, x509.ExtKeyUsageServerAuth, "register"),
		ca.pool(),
		func(context.Context, *x509.Certificate) error {
			checks.Add(1)

			if revoked.Load() {
				return errors.New("revoked")
			}

			return nil
		},
	)
	srv.StartTLS()
	t.Cleanup(srv.Close)

	cfg := &tls.Config{
		RootCAs: ca.pool(), ServerName: "register", MinVersion: tls.VersionTLS12,
		Certificates: []tls.Certificate{client}, ClientSessionCache: tls.NewLRUClientSessionCache(1),
	}
	connect := func() (*http.Response, error) {
		httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: cfg, DisableKeepAlives: true}}
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
		require.NoError(t, err)

		return httpClient.Do(req)
	}

	first, err := connect()
	require.NoError(t, err)
	require.NoError(t, first.Body.Close())
	assert.False(t, first.TLS.DidResume)

	resumed, err := connect()
	require.NoError(t, err)
	require.NoError(t, resumed.Body.Close())
	require.True(t, resumed.TLS.DidResume, "the client resumes the session")
	assert.EqualValues(t, 2, checks.Load(), "the resumption gets the revocation check")

	revoked.Store(true)

	refused, err := connect()
	if err == nil {
		_ = refused.Body.Close()
	}

	assert.Error(t, err, "a revoked certificate cannot resume")
}

// Verifies: CDSL-0086
func TestARevocationCheckHasALimitOf5Seconds(t *testing.T) {
	ca := newTestCA(t)
	client := ca.issue(t, x509.ExtKeyUsageClientAuth)

	var limit time.Duration

	err := mutualTLSRequest(t, ca, func(ctx context.Context, _ *x509.Certificate) error {
		deadline, ok := ctx.Deadline()
		require.True(t, ok)

		limit = time.Until(deadline)

		return nil
	}, &client)

	require.NoError(t, err)
	assert.InDelta(t, 5*time.Second, limit, float64(time.Second))
}
