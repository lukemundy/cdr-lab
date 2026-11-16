package root

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"maps"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/cds/revocation"
	"github.com/lukemundy/cdr-lab/pkg/platform/consolehttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/operatorhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/brands"
)

// certificateFiles writes one self-signed certificate, which is also a CA bundle, and its key.
func certificateFiles(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "bank.example"},
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

const operatorHost = "bank.test:8082"

// testBrandDir is a valid brand directory with the brand name "Coastal Bank".
const testBrandDir = "../adapter/branddir/testdata/valid"

var (
	serveMode  = mode{api: true, workers: true}
	apiMode    = mode{api: true}
	workerMode = mode{workers: true}
)

func apiSettings(t *testing.T) map[string]string {
	t.Helper()

	certFile, keyFile := certificateFiles(t)

	return map[string]string{
		"DATA_HOLDER_BRAND_DIR":  testBrandDir,
		"PUBLIC_BASE_URL":        "https://bank.example",
		"PUBLIC_TLS_CERT_FILE":   certFile,
		"PUBLIC_TLS_KEY_FILE":    keyFile,
		"SECURE_BASE_URL":        "https://mtls.bank.example",
		"REGISTER_JWKS_URL":      "https://register.example/jwks",
		"OPERATOR_TOKEN":         "test-operator-token",
		"OPERATOR_ALLOWED_HOSTS": operatorHost,
		"SMTP_ADDRESS":           "mail.bank.example:25",
		"MAIL_FROM":              "no-reply@bank.example",
		"SECURE_TLS_CERT_FILE":   certFile,
		"SECURE_TLS_KEY_FILE":    keyFile,
		"CDR_CA_BUNDLE_FILE":     certFile,
	}
}

func workerSettings() map[string]string {
	return map[string]string{
		"DATA_HOLDER_BRAND_ID":     "d7385786-d8b1-4d4b-bb80-b21616df1a4d",
		"REGISTER_PUBLIC_BASE_URL": "https://register.example",
		"REGISTER_STATUS_INTERVAL": "4m",
	}
}

func settingsOf(t *testing.T, m mode) map[string]string {
	t.Helper()

	settings := map[string]string{}
	if m.api {
		maps.Copy(settings, apiSettings(t))
	}

	if m.workers {
		maps.Copy(settings, workerSettings())
	}

	return settings
}

// setOnly sets the settings of the mode, and clears the settings of the other modes.
func setOnly(t *testing.T, m mode) {
	t.Helper()

	for name := range settingsOf(t, serveMode) {
		t.Setenv(name, "")
	}

	for name, value := range settingsOf(t, m) {
		t.Setenv(name, value)
	}
}

func TestTheConfigurationLoadsWhenEachSettingIsPresent(t *testing.T) {
	setOnly(t, serveMode)

	cfg, err := loadConfig(serveMode)

	require.NoError(t, err)
	assert.Equal(t, "https://bank.example", cfg.api.publicBaseURL)
	assert.Equal(t, "https://mtls.bank.example", cfg.api.secureBaseURL)
	assert.Equal(t, operatorhttp.Config{
		Token: "test-operator-token", AllowedHosts: []string{operatorHost},
	}, cfg.api.operator)
	assert.Len(t, cfg.api.cdrCAs, 1)
	assert.Equal(t, "Coastal Bank", cfg.api.brand.Name)
	assert.Equal(t, consolehttp.Config{}, cfg.api.console)
	assert.Equal(t, "d7385786-d8b1-4d4b-bb80-b21616df1a4d", cfg.workers.brandID)
	assert.Equal(t, "https://register.example", cfg.workers.registerPublicBaseURL)
	assert.Equal(t, 4*time.Minute, cfg.workers.statusInterval)
}

// Verifies: DH-0010
func TestEachModeLoadsWithOnlyTheSettingsThatItUses(t *testing.T) {
	for name, m := range map[string]mode{"serve": serveMode, "api": apiMode, "worker": workerMode} {
		t.Run(name, func(t *testing.T) {
			setOnly(t, m)

			_, err := loadConfig(m)

			require.NoError(t, err)
		})
	}
}

