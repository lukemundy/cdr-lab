package cds_test

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

const testInterval = 10 * time.Millisecond

func newServerCertificate(t *testing.T, cert tls.Certificate) *cds.ServerCertificate {
	t.Helper()

	c, err := cds.NewServerCertificate(t.Context(), "public", cert)
	require.NoError(t, err)

	return c
}

func certificatePEM(t *testing.T, cert tls.Certificate) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key})
}

func writeFile(t *testing.T, name string, data []byte) {
	t.Helper()
	require.NoError(t, os.WriteFile(name, data, 0o600))
}

func writeCertificateFiles(t *testing.T, dir string, cert tls.Certificate) {
	t.Helper()

	certPEM, keyPEM := certificatePEM(t, cert)
	writeFile(t, filepath.Join(dir, "tls.crt"), certPEM)
	writeFile(t, filepath.Join(dir, "tls.key"), keyPEM)
}

// secretDirectory has the layout of a Kubernetes Secret that is mounted as files.
type secretDirectory struct {
	dir      string
	versions int
}

func newSecretDirectory(t *testing.T, cert tls.Certificate) *secretDirectory {
	t.Helper()

	s := &secretDirectory{dir: t.TempDir()}
	s.update(t, cert)

	for _, name := range []string{"tls.crt", "tls.key"} {
		require.NoError(t, os.Symlink(filepath.Join("..data", name), filepath.Join(s.dir, name)))
	}

	return s
}

// update writes a new version directory, then changes the ..data symlink to it in one rename, as the kubelet does.
func (s *secretDirectory) update(t *testing.T, cert tls.Certificate) {
	t.Helper()

	s.versions++
	version := fmt.Sprintf("..version_%d", s.versions)
	require.NoError(t, os.Mkdir(filepath.Join(s.dir, version), 0o700))
	writeCertificateFiles(t, filepath.Join(s.dir, version), cert)

	require.NoError(t, os.Symlink(version, filepath.Join(s.dir, "..data_tmp")))
	require.NoError(t, os.Rename(filepath.Join(s.dir, "..data_tmp"), filepath.Join(s.dir, "..data")))
}

func certificateServer(t *testing.T, cert *cds.ServerCertificate) *httptest.Server {
	t.Helper()

	srv := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }),
	)
	srv.TLS = cds.TLSConfig(cert)
	srv.StartTLS()
	t.Cleanup(srv.Close)

	return srv
}

// servedCertificate does not fail the test, because assert.Eventually calls it in a different goroutine.
func servedCertificate(t *testing.T, srv *httptest.Server, ca testCA) ([]byte, error) {
	t.Helper()

	conn, err := handshake(t, srv, &tls.Config{
		RootCAs: ca.pool(), ServerName: "register", MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return nil, err
	}

	defer func() { _ = conn.Close() }()

	return conn.ConnectionState().PeerCertificates[0].Raw, nil
}

func requireServes(t *testing.T, srv *httptest.Server, ca testCA, want tls.Certificate) {
	t.Helper()

	got, err := servedCertificate(t, srv, ca)
	require.NoError(t, err)
	require.Equal(t, want.Certificate[0], got)
}

func assertEventuallyServes(t *testing.T, srv *httptest.Server, ca testCA, want tls.Certificate) {
	t.Helper()

	assert.Eventually(t, func() bool {
		got, err := servedCertificate(t, srv, ca)

		return err == nil && bytes.Equal(got, want.Certificate[0])
	}, 5*time.Second, testInterval)
}

// logBuffer is safe for the watch goroutine and the test at the same time.
type logBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *logBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *logBuffer) entries(t *testing.T, level string) []map[string]any {
	t.Helper()
	b.mu.Lock()
	defer b.mu.Unlock()

	var entries []map[string]any

	for line := range strings.Lines(b.buf.String()) {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("log line %q is not JSON: %v", line, err)
		}

		if entry["level"] == level {
			entries = append(entries, entry)
		}
	}

	return entries
}

func captureLogs(t *testing.T) *logBuffer {
	t.Helper()

	logs := &logBuffer{}
	previous := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return logs
}

// Verifies: CDSL-0088, CDSL-0089
func TestANewHandshakeGetsTheCertificateOfChangedFiles(t *testing.T) {
	ca := newTestCA(t)
	first, renewed := ca.issue(
		t,
		x509.ExtKeyUsageServerAuth,
		"register",
	), ca.issue(
		t,
		x509.ExtKeyUsageServerAuth,
		"register",
	)

	t.Run("files that change in place", func(t *testing.T) {
		dir := t.TempDir()
		writeCertificateFiles(t, dir, first)
		cert, err := cds.WatchCertificateFilesEvery(
			t.Context(), "public", filepath.Join(dir, "tls.crt"), filepath.Join(dir, "tls.key"), testInterval,
		)
		require.NoError(t, err)
		srv := certificateServer(t, cert)
		requireServes(t, srv, ca, first)

		writeCertificateFiles(t, dir, renewed)

		assertEventuallyServes(t, srv, ca, renewed)
	})

	t.Run("a Kubernetes Secret that changes by a symlink swap", func(t *testing.T) {
		secret := newSecretDirectory(t, first)
		cert, err := cds.WatchCertificateFilesEvery(
			t.Context(),
			"public",
			filepath.Join(secret.dir, "tls.crt"),
			filepath.Join(secret.dir, "tls.key"),
			testInterval,
		)
		require.NoError(t, err)
		srv := certificateServer(t, cert)
		requireServes(t, srv, ca, first)

		secret.update(t, renewed)

		assertEventuallyServes(t, srv, ca, renewed)
	})
}

