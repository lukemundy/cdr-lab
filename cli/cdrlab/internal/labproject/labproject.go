// Package labproject writes the compose project of the binary into a lab directory, with a record of the version of
// the binary, and removes it again.
package labproject

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/pemfile"
)

// RecordFile is the version record, relative to the lab directory.
const RecordFile = ".lab/cdrlab-version"

// Write writes each file of project into dir. It changes nothing when dir has a file of project and no record, or a
// record of a different version.
func Write(dir string, project fs.FS, version string) error {
	names, err := files(project)
	if err != nil {
		return err
	}

	recorded, found, err := readRecord(dir)
	if err != nil {
		return err
	}

	switch {
	case !found:
		if name, err := firstExisting(dir, names); err != nil || name != "" {
			return errors.Join(err, notOwnError(dir, name))
		}
	case recorded != version:
		return fmt.Errorf("the lab in %s is of cdrlab %s, and this cdrlab is %s: run cdrlab reset, then cdrlab up",
			dir, recorded, version)
	}

	// The record comes first, so that a write that stops part way leaves a lab that reset can remove.
	if err := pemfile.Write(filepath.Join(dir, RecordFile), []byte(version+"\n"), pemfile.AnyUser); err != nil {
		return err
	}

	for _, name := range names {
		data, err := fs.ReadFile(project, name)
		if err != nil {
			return err
		}

		// The containers read the files as a different user.
		if err := pemfile.Write(filepath.Join(dir, name), data, pemfile.AnyUser); err != nil {
			return err
		}
	}

	return nil
}

func notOwnError(dir, name string) error {
	if name == "" {
		return nil
	}

	return fmt.Errorf("%s has %s and no version record of cdrlab, so cdrlab did not write it: "+
		"run cdrlab up in an empty directory, or use -source in a checkout of the repository", dir, name)
}

// Recorded reports whether dir has a version record, that is, a lab of the release mode.
func Recorded(dir string) (bool, error) {
	_, found, err := readRecord(dir)
	return found, err
}

// Remove deletes each file of project, the record, and each directory that becomes empty. It deletes no other file.
func Remove(dir string, project fs.FS) error {
	names, err := files(project)
	if err != nil {
		return err
	}

	names = append(names, RecordFile)

	var dirs []string

	for _, name := range names {
		if err := os.Remove(filepath.Join(dir, name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		for d := filepath.Dir(name); d != "."; d = filepath.Dir(d) {
			dirs = append(dirs, d)
		}
	}

	// The deepest directories first, so that a parent is empty when its turn comes.
	slices.SortFunc(dirs, func(a, b string) int {
		return cmp.Or(strings.Count(b, "/")-strings.Count(a, "/"), strings.Compare(a, b))
	})

	for _, d := range slices.Compact(dirs) {
		err := os.Remove(filepath.Join(dir, d))
		if err != nil && !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, syscall.ENOTEMPTY) &&
			!errors.Is(err, syscall.EEXIST) {
			return err
		}
	}

	return nil
}

func files(project fs.FS) ([]string, error) {
	var names []string

	err := fs.WalkDir(project, ".", func(path string, e fs.DirEntry, err error) error {
		if err == nil && !e.IsDir() {
			names = append(names, path)
		}

		return err
	})

	return names, err
}

func readRecord(dir string) (version string, found bool, err error) {
	data, found, err := pemfile.Read(filepath.Join(dir, RecordFile))

	return strings.TrimSpace(string(data)), found, err
}

func firstExisting(dir string, names []string) (string, error) {
	for _, name := range names {
		_, err := os.Lstat(filepath.Join(dir, name))
		if err == nil {
			return name, nil
		}

		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
	}

	return "", nil
}
