package publiccert_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/cainit"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certcheck"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/pemfile"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/publiccert"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

func setup(t *testing.T) publiccert.Config {
	t.Helper()

	dir := t.TempDir()
	_, err := cainit.RunDir(dir, 2048, time.Now)
	require.NoError(t, err)

	return publiccert.Config{
		Dir:         dir,
		Cert:        filepath.Join(dir, "register", "public.crt"),
		Key:         filepath.Join(dir, "register", "public.key"),
		KeyPerm:     pemfile.OwnerOnly,
		Names:       []string{"register", "localhost"},
		RenewBefore: certcheck.DefaultRenewBefore,
		Now:         time.Now,
	}
}

func issued(t *testing.T, cfg publiccert.Config) bool {
	t.Helper()

	cert, err := publiccert.Run(cfg)
	require.NoError(t, err)

	return cert != nil
}

func read(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return data
}

func mode(t *testing.T, path string) os.FileMode {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)

	return info.Mode().Perm()
}

// Verifies: CLI-0130, CLI-0131
func TestTheCommandIssuesATLSServerCertificateFromThePublicCAOfTheDirectory(t *testing.T) {
	cfg := setup(t)

	assert.True(t, issued(t, cfg))

	chain, err := ca.ReadCertificates(cfg.Cert)
	require.NoError(t, err)
	require.Len(t, chain, 1)
	assert.Equal(t, "register", chain[0].Subject.CommonName)
	assert.ElementsMatch(t, []string{"register", "localhost"}, chain[0].DNSNames)

	roots, err := ca.ReadCertificates(filepath.Join(cfg.Dir, "public/root.crt"))
	require.NoError(t, err)
	require.NoError(t, chain[0].CheckSignatureFrom(roots[0]))

	key, err := ca.ReadKey(cfg.Key)
	require.NoError(t, err)
	assert.True(t, certcheck.SameKey(chain[0].PublicKey, key.Public()))

	rsaKey, ok := key.(*rsa.PrivateKey)
	require.True(t, ok)
	assert.Equal(t, 2048, rsaKey.N.BitLen())
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, chain[0].ExtKeyUsage)
}

// Verifies: CLI-0131
func TestOnlyTheOwnerCanReadTheKeyFileUnlessAnInputPermitsAnyUser(t *testing.T) {
	cfg := setup(t)
	issued(t, cfg)
	assert.Equal(t, os.FileMode(0o600), mode(t, cfg.Key))

	cfg.Cert, cfg.Key = filepath.Join(cfg.Dir, "b.crt"), filepath.Join(cfg.Dir, "b.key")
	cfg.KeyPerm = pemfile.AnyUser
	issued(t, cfg)
	assert.Equal(t, os.FileMode(0o644), mode(t, cfg.Key))
}

// Verifies: CLI-0132, CLI-0062
func TestTheCommandChangesNothingWhenTheCertificateIsCurrent(t *testing.T) {
	cfg := setup(t)
	issued(t, cfg)
	cert, key := read(t, cfg.Cert), read(t, cfg.Key)

	assert.False(t, issued(t, cfg))
	assert.Equal(t, cert, read(t, cfg.Cert))
	assert.Equal(t, key, read(t, cfg.Key))
}

// Verifies: CLI-0132
func TestTheCommandRenewsACertificateThatIsNotCurrent(t *testing.T) {
	for name, change := range map[string]func(*testing.T, *publiccert.Config){
		"it expires within the renewal window": func(_ *testing.T, cfg *publiccert.Config) {
			cfg.Now = func() time.Time { return time.Now().AddDate(0, 12, 10) }
		},
		"a DNS name was added": func(_ *testing.T, cfg *publiccert.Config) {
			cfg.Names = append(cfg.Names, "register.lab")
		},
		"it does not chain to the public root": func(t *testing.T, cfg *publiccert.Config) {
			for _, name := range []string{"public/root.crt", "public/root.key"} {
				require.NoError(t, os.Remove(filepath.Join(cfg.Dir, name)))
			}

			for _, name := range []string{"cdr/root.crt", "cdr/intermediate.crt", "cdr/intermediate.key", "cdr/ca-bundle.pem"} {
				require.NoError(t, os.Remove(filepath.Join(cfg.Dir, name)))
			}

			_, err := cainit.RunDir(cfg.Dir, 2048, time.Now)
			require.NoError(t, err)
		},
		"the key does not agree": func(t *testing.T, cfg *publiccert.Config) {
			other, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)
			data, err := ca.EncodeKey(other)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(cfg.Key, data, 0o600))
		},
	} {
		t.Run(name, func(t *testing.T) {
			cfg := setup(t)
			issued(t, cfg)

			change(t, &cfg)

			assert.True(t, issued(t, cfg))
			assert.False(t, issued(t, cfg))
		})
	}
}

func TestTheCommandNeedsThePublicCAOfTheDirectory(t *testing.T) {
	cfg := setup(t)
	cfg.Dir = t.TempDir()

	_, err := publiccert.Run(cfg)

	require.Error(t, err)
	assert.NoFileExists(t, cfg.Cert)
}

func TestTheCommandRefusesANameThatIsNotAHostName(t *testing.T) {
	cfg := setup(t)
	cfg.Names = []string{"not a host"}

	_, err := publiccert.Run(cfg)

	require.Error(t, err)
	assert.NoFileExists(t, cfg.Key)
}
