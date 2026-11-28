package root

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testBrandDir is a valid brand directory with the brand name "Fernleaf".
const testBrandDir = "../adapter/branddir/testdata/valid"

const webAppHost = "localhost:8585"

// certificateFiles writes one self-signed certificate and its key.
func certificateFiles(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "sorrel.example"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	dir := t.TempDir()
	certFile, keyFile = filepath.Join(dir, "tls.crt"), filepath.Join(dir, "tls.key")
	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0o600))

	return certFile, keyFile
}

func serveSettings(t *testing.T) map[string]string {
	t.Helper()

	certFile, keyFile := certificateFiles(t)

	return map[string]string{
		"DATA_RECIPIENT_BRAND_DIR": testBrandDir,
		"WEB_APP_ALLOWED_HOSTS":    webAppHost,
		"WEB_APP_TLS_CERT_FILE":    certFile,
		"WEB_APP_TLS_KEY_FILE":     keyFile,
	}
}

// setServe sets each setting of serve, and clears the other brand setting.
func setServe(t *testing.T) {
	t.Helper()

	t.Setenv("DATA_RECIPIENT_BRAND", "")

	for name, value := range serveSettings(t) {
		t.Setenv(name, value)
	}
}

func TestTheConfigurationLoadsWhenEachSettingIsPresent(t *testing.T) {
	setServe(t)

	cfg, err := loadConfig()

	require.NoError(t, err)
	assert.Equal(t, "Fernleaf", cfg.brand.Name)
	assert.Equal(t, []string{webAppHost}, cfg.webAppHosts)
	assert.NotEmpty(t, cfg.webAppCert.cert)
	assert.NotEmpty(t, cfg.webAppCert.key)
}

// Verifies: DR-0011, DR-0206
func TestServeDoesNotStartWithoutASettingAndTheErrorNamesIt(t *testing.T) {
	// The error of a missing brand names the two brand settings.
	tests := map[string]string{
		"DATA_RECIPIENT_BRAND_DIR": "DATA_RECIPIENT_BRAND",
		"WEB_APP_ALLOWED_HOSTS":    "WEB_APP_ALLOWED_HOSTS",
		"WEB_APP_TLS_CERT_FILE":    "WEB_APP_TLS_CERT_FILE",
		"WEB_APP_TLS_KEY_FILE":     "WEB_APP_TLS_KEY_FILE",
	}
	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			setServe(t)
			t.Setenv(name, "")

			_, err := loadConfig()

			assert.ErrorContains(t, err, want)
		})
	}
}

// Verifies: DR-0107, DR-0110
func TestServeNeedsExactlyOneBrandSetting(t *testing.T) {
	setServe(t)
	t.Setenv("DATA_RECIPIENT_BRAND", "sorrel")

	_, err := loadConfig()

	assert.ErrorContains(t, err, "set only one of DATA_RECIPIENT_BRAND and DATA_RECIPIENT_BRAND_DIR")
}

// Verifies: DR-0111, DR-0121
func TestThePresetBrandSorrelLoadsWithNoFileOutsideTheDataRecipient(t *testing.T) {
	setServe(t)
	t.Setenv("DATA_RECIPIENT_BRAND_DIR", "")
	t.Setenv("DATA_RECIPIENT_BRAND", "sorrel")

	cfg, err := loadConfig()

	require.NoError(t, err)
	assert.Equal(t, "Sorrel", cfg.brand.Name)
}

func TestAnUnknownPresetBrandGivesTheNamesOfThePresets(t *testing.T) {
	setServe(t)
	t.Setenv("DATA_RECIPIENT_BRAND_DIR", "")
	t.Setenv("DATA_RECIPIENT_BRAND", "fernleaf")

	_, err := loadConfig()

	assert.ErrorContains(t, err, "sorrel")
}

// Verifies: DR-0120
func TestABrandDirectoryThatIsNotValidStopsTheStartAndTheErrorNamesTheFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, os.DirFS(testBrandDir)))
	require.NoError(t, os.Remove(filepath.Join(dir, "mark.svg")))

	setServe(t)
	t.Setenv("DATA_RECIPIENT_BRAND_DIR", dir)

	_, err := loadConfig()

	require.ErrorContains(t, err, "DATA_RECIPIENT_BRAND_DIR")
	assert.ErrorContains(t, err, "mark.svg")
}
