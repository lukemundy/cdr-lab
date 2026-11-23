package cainit

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/pemfile"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

type dirFile struct {
	name string
	data []byte
	perm os.FileMode
}

// dirFiles has no root key of the CDR CA, because no component needs it after the set-up.
var dirFiles = []string{
	"cdr/root.crt",
	"cdr/intermediate.crt",
	"cdr/intermediate.key",
	"cdr/ca-bundle.pem",
	"public/root.crt",
	"public/root.key",
}

// RunDir makes the mock CDR CA and the mock public CA in dir, only when none of their files exists. It gives
// created false when all of them exist.
func RunDir(dir string, keyBits int, now func() time.Time) (created bool, err error) {
	var missing []string

	for _, name := range dirFiles {
		_, err := os.Stat(filepath.Join(dir, name))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			missing = append(missing, name)
		case err != nil:
			return false, err
		}
	}

	switch len(missing) {
	case 0:
		return false, nil
	case len(dirFiles):
	default:
		return false, fmt.Errorf("%s is absent, but other files of the lab CAs exist: "+
			"delete the directory %s to make new CAs", filepath.Join(dir, missing[0]), dir)
	}

	files, err := newDirCAs(keyBits, now)
	if err != nil {
		return false, err
	}

	return true, write(dir, files)
}

func newDirCAs(keyBits int, now func() time.Time) ([]dirFile, error) {
	var keys [3]*rsa.PrivateKey

	for i := range keys {
		key, err := rsa.GenerateKey(rand.Reader, keyBits)
		if err != nil {
			return nil, err
		}

		keys[i] = key
	}

	// X.509 times have no fractional seconds.
	lab, err := ca.NewLab(
		ca.LabKeys{CDRRoot: keys[0], CDRIntermediate: keys[1], PublicRoot: keys[2]},
		now().UTC().Truncate(time.Second),
	)
	if err != nil {
		return nil, err
	}

	intermediateKey, err := ca.EncodeKey(lab.CDRIntermediate.Key)
	if err != nil {
		return nil, err
	}

	publicKey, err := ca.EncodeKey(lab.PublicRoot.Key)
	if err != nil {
		return nil, err
	}

	cdrRoot, intermediate := lab.CDRRoot.Certificate, lab.CDRIntermediate.Certificate

	return []dirFile{
		{"cdr/root.crt", ca.EncodeCertificates(cdrRoot), pemfile.AnyUser},
		{"cdr/intermediate.crt", ca.EncodeCertificates(intermediate), pemfile.AnyUser},
		// The Register container reads the intermediate key as a different user (pki/DESIGN.md#pki-0074).
		{"cdr/intermediate.key", intermediateKey, pemfile.AnyUser},
		{"cdr/ca-bundle.pem", ca.EncodeCertificates(intermediate, cdrRoot), pemfile.AnyUser},
		{"public/root.crt", ca.EncodeCertificates(lab.PublicRoot.Certificate), pemfile.AnyUser},
		{"public/root.key", publicKey, pemfile.OwnerOnly},
	}, nil
}

// write removes each file that it wrote when a write fails, because a part of the files is a broken state.
func write(dir string, files []dirFile) error {
	for i, f := range files {
		err := pemfile.Write(filepath.Join(dir, f.name), f.data, f.perm)
		if err == nil {
			continue
		}

		for _, written := range files[:i] {
			err = errors.Join(err, pemfile.Remove(filepath.Join(dir, written.name)))
		}

		return err
	}

	return nil
}
