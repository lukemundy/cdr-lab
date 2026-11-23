// Package certissue gets a CDR certificate through the Register issue path and writes it to an output, a TLS Secret
// or files, only when the output has no current certificate.
package certissue

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certcheck"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

type Config struct {
	Kind          ca.Kind
	LegalEntityID string
	// RecordID is the brand ID of a server certificate, or the software product ID of a client certificate.
	RecordID     string
	CommonName   string
	Organisation string
	DNSNames     []string
	// CSR is a PEM CSR from the user. The command then sends it as it is, takes the subject and the DNS names from
	// it, and writes no key.
	CSR         []byte
	RenewBefore time.Duration
	Now         func() time.Time
}

type Output interface {
	// Load gives no chain when the output has no certificate that the command can read. The key is nil when the
	// output has no key that the command can read.
	Load(ctx context.Context) (chain []*x509.Certificate, key crypto.PublicKey, err error)
	// Save gets no key when the user gave the CSR.
	Save(ctx context.Context, chain []*x509.Certificate, keyPEM []byte, bundle []*x509.Certificate) error
}

// Run gives the new certificate, or nil when the output has a current certificate.
func Run(ctx context.Context, out Output, reg *register.Client, cfg Config) (*x509.Certificate, error) {
	cfg, csrKey, err := withCSRNames(cfg)
	if err != nil {
		return nil, err
	}

	bundle, err := reg.CABundle(ctx)
	if err != nil {
		return nil, err
	}

	chain, key, err := out.Load(ctx)
	if err != nil {
		return nil, err
	}

	if csrKey != nil {
		key = csrKey
	}

	if certcheck.Current(chain, bundle, want(cfg, key)) {
		return nil, nil
	}

	csr, keyPEM, err := request(cfg)
	if err != nil {
		return nil, err
	}

	chain, err = issue(ctx, reg, cfg, csr)
	if err != nil {
		return nil, err
	}

	if err := out.Save(ctx, chain, keyPEM, bundle); err != nil {
		return nil, err
	}

	return chain[0], nil
}

// withCSRNames gives cfg with the subject and the DNS names of the CSR of the user, and the key of the CSR. With no
// CSR, it gives cfg and no key.
func withCSRNames(cfg Config) (Config, crypto.PublicKey, error) {
	if cfg.CSR == nil {
		return cfg, nil, nil
	}

	csr, err := ca.ParseCSR(cfg.CSR)
	if err != nil {
		return cfg, nil, fmt.Errorf("the CSR file: %w", err)
	}

	cfg.CommonName, cfg.DNSNames, cfg.Organisation = csr.Subject.CommonName, csr.DNSNames, ""
	if len(csr.Subject.Organization) > 0 {
		cfg.Organisation = csr.Subject.Organization[0]
	}

	return cfg, csr.PublicKey, nil
}

// request gives the CSR of the user, or a CSR and the PEM of a new key.
func request(cfg Config) (csr, keyPEM []byte, err error) {
	if cfg.CSR != nil {
		return cfg.CSR, nil, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, ca.LeafKeyBits)
	if err != nil {
		return nil, nil, err
	}

	csr, err = ca.NewCSR(key, cfg.Kind, cfg.CommonName, cfg.Organisation, cfg.DNSNames)
	if err != nil {
		return nil, nil, fmt.Errorf("the inputs do not give a CSR with the CDS profile:\n%w", err)
	}

	keyPEM, err = ca.EncodeKey(key)

	return csr, keyPEM, err
}

func issue(ctx context.Context, reg *register.Client, cfg Config, csr []byte) ([]*x509.Certificate, error) {
	if cfg.Kind == ca.Server {
		return reg.IssueServer(ctx, cfg.LegalEntityID, cfg.RecordID, csr)
	}

	return reg.IssueClient(ctx, cfg.LegalEntityID, cfg.RecordID, csr)
}

// want has the CN of a server certificate also as a DNS name, because the CA adds the CN of a server CSR to its DNS
// names.
func want(cfg Config, key crypto.PublicKey) certcheck.Want {
	w := certcheck.Want{
		Usage:        x509.ExtKeyUsageClientAuth,
		CommonName:   cfg.CommonName,
		Organisation: []string{cfg.Organisation},
		DNSNames:     []string{},
		PublicKey:    key,
		RenewBefore:  cfg.RenewBefore,
		Now:          cfg.Now(),
	}

	if cfg.Kind == ca.Server {
		w.Usage = x509.ExtKeyUsageServerAuth

		w.DNSNames = append([]string{cfg.CommonName}, cfg.DNSNames...)
	}

	return w
}
