// Package pemfile reads and writes the certificate and key files of the file output of the commands.
package pemfile

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	OwnerOnly os.FileMode = 0o600
	// AnyUser is for the lab files that containers read as a different user (pki/DESIGN.md#pki-0095).
	AnyUser os.FileMode = 0o644
	// dirMode lets the containers of the lab read the files of the directory.
	dirMode os.FileMode = 0o755
)

// Write replaces the file with a rename, so that a reader gets the earlier file or the new file, never a part of a
// file. It makes the directory of the file when it is absent.
func Write(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return err
	}

	dir, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer func() { _ = dir.Close() }()

	suffix := make([]byte, 8)
	_, _ = rand.Read(suffix)
	temp := "." + filepath.Base(path) + ".tmp-" + hex.EncodeToString(suffix)

	f, err := dir.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	// The umask can remove permissions from the mode of OpenFile.
	err = errors.Join(err, f.Chmod(perm), f.Sync(), f.Close())
	if err == nil {
		err = dir.Rename(temp, filepath.Base(path))
	}

	if err != nil {
		return errors.Join(err, ignoreAbsent(dir.Remove(temp)))
	}

	return nil
}

// Read gives found false when the file is absent.
func Read(path string) (data []byte, found bool, err error) {
	data, err = fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}

	return data, err == nil, err
}

func Remove(path string) error {
	return ignoreAbsent(os.Remove(path))
}

func ignoreAbsent(err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return err
}
