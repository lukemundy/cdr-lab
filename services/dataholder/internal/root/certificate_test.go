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

	certFile, keyFile := certificateFiles(t)
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

// Verifies: DH-2648
func TestEachTLSListenerUsesARenewedCertificateWithNoRestart(t *testing.T) {
	setOnly(t, apiMode)
	t.Setenv("BANK_SITE_ENABLED", "true")
	t.Setenv("BANK_SITE_ALLOWED_HOSTS", "bank.example")
	public := newRenewableTLSFiles(t, "PUBLIC_TLS_CERT_FILE", "PUBLIC_TLS_KEY_FILE")
	secure := newRenewableTLSFiles(t, "SECURE_TLS_CERT_FILE", "SECURE_TLS_KEY_FILE")
	proc := startAPIs(t)

	renewedPublic := public.renew(t)
	renewed := map[int][]byte{apiPort: renewedPublic, bankSitePort: renewedPublic, securePort: secure.renew(t)}

	var ports []int

	for _, l := range proc.listeners {
		if l.TLSConfig == nil {
			continue
		}

		ports = append(ports, l.Port)
		assert.Eventually(t, func() bool {
			cert, err := l.TLSConfig.GetCertificate(nil)

			return err == nil && bytes.Equal(cert.Certificate[0], renewed[l.Port])
		}, 15*time.Second, 100*time.Millisecond, "port %d", l.Port)
	}

	assert.ElementsMatch(t, []int{apiPort, securePort, bankSitePort}, ports)
}

func TestTheAPIsDoNotStartWithATLSCertificateFileThatDoesNotExist(t *testing.T) {
	setOnly(t, apiMode)
	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	for _, setting := range []string{"PUBLIC_TLS_CERT_FILE", "SECURE_TLS_CERT_FILE"} {
		t.Run(setting, func(t *testing.T) {
			t.Setenv(setting, "does-not-exist.pem")

			cfg, err := loadConfig(apiMode)
			require.NoError(t, err)

			_, err = newProcess(t.Context(), pool, cfg, apiMode)

			assert.ErrorContains(t, err, "TLS certificate")
		})
	}
}
