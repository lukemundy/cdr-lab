// Package labtest has the helpers that the conformance tests of all components share.
package labtest

import (
	"crypto/tls"
	"crypto/x509"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// No CDR timing rule is under test, so these limits allow for a stop of the lab of up to 15 s on a loaded CI node.
const (
	RequestTimeout = 30 * time.Second
	// PollLimit is the limit of a poll for which no requirement or lab configuration gives a time.
	PollLimit = 15 * time.Second
)

// Client trusts only the mock public root CA, as a client of the production public APIs trusts only the public CAs.
func Client(t *testing.T) *http.Client {
	t.Helper()

	return &http.Client{
		Timeout:   RequestTimeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: PublicRoot(t), MinVersion: tls.VersionTLS12}},
	}
}

func PublicRoot(t *testing.T) *x509.CertPool {
	t.Helper()

	return certPool(t, "PKI_PUBLIC_ROOT")
}

func CDRCABundle(t *testing.T) *x509.CertPool {
	t.Helper()

	return certPool(t, "PKI_CA_BUNDLE")
}

// ClientCertificate loads a certificate and its key from the files that two environment variables give.
func ClientCertificate(t *testing.T, certVariable, keyVariable string) tls.Certificate {
	t.Helper()

	cert, err := tls.LoadX509KeyPair(Env(t, certVariable), Env(t, keyVariable))
	if err != nil {
		t.Fatal(err)
	}

	return cert
}

func certPool(t *testing.T, variable string) *x509.CertPool {
	t.Helper()

	path := Env(t, variable)

	data, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatal(err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		t.Fatalf("%s has no PEM certificate", path)
	}

	return pool
}

// Env fails and does not skip, so that a missing configuration cannot pass in CI.
func Env(t *testing.T, variable string) string {
	t.Helper()

	value := os.Getenv(variable)
	if value == "" {
		t.Fatalf("set %s", variable)
	}

	return value
}