// Verifies: CDSL-0089
func TestTheWatchReadsTheFilesAgainWithinOneMinute(t *testing.T) {
	assert.LessOrEqual(t, cds.CertificateCheckInterval, time.Minute)
}

// Verifies: CDSL-0088, CDSL-0090
func TestANewHandshakeGetsACertificateThatTheServiceReplaces(t *testing.T) {
	ca := newTestCA(t)
	cert := newServerCertificate(t, ca.issue(t, x509.ExtKeyUsageServerAuth, "register"))
	srv := certificateServer(t, cert)
	replacement := ca.issue(t, x509.ExtKeyUsageServerAuth, "register")

	require.NoError(t, cert.Replace(t.Context(), replacement))

	requireServes(t, srv, ca, replacement)
}

// Verifies: CDSL-0091
func TestAnOpenConnectionKeepsTheCertificateOfItsHandshake(t *testing.T) {
	ca := newTestCA(t)
	first := ca.issue(t, x509.ExtKeyUsageServerAuth, "register")
	cert := newServerCertificate(t, first)
	srv := certificateServer(t, cert)
	conn, err := handshake(
		t,
		srv,
		&tls.Config{RootCAs: ca.pool(), ServerName: "register", MinVersion: tls.VersionTLS12},
	)
	require.NoError(t, err)

	require.NoError(t, cert.Replace(t.Context(), ca.issue(t, x509.ExtKeyUsageServerAuth, "register")))

	_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: register\r\n\r\n"))
	require.NoError(t, err)
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, first.Certificate[0], conn.ConnectionState().PeerCertificates[0].Raw)
}

// Verifies: CDSL-0092, CDSL-0093
//
//nolint:paralleltest // the test changes the default logger
func TestFilesThatDoNotAgreeLeaveTheLastValidCertificateInUse(t *testing.T) {
	logs := captureLogs(t)
	ca := newTestCA(t)
	first, renewed := ca.issue(
		t,
		x509.ExtKeyUsageServerAuth,
		"register",
	), ca.issue(
		t,
		x509.ExtKeyUsageServerAuth,
		"register",
	)
	dir := t.TempDir()
	writeCertificateFiles(t, dir, first)
	cert, err := cds.WatchCertificateFilesEvery(
		t.Context(), "public", filepath.Join(dir, "tls.crt"), filepath.Join(dir, "tls.key"), testInterval,
	)
	require.NoError(t, err)
	srv := certificateServer(t, cert)

	_, renewedKey := certificatePEM(t, renewed)
	writeFile(t, filepath.Join(dir, "tls.key"), renewedKey)

	require.Eventually(t, func() bool { return len(logs.entries(t, "WARN")) > 0 }, 5*time.Second, testInterval)
	time.Sleep(10 * testInterval)

	warnings := logs.entries(t, "WARN")
	require.Len(t, warnings, 1, "one warning for each change of the files")
	assert.Equal(t, "public", warnings[0]["app.tls.certificate.name"])
	assert.Contains(t, warnings[0]["error"], "private key does not match public key")
	requireServes(t, srv, ca, first)

	renewedCert, _ := certificatePEM(t, renewed)
	writeFile(t, filepath.Join(dir, "tls.crt"), renewedCert)

	assertEventuallyServes(t, srv, ca, renewed)
}

// Verifies: CDSL-0093
//
//nolint:paralleltest // the test changes the default logger
func TestTheLogGivesTheSerialNumberAndExpiryOfEachNewCertificate(t *testing.T) {
	logs := captureLogs(t)
	ca := newTestCA(t)
	first, replacement := ca.issue(
		t,
		x509.ExtKeyUsageServerAuth,
		"register",
	), ca.issue(
		t,
		x509.ExtKeyUsageServerAuth,
		"register",
	)
	cert := newServerCertificate(t, first)

	require.NoError(t, cert.Replace(t.Context(), replacement))

	infos := logs.entries(t, "INFO")
	require.Len(t, infos, 2)

	for i, want := range []tls.Certificate{first, replacement} {
		leaf, err := x509.ParseCertificate(want.Certificate[0])
		require.NoError(t, err)
		assert.Equal(t, "public", infos[i]["app.tls.certificate.name"])
		assert.Equal(t, strings.ToUpper(leaf.SerialNumber.Text(16)), infos[i]["app.tls.server.serial_number"])
		assert.Equal(t, leaf.NotAfter.UTC().Format(time.RFC3339), infos[i]["tls.server.not_after"])
	}
}

func TestTheWatchFailsWhenTheFilesAtStartAreNotValid(t *testing.T) {
	ca := newTestCA(t)
	dir := t.TempDir()
	writeCertificateFiles(t, dir, ca.issue(t, x509.ExtKeyUsageServerAuth, "register"))
	_, otherKey := certificatePEM(t, ca.issue(t, x509.ExtKeyUsageServerAuth, "register"))
	writeFile(t, filepath.Join(dir, "tls.key"), otherKey)

	tests := map[string]string{"a key that does not agree": "tls.crt", "a missing file": "missing.crt"}
	for name, certFile := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := cds.WatchCertificateFiles(
				t.Context(),
				"public",
				filepath.Join(dir, certFile),
				filepath.Join(dir, "tls.key"),
			)

			assert.Error(t, err)
		})
	}
}

func TestAReplacementWithNoCertificateIsRefused(t *testing.T) {
	ca := newTestCA(t)
	cert := newServerCertificate(t, ca.issue(t, x509.ExtKeyUsageServerAuth, "register"))

	assert.Error(t, cert.Replace(t.Context(), tls.Certificate{}))
}