// Verifies: DH-0011
func TestEachModeNeedsEachSettingThatItUses(t *testing.T) {
	for name, m := range map[string]mode{"serve": serveMode, "api": apiMode, "worker": workerMode} {
		for absent := range settingsOf(t, m) {
			t.Run(name+"/"+absent, func(t *testing.T) {
				setOnly(t, m)
				t.Setenv(absent, "")

				_, err := loadConfig(m)

				assert.ErrorContains(t, err, absent)
			})
		}
	}
}

func TestTheConfigurationRefusesFilesThatDoNotExist(t *testing.T) {
	tests := map[string]string{
		"PUBLIC_TLS_CERT_FILE": "load the TLS certificate of the public APIs",
		"SECURE_TLS_CERT_FILE": "load the TLS certificate of the secure APIs",
		"CDR_CA_BUNDLE_FILE":   "load the CDR CA bundle",
	}
	for setting, wantError := range tests {
		t.Run(setting, func(t *testing.T) {
			setOnly(t, apiMode)

			t.Setenv(setting, "does-not-exist.pem")

			_, err := loadConfig(apiMode)

			assert.ErrorContains(t, err, wantError)
		})
	}
}

func TestTheStatusIntervalIsADurationOfMoreThanZeroAndAtMost4Minutes(t *testing.T) {
	for _, interval := range []string{"4m1s", "0s", "-1m", "5", "soon"} {
		t.Run(interval, func(t *testing.T) {
			setOnly(t, workerMode)

			t.Setenv("REGISTER_STATUS_INTERVAL", interval)

			_, err := loadConfig(workerMode)

			assert.ErrorContains(t, err, "REGISTER_STATUS_INTERVAL")
		})
	}
}

func TestFixedOTPsAreOptional(t *testing.T) {
	setOnly(t, apiMode)

	t.Setenv("FIXED_OTPS", "")

	cfg, err := loadConfig(apiMode)

	require.NoError(t, err)
	assert.Empty(t, cfg.api.fixedOTPs)
}

func TestFixedOTPsLoadAndAnEntryThatIsNotValidStopsTheStart(t *testing.T) {
	setOnly(t, apiMode)

	t.Setenv("FIXED_OTPS", "alex@example.com:111111")

	cfg, err := loadConfig(apiMode)
	require.NoError(t, err)
	assert.Equal(t, "111111", cfg.api.fixedOTPs.NewOTP("alex@example.com", 42))

	t.Setenv("FIXED_OTPS", "alex@example.com")

	_, err = loadConfig(apiMode)
	assert.ErrorContains(t, err, "FIXED_OTPS")
}

func TestTheAPIsLoadTheConsoleSettings(t *testing.T) {
	setOnly(t, apiMode)
	t.Setenv("CONSOLE_ENABLED", "true")
	t.Setenv("CONSOLE_ALLOWED_HOSTS", "bank.test:8084")

	cfg, err := loadConfig(apiMode)

	require.NoError(t, err)
	assert.Equal(t, consolehttp.Config{Enabled: true, AllowedHosts: []string{"bank.test:8084"}}, cfg.api.console)

	t.Setenv("CONSOLE_ALLOWED_HOSTS", "")

	_, err = loadConfig(apiMode)
	assert.ErrorContains(t, err, "CONSOLE_ALLOWED_HOSTS")
}

// Verifies: DH-1487
func TestTheWorkersNeedNoConsoleSettingAndNoBrandSetting(t *testing.T) {
	setOnly(t, workerMode)
	t.Setenv("CONSOLE_ENABLED", "true")
	t.Setenv("CONSOLE_ALLOWED_HOSTS", "")
	t.Setenv("DATA_HOLDER_BRAND", "")
	t.Setenv("DATA_HOLDER_BRAND_DIR", "")

	_, err := loadConfig(workerMode)

	require.NoError(t, err)
}

// Verifies: DH-0467
func TestTheCRLMaximumAgeIsOptionalAndIsADurationOfZeroOrMore(t *testing.T) {
	for value, want := range map[string]time.Duration{"": revocation.DefaultMaxAge, "0s": 0, "30s": 30 * time.Second} {
		t.Run(value, func(t *testing.T) {
			setOnly(t, apiMode)

			t.Setenv("CRL_MAX_AGE", value)

			cfg, err := loadConfig(apiMode)

			require.NoError(t, err)
			assert.Equal(t, want, cfg.api.crlMaxAge)
		})
	}

	for _, value := range []string{"-1s", "5", "soon"} {
		t.Run(value, func(t *testing.T) {
			setOnly(t, apiMode)

			t.Setenv("CRL_MAX_AGE", value)

			_, err := loadConfig(apiMode)

			assert.ErrorContains(t, err, "CRL_MAX_AGE")
		})
	}
}

