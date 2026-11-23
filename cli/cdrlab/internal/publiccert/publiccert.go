// Package publiccert issues a public TLS server certificate from the mock public CA in a directory, only when the
// certificate file has no current certificate. It makes no network calls.
package publiccert

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"path/filepath"
	"time"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certcheck"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/pemfile"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

type Config struct {
	// Dir has the files of the CA set-up.
	Dir     string
	Cert    string
	Key     string
	KeyPerm os.FileMode
	// Names are the DNS names. The first name is the CN.
	Names       []string
	RenewBefore time.Duration
	Now         func() time.Time
}

// Run gives the new certificate, or nil when the certificate file has a current certificate.
func Run(cfg Config) (*x509.Certificate, error) {
	roots, err := ca.ReadCertificates(filepath.Join(cfg.Dir, "public/root.crt"))
	if err != nil {
		return nil, err
	}

	rootKey, err := ca.ReadKey(filepath.Join(cfg.Dir, "public/root.key"))
	if err != nil {
		return nil, err
	}

	if current(cfg, roots) {
		return nil, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, ca.LeafKeyBits)
	if err != nil {
		return nil, err
	}

	root := ca.Authority{Certificate: roots[0], Key: rootKey}

	// X.509 times have no fractional seconds.
	cert, err := root.IssuePublicServer(&key.PublicKey, cfg.Names, cfg.Now().UTC().Truncate(time.Second))
	if err != nil {
		return nil, err
	}

	keyPEM, err := ca.EncodeKey(key)
	if err != nil {
		return nil, err
	}

	if err := pemfile.Write(cfg.Key, keyPEM, cfg.KeyPerm); err != nil {
		return nil, err
	}

	return cert, pemfile.Write(cfg.Cert, ca.EncodeCertificates(cert), pemfile.AnyUser)
}

func current(cfg Config, roots []*x509.Certificate) bool {
	certPEM, _, certErr := pemfile.Read(cfg.Cert)
	keyPEM, _, keyErr := pemfile.Read(cfg.Key)

	chain, err := ca.ParseCertificates(certPEM)
	if certErr != nil || keyErr != nil || err != nil {
		return false
	}

	var public crypto.PublicKey
	if key, err := ca.ParseKey(keyPEM); err == nil {
		public = key.Public()
	}

	return certcheck.Current(chain, roots[:1], certcheck.Want{
		Usage:       x509.ExtKeyUsageServerAuth,
		CommonName:  cfg.Names[0],
		DNSNames:    cfg.Names,
		PublicKey:   public,
		RenewBefore: cfg.RenewBefore,
		Now:         cfg.Now(),
	})
}
