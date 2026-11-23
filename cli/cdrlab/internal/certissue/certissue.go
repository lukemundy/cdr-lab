// Package certissue gets a CDR certificate through the Register issue path and writes it to a TLS Secret, only when the
// Secret has no current certificate.
package certissue

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"slices"
	"time"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

const DefaultRenewBefore = 30 * 24 * time.Hour

type Config struct {
	Kind          ca.Kind
	LegalEntityID string
	// RecordID is the brand ID of a server certificate, or the software product ID of a client certificate.
	RecordID     string
	CommonName   string
	Organisation string
	DNSNames     []string
	Secret       string
	RenewBefore  time.Duration
	Now          func() time.Time
}

// Run gives the new certificate, or nil when the Secret has a current certificate.
func Run(ctx context.Context, secrets *kube.Client, reg *register.Client, cfg Config) (*x509.Certificate, error) {
	bundle, err := reg.CABundle(ctx)
	if err != nil {
		return nil, err
	}

	secret, found, err := secrets.Get(ctx, cfg.Secret)
	if err != nil {
		return nil, err
	}

	if found && secret.Type != kube.TLSType {
		return nil, fmt.Errorf("the Secret %q has the type %q, not %s", cfg.Secret, secret.Type, kube.TLSType)
	}

	if found && current(secret, bundle, cfg) {
		return nil, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, ca.LeafKeyBits)
	if err != nil {
		return nil, err
	}

	csr, err := ca.NewCSR(key, cfg.Kind, cfg.CommonName, cfg.Organisation, cfg.DNSNames)
	if err != nil {
		return nil, fmt.Errorf("the inputs do not give a CSR with the CDS profile:\n%w", err)
	}

	chain, err := issue(ctx, reg, cfg, csr)
	if err != nil {
		return nil, err
	}

	keyPEM, err := ca.EncodeKey(key)
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		"tls.crt": ca.EncodeCertificates(chain...),
		"tls.key": keyPEM,
		"ca.crt":  ca.EncodeCertificates(bundle...),
	}

	if found {
		secret.Data = data
		err = secrets.Update(ctx, secret)
	} else {
		err = secrets.Create(ctx, kube.Secret{Name: cfg.Secret, Type: kube.TLSType, Data: data})
	}

	if err != nil {
		return nil, err
	}

	return chain[0], nil
}

func issue(ctx context.Context, reg *register.Client, cfg Config, csr []byte) ([]*x509.Certificate, error) {
	if cfg.Kind == ca.Server {
		return reg.IssueServer(ctx, cfg.LegalEntityID, cfg.RecordID, csr)
	}

	return reg.IssueClient(ctx, cfg.LegalEntityID, cfg.RecordID, csr)
}

func current(secret kube.Secret, bundle []*x509.Certificate, cfg Config) bool {
	chain, err := ca.ParseCertificates(secret.Data["tls.crt"])
	if err != nil {
		return false
	}

	key, err := ca.ParseKey(secret.Data["tls.key"])
	if err != nil {
		return false
	}

	cert, now := chain[0], cfg.Now()

	return cert.NotAfter.After(now.Add(cfg.RenewBefore)) &&
		chainsTo(cert, slices.Concat(chain[1:], bundle), now, cfg.Kind) &&
		cert.Subject.CommonName == cfg.CommonName &&
		slices.Equal(cert.Subject.Organization, []string{cfg.Organisation}) &&
		sameNames(cert.DNSNames, cfg) &&
		sameKey(cert.PublicKey, key.Public())
}

func chainsTo(cert *x509.Certificate, others []*x509.Certificate, now time.Time, kind ca.Kind) bool {
	roots, intermediates := x509.NewCertPool(), x509.NewCertPool()

	for _, c := range others {
		if c.IsCA && c.CheckSignatureFrom(c) == nil {
			roots.AddCert(c)
		} else {
			intermediates.AddCert(c)
		}
	}

	usage := x509.ExtKeyUsageServerAuth
	if kind == ca.Client {
		usage = x509.ExtKeyUsageClientAuth
	}

	_, err := cert.Verify(x509.VerifyOptions{
		Roots: roots, Intermediates: intermediates, CurrentTime: now, KeyUsages: []x509.ExtKeyUsage{usage},
	})

	return err == nil
}

// sameNames compares as sets, because the CA adds the CN of a server CSR to its DNS names.
func sameNames(names []string, cfg Config) bool {
	want := []string{}
	if cfg.Kind == ca.Server {
		want = append(want, cfg.CommonName)
		want = append(want, cfg.DNSNames...)
	}

	return slices.Equal(unique(names), unique(want))
}

func unique(names []string) []string {
	sorted := slices.Sorted(slices.Values(names))
	return slices.Compact(sorted)
}

func sameKey(a, b crypto.PublicKey) bool {
	k, ok := a.(interface{ Equal(x crypto.PublicKey) bool })
	return ok && k.Equal(b)
}