func TestTheBrandIsAPresetBrandOrABrandDirectoryAndNotBoth(t *testing.T) {
	preset := brands.Names()[0]

	tests := map[string]struct {
		preset, dir string
		want        string
	}{
		"preset":    {preset: preset},
		"directory": {dir: testBrandDir},
		"both": {
			preset: preset,
			dir:    testBrandDir,
			want:   "set only one of DATA_HOLDER_BRAND and DATA_HOLDER_BRAND_DIR",
		},
		"neither": {want: "set DATA_HOLDER_BRAND to the name of a preset brand"},
		"unknown preset": {
			preset: "nope",
			want:   `DATA_HOLDER_BRAND: "nope" is not a preset brand: give one of ` + preset,
		},
		"no directory": {
			dir:  "testdata/nothing-here",
			want: "DATA_HOLDER_BRAND_DIR testdata/nothing-here: open the brand directory",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setOnly(t, apiMode)
			t.Setenv("DATA_HOLDER_BRAND", tc.preset)
			t.Setenv("DATA_HOLDER_BRAND_DIR", tc.dir)

			cfg, err := loadConfig(apiMode)

			if tc.want == "" {
				require.NoError(t, err)
				assert.NotEmpty(t, cfg.api.brand.Name)
			} else {
				assert.ErrorContains(t, err, tc.want)
			}
		})
	}
}

func TestASymbolicLinkOutOfTheBrandDirectoryIsRefused(t *testing.T) {
	outside := t.TempDir()
	require.NoError(t, os.CopyFS(outside, os.DirFS(testBrandDir)))

	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, os.DirFS(testBrandDir)))
	require.NoError(t, os.Remove(filepath.Join(dir, "brand.json")))
	require.NoError(t, os.Symlink(filepath.Join(outside, "brand.json"), filepath.Join(dir, "brand.json")))

	setOnly(t, apiMode)
	t.Setenv("DATA_HOLDER_BRAND_DIR", dir)

	_, err := loadConfig(apiMode)

	require.ErrorContains(t, err, "brand.json: ")
}

func TestADirectoryThatBreaksTheContractStopsTheStartAndTheErrorNamesTheFileAndTheRule(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, os.DirFS(testBrandDir)))
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(dir, "brand.json"),
			[]byte(`{"brandName": "Other Bank", "bsb": "376100", "layout": "meadow"}`),
			0o600,
		),
	)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o600))

	setOnly(t, apiMode)
	t.Setenv("DATA_HOLDER_BRAND_DIR", dir)

	_, err := loadConfig(apiMode)

	require.Error(t, err)

	message := err.Error()
	assert.Contains(t, message, "DATA_HOLDER_BRAND_DIR "+dir+": the brand is not valid:\n")
	assert.Contains(t, message, "notes.txt: the brand directory contract has no file of this name")
	assert.Contains(t, message, `products.json: products are not valid:`)
	assert.Contains(t, message, `brand "Coastal Bank" is not the brand name "Other Bank"`)
}

// The brand directory is the only source of the display name of the sender.
func TestTheSenderAddressIsAnAddressOnly(t *testing.T) {
	for _, from := range []string{
		"Wattle Bank <no-reply@bank.example>",
		`"Wattle Bank" <no-reply@bank.example>`,
		"<no-reply@bank.example>",
		"no-reply@bank.example (Wattle Bank)",
		"no-reply",
		"a@bank.example, b@bank.example",
	} {
		t.Run(from, func(t *testing.T) {
			setOnly(t, apiMode)
			t.Setenv("MAIL_FROM", from)

			_, err := loadConfig(apiMode)

			require.ErrorContains(t, err, "MAIL_FROM: give only the sender address, with no display name")
		})
	}
}

func TestALogoPNGThatIsNotValidStopsTheStart(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, os.DirFS(testBrandDir)))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "logo.png"), []byte("GIF89a"), 0o600))

	setOnly(t, apiMode)
	t.Setenv("DATA_HOLDER_BRAND_DIR", dir)

	_, err := loadConfig(apiMode)

	require.ErrorContains(t, err, "logo.png: the file is not a valid PNG image")
}
