package root

import (
	"bytes"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
)

// renewableTLSFiles are the certificate file and the key file of one listener.
type renewableTLSFiles struct{ cert, key string }

func newRenewableTLSFiles(t *testing.T, certSetting, keySetting string) renewableTLSFiles {
	t.Helper()

	dir := t.TempDir()
	files := renewableTLSFiles{cert: filepath.Join(dir, "tls.crt"), key: filepath.Join(dir, "tls.key")}
	files.renew(t)
	t.Setenv(certSetting, files.cert)
	t.Setenv(keySetting, files.key)

	return files
}

// renew writes a new certificate and key, and gives the DER of the certificate.
func (f renewableTLSFiles) renew(t *testing.T) []byte {
	t.Helper()

	certFile, keyFile := tlsFiles(t)
	certPEM, err := os.ReadFile(certFile) //nolint:gosec // G304: the test makes the file
	require.NoError(t, err)
	keyPEM, err := os.ReadFile(keyFile) //nolint:gosec // G304: the test makes the file
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(f.cert, certPEM, 0o600)) //nolint:gosec // G703: the file name is in t.TempDir
	require.NoError(t, os.WriteFile(f.key, keyPEM, 0o600))   //nolint:gosec // G703: the file name is in t.TempDir

	block, _ := pem.Decode(certPEM)

	return block.Bytes
}

func startAPIs(t *testing.T) process {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	cfg, err := loadConfig(apiMode)
	require.NoError(t, err)

	proc, err := newProcess(t.Context(), pool, cfg, apiMode)
	require.NoError(t, err)

	return proc
}

// Verifies: REG-1527
func TestEachTLSListenerUsesARenewedCertificateWithNoRestart(t *testing.T) {
	setLabSettings(t)
	public := newRenewableTLSFiles(t, "PUBLIC_TLS_CERT_FILE", "PUBLIC_TLS_KEY_FILE")
	proc := startAPIs(t)

	renewed := public.renew(t)

	assert.Eventually(t, func() bool {
		_, tlsConfig := listener(t, proc, apiPort)
		cert, err := tlsConfig.GetCertificate(nil)

		return err == nil && bytes.Equal(cert.Certificate[0], renewed)
	}, 15*time.Second, 100*time.Millisecond)
}

// Verifies: REG-1515, REG-1518, REG-1521, REG-1528
func TestTheMTLSListenerServesACertificateThatTheRegisterIssuesWithNoFile(t *testing.T) {
	pool := databasetest.New(t)
	proc := startRegister(t, pool)

	_, tlsConfig := listener(t, proc, securePort)
	cert, err := tlsConfig.GetCertificate(nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"Mock CDR Register"}, cert.Leaf.Subject.Organization)
	assert.Equal(t, "register", cert.Leaf.Subject.CommonName)
	assert.Equal(t, []string{"register", "mtls.register.example"}, cert.Leaf.DNSNames)
	assert.NoError(t, cert.Leaf.CheckSignatureFrom(intermediateAuthority(t).Certificate))

	var events int
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT count(*) FROM events").Scan(&events))
	assert.Zero(t, events)
}

// Verifies: REG-1519
func TestTheAPIsDoNotStartWithNoDNSNamesOfTheMTLSListener(t *testing.T) {
	for name, value := range map[string]string{"absent": "", "an empty name": "register,"} {
		t.Run(name, func(t *testing.T) {
			setLabSettings(t)
			t.Setenv("SECURE_TLS_DNS_NAMES", value)

			_, err := loadConfig(apiMode)

			assert.ErrorContains(t, err, "SECURE_TLS_DNS_NAMES")
		})
	}
}

// Verifies: REG-1522
func TestTheAPIsDoNotStartWhenTheIssueOfTheMTLSCertificateFails(t *testing.T) {
	setLabSettings(t)
	t.Setenv("SECURE_TLS_DNS_NAMES", "not_a_host_name")
	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	cfg, err := loadConfig(apiMode)
	require.NoError(t, err)

	_, err = newProcess(t.Context(), pool, cfg, apiMode)

	assert.ErrorContains(t, err, "issue the certificate of the mTLS listener")
}
