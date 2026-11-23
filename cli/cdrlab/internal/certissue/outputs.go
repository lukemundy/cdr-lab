package certissue

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/pemfile"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

// Secret is a TLS Secret with the certificate chain, the key and the CDR CA bundle.
type Secret struct {
	client *kube.Client
	name   string
	loaded kube.Secret
	found  bool
}

func NewSecret(client *kube.Client, name string) *Secret {
	return &Secret{client: client, name: name}
}

func (s *Secret) String() string { return fmt.Sprintf("the Secret %q", s.name) }

func (s *Secret) Load(ctx context.Context) ([]*x509.Certificate, crypto.PublicKey, error) {
	secret, found, err := s.client.Get(ctx, s.name)
	if err != nil {
		return nil, nil, err
	}

	if found && secret.Type != kube.TLSType {
		return nil, nil, fmt.Errorf("the Secret %q has the type %q, not %s", s.name, secret.Type, kube.TLSType)
	}

	s.loaded, s.found = secret, found

	return parse(secret.Data["tls.crt"], secret.Data["tls.key"])
}

func (s *Secret) Save(ctx context.Context, chain []*x509.Certificate, keyPEM []byte, bundle []*x509.Certificate) error {
	if keyPEM == nil {
		return errors.New("a TLS Secret needs a key, so it cannot take the certificate of a CSR file")
	}

	data := map[string][]byte{
		"tls.crt": ca.EncodeCertificates(chain...),
		"tls.key": keyPEM,
		"ca.crt":  ca.EncodeCertificates(bundle...),
	}

	if s.found {
		s.loaded.Data = data
		return s.client.Update(ctx, s.loaded)
	}

	return s.client.Create(ctx, kube.Secret{Name: s.name, Type: kube.TLSType, Data: data})
}

// Files is a certificate file with the chain and a key file. With a CSR from the user, Key is empty.
type Files struct {
	Cert    string
	Key     string
	KeyPerm os.FileMode
}

func (f Files) String() string { return "the file " + f.Cert }

func (f Files) Load(context.Context) ([]*x509.Certificate, crypto.PublicKey, error) {
	certPEM, found, err := pemfile.Read(f.Cert)
	if err != nil || !found {
		return nil, nil, err
	}

	var keyPEM []byte

	if f.Key != "" {
		if keyPEM, _, err = pemfile.Read(f.Key); err != nil {
			return nil, nil, err
		}
	}

	return parse(certPEM, keyPEM)
}

// Save writes the key first. If the write of the certificate then fails, the key does not agree with the
// certificate, and the next run gets a new certificate.
func (f Files) Save(_ context.Context, chain []*x509.Certificate, keyPEM []byte, _ []*x509.Certificate) error {
	if keyPEM != nil {
		if err := pemfile.Write(f.Key, keyPEM, f.KeyPerm); err != nil {
			return err
		}
	}

	return pemfile.Write(f.Cert, ca.EncodeCertificates(chain...), pemfile.AnyUser)
}

// parse gives no chain or no key for a text that it cannot read, so that the run gets a new certificate.
func parse(certPEM, keyPEM []byte) ([]*x509.Certificate, crypto.PublicKey, error) {
	chain, err := ca.ParseCertificates(certPEM)
	if err != nil {
		return nil, nil, nil
	}

	key, err := ca.ParseKey(keyPEM)
	if err != nil {
		return chain, nil, nil
	}

	return chain, key.Public(), nil
}
