package pemfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/pemfile"
)

func TestWriteMakesTheDirectoryAndTheFileWithTheMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wattle-bank", "secure.key")

	require.NoError(t, pemfile.Write(path, []byte("key"), pemfile.OwnerOnly))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "key", string(data))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	dir, err := os.Stat(filepath.Dir(path))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), dir.Mode().Perm())
}

func TestWriteGivesTheModeAlsoWithAUmask(t *testing.T) {
	path := filepath.Join(t.TempDir(), "public.key")

	require.NoError(t, pemfile.Write(path, []byte("key"), pemfile.AnyUser))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

// Verifies: CLI-0126
func TestWriteReplacesTheFileAsOneUnitAndLeavesNoOtherFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secure.crt")
	require.NoError(t, pemfile.Write(path, []byte("earlier"), pemfile.AnyUser))

	before, err := os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, pemfile.Write(path, []byte("new"), pemfile.AnyUser))

	after, err := os.Stat(path)
	require.NoError(t, err)
	assert.False(t, os.SameFile(before, after), "a rename replaces the file, so a reader never gets a part of a file")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "new", string(data))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestAFailedWriteLeavesTheEarlierFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secure.crt")
	require.NoError(t, pemfile.Write(path, []byte("earlier"), pemfile.AnyUser))
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	require.Error(t, pemfile.Write(path, []byte("new"), pemfile.AnyUser))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "earlier", string(data))
}

func TestReadGivesFoundFalseForAFileThatIsAbsent(t *testing.T) {
	_, found, err := pemfile.Read(filepath.Join(t.TempDir(), "absent.crt"))

	require.NoError(t, err)
	assert.False(t, found)
}

func TestReadGivesTheData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.crt")
	require.NoError(t, pemfile.Write(path, []byte("cert"), pemfile.AnyUser))

	data, found, err := pemfile.Read(path)

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "cert", string(data))
}
