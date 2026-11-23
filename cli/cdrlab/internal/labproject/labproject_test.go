package labproject

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var project = fstest.MapFS{
	"compose.yaml":             {Data: []byte("services: {}\n")},
	"lab/certificates.json":    {Data: []byte("{}\n")},
	"lab/postgres/init.sql":    {Data: []byte("select 1;\n")},
	"services/register/a.json": {Data: []byte("[]\n")},
}

func readFile(t *testing.T, dir, name string) string {
	t.Helper()

	data, err := fs.ReadFile(os.DirFS(dir), name)
	require.NoError(t, err)

	return string(data)
}

func fileMode(t *testing.T, path string) os.FileMode {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)

	return info.Mode().Perm()
}

// Verifies: CLI-0140
func TestWriteWritesEachFileAndTheVersionRecord(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Write(dir, project, "v0.3.1"))

	for name, f := range project {
		assert.Equal(t, string(f.Data), readFile(t, dir, name))
		// The containers read the files as a different user.
		assert.Equal(t, os.FileMode(0o644), fileMode(t, filepath.Join(dir, name)))
	}

	assert.Equal(t, "v0.3.1\n", readFile(t, dir, RecordFile))

	recorded, err := Recorded(dir)
	require.NoError(t, err)
	assert.True(t, recorded)
}

// Verifies: CLI-0140
func TestWriteAgainWithTheSameVersionWritesTheProjectAgain(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Write(dir, project, "v0.3.1"))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte("changed\n"), 0o600))

	require.NoError(t, Write(dir, project, "v0.3.1"))

	assert.Equal(t, "services: {}\n", readFile(t, dir, "compose.yaml"))
}

// Verifies: CLI-0144
func TestWriteChangesNothingWhenAFileOfTheProjectHasNoRecord(t *testing.T) {
	for _, name := range []string{"compose.yaml", "lab/postgres/init.sql"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o750))
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("mine\n"), 0o600))

			err := Write(dir, project, "v0.3.1")

			require.Error(t, err)
			assert.Contains(t, err.Error(), name)
			assert.Contains(t, err.Error(), "-source")
			assert.NoFileExists(t, filepath.Join(dir, RecordFile))

			assert.Equal(t, "mine\n", readFile(t, dir, name))
		})
	}
}

// Verifies: CLI-0145
func TestWriteChangesNothingWhenTheRecordHasADifferentVersion(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Write(dir, project, "v0.3.1"))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte("old\n"), 0o600))

	err := Write(dir, project, "v0.4.0")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "v0.3.1")
	assert.Contains(t, err.Error(), "v0.4.0")
	assert.Contains(t, err.Error(), "cdrlab reset")

	assert.Equal(t, "old\n", readFile(t, dir, "compose.yaml"))
}

// Verifies: CLI-0156
func TestRemoveDeletesOnlyTheProjectTheRecordAndTheDirectoriesThatBecomeEmpty(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Write(dir, project, "v0.3.1"))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "lab/notes.txt"), []byte("mine\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.txt"), []byte("mine\n"), 0o600))

	require.NoError(t, Remove(dir, project))

	for name := range project {
		assert.NoFileExists(t, filepath.Join(dir, name))
	}

	assert.NoFileExists(t, filepath.Join(dir, RecordFile))
	assert.NoDirExists(t, filepath.Join(dir, ".lab"))
	assert.NoDirExists(t, filepath.Join(dir, "services"))
	assert.NoDirExists(t, filepath.Join(dir, "lab/postgres"))
	assert.FileExists(t, filepath.Join(dir, "lab/notes.txt"))
	assert.FileExists(t, filepath.Join(dir, "other.txt"))

	recorded, err := Recorded(dir)
	require.NoError(t, err)
	assert.False(t, recorded)
}

func TestRemoveOfADirectoryWithPartOfTheProjectDeletesTheRest(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Write(dir, project, "v0.3.1"))
	require.NoError(t, os.Remove(filepath.Join(dir, "compose.yaml")))

	require.NoError(t, Remove(dir, project))

	assert.NoFileExists(t, filepath.Join(dir, "lab/certificates.json"))
}
