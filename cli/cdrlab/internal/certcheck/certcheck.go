// Package certcheck finds when the certificate of an output is current, so that a command does not issue it again.
package certcheck

import (
	"crypto"
	"crypto/x509"
	"slices"
	"time"
)

// DefaultRenewBefore is the renewal window of the commands (cli/cdrlab/DESIGN.md#cli-0101).
const DefaultRenewBefore = 30 * 24 * time.Hour

type Want struct {
	Usage      x509.ExtKeyUsage
	CommonName string
	// Organisation is not checked when it is nil.
	Organisation []string
	// DNSNames is compared as a set.
	DNSNames    []string
	PublicKey   crypto.PublicKey
	RenewBefore time.Duration
	Now         time.Time
}

// Current is true when the first certificate of chain expires after the renewal window, chains to trusted with the
// rest of chain, and has the names and the key of w.
func Current(chain, trusted []*x509.Certificate, w Want) bool {
	if len(chain) == 0 {
		return false
	}

	cert := chain[0]

	return cert.NotAfter.After(w.Now.Add(w.RenewBefore)) &&
		chainsTo(cert, slices.Concat(chain[1:], trusted), w) &&
		cert.Subject.CommonName == w.CommonName &&
		(w.Organisation == nil || slices.Equal(cert.Subject.Organization, w.Organisation)) &&
		slices.Equal(unique(cert.DNSNames), unique(w.DNSNames)) &&
		SameKey(cert.PublicKey, w.PublicKey)
}

func chainsTo(cert *x509.Certificate, others []*x509.Certificate, w Want) bool {
	roots, intermediates := x509.NewCertPool(), x509.NewCertPool()

	for _, c := range others {
		if c.IsCA && c.CheckSignatureFrom(c) == nil {
			roots.AddCert(c)
		} else {
			intermediates.AddCert(c)
		}
	}

	_, err := cert.Verify(x509.VerifyOptions{
		Roots: roots, Intermediates: intermediates, CurrentTime: w.Now, KeyUsages: []x509.ExtKeyUsage{w.Usage},
	})

	return err == nil
}

func unique(names []string) []string {
	return slices.Compact(slices.Sorted(slices.Values(names)))
}

func SameKey(a, b crypto.PublicKey) bool {
	k, ok := a.(interface{ Equal(x crypto.PublicKey) bool })
	return ok && b != nil && k.Equal(b)
}
