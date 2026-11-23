package cainit_test

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/cainit"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

func runDir(t *testing.T, dir string) (bool, error) {
	t.Helper()

	return cainit.RunDir(dir, 2048, func() time.Time { return now })
}

func mode(t *testing.T, path string) os.FileMode {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)

	return info.Mode().Perm()
}

// Verifies: CLI-0011, CLI-0016, CLI-0119, CLI-0120
func TestTheCASetUpWritesTheFilesOfTheCDRCAAndThePublicCAWithNoCDRRootKey(t *testing.T) {
	dir := t.TempDir()

	created, err := runDir(t, dir)
	require.NoError(t, err)
	assert.True(t, created)

	for name, want := range map[string]os.FileMode{
		"cdr/root.crt":         0o644,
		"cdr/intermediate.crt": 0o644,
		"cdr/intermediate.key": 0o644,
		"cdr/ca-bundle.pem":    0o644,
		"public/root.crt":      0o644,
		"public/root.key":      0o600,
	} {
		assert.Equal(t, want, mode(t, filepath.Join(dir, name)), name)
	}

	var all []string
	require.NoError(t, filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(dir, path)
			all = append(all, rel)
		}

		return err
	}))
	assert.NotContains(t, all, "cdr/root.key")
	assert.Len(t, all, 6)

	root := readCert(t, dir, "cdr/root.crt")
	intermediate := readCert(t, dir, "cdr/intermediate.crt")
	bundle, err := ca.ReadCertificates(filepath.Join(dir, "cdr/ca-bundle.pem"))
	require.NoError(t, err)
	require.Len(t, bundle, 2)
	assert.Equal(t, intermediate.Raw, bundle[0].Raw)
	assert.Equal(t, root.Raw, bundle[1].Raw)
	require.NoError(t, intermediate.CheckSignatureFrom(root))

	key, err := ca.ReadKey(filepath.Join(dir, "cdr/intermediate.key"))
	require.NoError(t, err)
	assert.True(t, sameKey(intermediate.PublicKey, key.Public()))

	public := readCert(t, dir, "public/root.crt")
	assert.Equal(t, "CN=Mock Public Root CA,O=cdr-lab,C=AU", public.Subject.String())
	assert.NotEqual(t, root.Raw, public.Raw)
	publicKey, err := ca.ReadKey(filepath.Join(dir, "public/root.key"))
	require.NoError(t, err)
	assert.True(t, sameKey(public.PublicKey, publicKey.Public()))
}

func readCert(t *testing.T, dir, name string) *x509.Certificate {
	t.Helper()

	certs, err := ca.ReadCertificates(filepath.Join(dir, name))
	require.NoError(t, err)

	return certs[0]
}

// Verifies: CLI-0121, CLI-0062
func TestTheCASetUpChangesNoFilesThatExist(t *testing.T) {
	dir := t.TempDir()
	_, err := runDir(t, dir)
	require.NoError(t, err)

	before := readCert(t, dir, "cdr/root.crt")

	created, err := runDir(t, dir)

	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, before.Raw, readCert(t, dir, "cdr/root.crt").Raw)
}

// Verifies: CLI-0122
func TestTheCASetUpFailsAndChangesNothingWhenOnlySomeFilesExist(t *testing.T) {
	dir := t.TempDir()
	_, err := runDir(t, dir)
	require.NoError(t, err)
	require.NoError(t, os.Remove(filepath.Join(dir, "public/root.key")))

	before := readCert(t, dir, "cdr/root.crt")

	_, err = runDir(t, dir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), filepath.Join(dir, "public/root.key"))
	assert.Equal(t, before.Raw, readCert(t, dir, "cdr/root.crt").Raw)
	assert.NoFileExists(t, filepath.Join(dir, "public/root.key"))
}

// Verifies: CLI-0123
func TestAFailedWriteRemovesEachFileOfTheRun(t *testing.T) {
	dir := t.TempDir()
	public := filepath.Join(dir, "public")
	require.NoError(t, os.Mkdir(public, 0o755))
	require.NoError(t, os.Chmod(public, 0o555))
	t.Cleanup(func() { _ = os.Chmod(public, 0o755) })

	_, err := runDir(t, dir)

	require.Error(t, err)

	for _, name := range []string{"cdr/root.crt", "cdr/intermediate.crt", "cdr/intermediate.key", "cdr/ca-bundle.pem"} {
		assert.NoFileExists(t, filepath.Join(dir, name))
	}
}
