// Package labdir keeps the lab CAs as PEM files in a directory.
package labdir

import (
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

const (
	cdrRootCert         = "cdr/root.crt"
	cdrRootKey          = "cdr/root.key"
	cdrIntermediateCert = "cdr/intermediate.crt"
	cdrIntermediateKey  = "cdr/intermediate.key"
	CDRBundle           = "cdr/ca-bundle.pem"
	PublicRootCert      = "public/root.crt"
	publicRootKey       = "public/root.key"
)

var files = []string{
	cdrRootCert,
	cdrRootKey,
	cdrIntermediateCert,
	cdrIntermediateKey,
	CDRBundle,
	PublicRootCert,
	publicRootKey,
}

type DirState int

const (
	Empty DirState = iota + 1
	Complete
)

// State gives an error when only some of the lab CA files exist.
func State(dir string) (DirState, error) {
	var missing []string

	for _, name := range files {
		_, err := os.Stat(filepath.Join(dir, name))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			missing = append(missing, name)
		case err != nil:
			return 0, err
		}
	}

	switch len(missing) {
	case 0:
		return Complete, nil
	case len(files):
		return Empty, nil
	default:
		return 0, fmt.Errorf("%s is missing, but other lab CA files exist", filepath.Join(dir, missing[0]))
	}
}

// Save does not replace a file that exists.
func Save(dir string, lab ca.Lab) error {
	for _, sub := range []string{"cdr", "public"} {
		//nolint:gosec // lab containers read the CA certificates and the intermediate key as another user
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return err
		}
	}

	return errors.Join(
		WriteCertificates(filepath.Join(dir, cdrRootCert), lab.CDRRoot.Certificate),
		WriteKey(filepath.Join(dir, cdrRootKey), lab.CDRRoot.Key, 0o600),
		WriteCertificates(filepath.Join(dir, cdrIntermediateCert), lab.CDRIntermediate.Certificate),
		// The Register container reads the intermediate key as another user (pki/DESIGN.md: Keys).
		WriteKey(filepath.Join(dir, cdrIntermediateKey), lab.CDRIntermediate.Key, 0o644),
		WriteCertificates(filepath.Join(dir, CDRBundle), lab.CDRIntermediate.Certificate, lab.CDRRoot.Certificate),
		WriteCertificates(filepath.Join(dir, PublicRootCert), lab.PublicRoot.Certificate),
		WriteKey(filepath.Join(dir, publicRootKey), lab.PublicRoot.Key, 0o600),
	)
}

func Load(dir string) (ca.Lab, error) {
	cdrRoot, rootErr := loadAuthority(dir, cdrRootCert, cdrRootKey)
	cdrIntermediate, intermediateErr := loadAuthority(dir, cdrIntermediateCert, cdrIntermediateKey)
	publicRoot, publicErr := LoadPublicRoot(dir)
	lab := ca.Lab{CDRRoot: cdrRoot, CDRIntermediate: cdrIntermediate, PublicRoot: publicRoot}

	return lab, errors.Join(rootErr, intermediateErr, publicErr)
}

func LoadPublicRoot(dir string) (ca.Authority, error) {
	return loadAuthority(dir, PublicRootCert, publicRootKey)
}

func loadAuthority(dir, certName, keyName string) (ca.Authority, error) {
	certs, certErr := ca.ReadCertificates(filepath.Join(dir, certName))
	key, keyErr := ca.ReadKey(filepath.Join(dir, keyName))
	authority := ca.Authority{Key: key}

	if len(certs) > 0 {
		authority.Certificate = certs[0]
	}

	return authority, errors.Join(certErr, keyErr)
}

func WriteCertificates(path string, certs ...*x509.Certificate) error {
	return writeNew(path, ca.EncodeCertificates(certs...), 0o644)
}

func WriteKey(path string, key crypto.Signer, perm os.FileMode) error {
	data, err := ca.EncodeKey(key)
	if err != nil {
		return err
	}

	return writeNew(path, data, perm)
}

func writeNew(path string, data []byte, perm os.FileMode) error {
	dir, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer func() { _ = dir.Close() }()

	f, err := dir.OpenFile(filepath.Base(path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	// The umask can remove permissions from the mode of OpenFile.
	return errors.Join(err, f.Chmod(perm), f.Close())
}

func ReadCSR(path string) (*x509.CertificateRequest, error) {
	data, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	csr, err := ca.ParseCSR(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return csr, nil
}
