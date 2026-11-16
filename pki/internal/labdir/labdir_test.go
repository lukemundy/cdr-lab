package labdir_test

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pki/internal/labdir"
)

func newLab(t *testing.T) ca.Lab {
	t.Helper()

	var keys []*rsa.PrivateKey

	for range 3 {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		keys = append(keys, k)
	}

	lab, err := ca.NewLab(ca.LabKeys{CDRRoot: keys[0], CDRIntermediate: keys[1], PublicRoot: keys[2]}, time.Now())
	require.NoError(t, err)

	return lab
}

func TestASavedLabLoadsAgain(t *testing.T) {
	dir := t.TempDir()
	lab := newLab(t)

	require.NoError(t, labdir.Save(dir, lab))
	got, err := labdir.Load(dir)

	require.NoError(t, err)
	assert.Equal(t, lab.CDRRoot.Certificate.Raw, got.CDRRoot.Certificate.Raw)
	assert.Equal(t, lab.CDRIntermediate.Certificate.Raw, got.CDRIntermediate.Certificate.Raw)
	assert.Equal(t, lab.PublicRoot.Certificate.Raw, got.PublicRoot.Certificate.Raw)
	assert.True(t, lab.CDRIntermediate.Key.Public().(*rsa.PublicKey).Equal(got.CDRIntermediate.Key.Public()))
	assert.True(t, lab.PublicRoot.Key.Public().(*rsa.PublicKey).Equal(got.PublicRoot.Key.Public()))
}

// Verifies: PKI-0073, PKI-0074
func TestOnlyTheOwnerCanReadTheRootKeys(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, labdir.Save(dir, newLab(t)))

	for _, name := range []string{"cdr/root.key", "public/root.key"} {
		info, err := os.Stat(filepath.Join(dir, name))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), name)
	}

	readableByAll := []string{
		"cdr/intermediate.key", "cdr/root.crt", "cdr/intermediate.crt", "cdr/ca-bundle.pem", "public/root.crt",
	}
	for _, name := range readableByAll {
		info, err := os.Stat(filepath.Join(dir, name))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o644), info.Mode().Perm(), name)
	}
}

// Verifies: PKI-0013
func TestTheCDRBundleHasTheIntermediateAndTheRoot(t *testing.T) {
	dir := t.TempDir()
	lab := newLab(t)
	require.NoError(t, labdir.Save(dir, lab))

	bundle, err := ca.ReadCertificates(filepath.Join(dir, "cdr/ca-bundle.pem"))

	require.NoError(t, err)
	require.Len(t, bundle, 2)
	assert.Equal(t, lab.CDRIntermediate.Certificate.Raw, bundle[0].Raw)
	assert.Equal(t, lab.CDRRoot.Certificate.Raw, bundle[1].Raw)
}

func TestTheStateOfALabDirectory(t *testing.T) {
	dir := t.TempDir()

	state, err := labdir.State(dir)
	require.NoError(t, err)
	assert.Equal(t, labdir.Empty, state)

	require.NoError(t, labdir.Save(dir, newLab(t)))
	state, err = labdir.State(dir)
	require.NoError(t, err)
	assert.Equal(t, labdir.Complete, state)

	require.NoError(t, os.Remove(filepath.Join(dir, "public/root.key")))
	_, err = labdir.State(dir)
	assert.EqualError(t, err, filepath.Join(dir, "public/root.key")+" is missing, but other lab CA files exist")
}

func TestADirectoryThatDoesNotExistIsEmpty(t *testing.T) {
	state, err := labdir.State(filepath.Join(t.TempDir(), "pki"))

	require.NoError(t, err)
	assert.Equal(t, labdir.Empty, state)
}

func TestSaveRefusesToReplaceALab(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, labdir.Save(dir, newLab(t)))

	assert.ErrorIs(t, labdir.Save(dir, newLab(t)), os.ErrExist)
}
