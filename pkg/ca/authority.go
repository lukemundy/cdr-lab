package ca

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"slices"
	"time"
)

const (
	rootValidityYears         = 20
	intermediateValidityYears = 10
	// leafValidityMonths is the validity of a production CDR certificate from the CDR Intermediate CA 2025.
	leafValidityMonths = 13
)

type Authority struct {
	Certificate *x509.Certificate
	Key         crypto.Signer
	// CRLURL is the URL of the CRL of this authority, which each certificate that it issues gives.
	CRLURL string
}

type Lab struct {
	CDRRoot         Authority
	CDRIntermediate Authority
	PublicRoot      Authority
}

type LabKeys struct {
	CDRRoot, CDRIntermediate, PublicRoot crypto.Signer
}

var (
	cdrRootSubject = pkix.Name{
		Country:            []string{"AU"},
		Organization:       []string{"cdr-lab"},
		OrganizationalUnit: []string{cdrOrganisationalUnit},
		CommonName:         "Mock CDR Root CA",
	}
	cdrIntermediateSubject = pkix.Name{
		Country:            []string{"AU"},
		Organization:       []string{"cdr-lab"},
		OrganizationalUnit: []string{cdrOrganisationalUnit},
		CommonName:         "Mock CDR Intermediate CA",
	}
	publicRootSubject = pkix.Name{
		Country:      []string{"AU"},
		Organization: []string{"cdr-lab"},
		CommonName:   "Mock Public Root CA",
	}
)

func NewLab(keys LabKeys, now time.Time) (Lab, error) {
	cdrRoot, err := newRoot(cdrRootSubject, keys.CDRRoot, now)
	if err != nil {
		return Lab{}, err
	}

	cdrIntermediate, err := cdrRoot.issue(&x509.Certificate{
		Subject:               cdrIntermediateSubject,
		NotBefore:             now,
		NotAfter:              now.AddDate(intermediateValidityYears, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}, keys.CDRIntermediate.Public())
	if err != nil {
		return Lab{}, err
	}

	publicRoot, err := newRoot(publicRootSubject, keys.PublicRoot, now)
	if err != nil {
		return Lab{}, err
	}

	return Lab{
		CDRRoot:         cdrRoot,
		CDRIntermediate: Authority{Certificate: cdrIntermediate, Key: keys.CDRIntermediate},
		PublicRoot:      publicRoot,
	}, nil
}

func newRoot(subject pkix.Name, key crypto.Signer, now time.Time) (Authority, error) {
	tmpl := &x509.Certificate{
		Subject:               subject,
		NotBefore:             now,
		NotAfter:              now.AddDate(rootValidityYears, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	cert, err := Authority{Key: key}.issue(tmpl, key.Public())

	return Authority{Certificate: cert, Key: key}, err
}

// Issue gives a CDR certificate for a CSR that follows the CDS profile.
func (a Authority) Issue(csr *x509.CertificateRequest, kind Kind, now time.Time) (*x509.Certificate, error) {
	if err := ValidateCSR(csr, kind); err != nil {
		return nil, err
	}

	var names []string
	if kind == Server {
		names = append([]string{csr.Subject.CommonName}, csr.DNSNames...)
	}

	return a.issue(a.leaf(csr.Subject, kind, names, now), csr.PublicKey)
}

// IssuePublicServer gives a TLS server certificate for lab hosts, with the first name as the subject CN.
func (a Authority) IssuePublicServer(key crypto.PublicKey, names []string, now time.Time) (*x509.Certificate, error) {
	if len(names) == 0 {
		return nil, errors.New("at least one DNS name is required")
	}

	errs := hostNameRules("DNS name", names[0], maxCommonNameLength)
	for _, name := range names[1:] {
		errs = append(errs, hostNameRules("DNS name", name, maxHostNameLength)...)
	}

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	return a.issue(a.leaf(pkix.Name{CommonName: names[0]}, Server, names, now), key)
}

func (a Authority) leaf(subject pkix.Name, kind Kind, names []string, now time.Time) *x509.Certificate {
	usage := x509.ExtKeyUsageServerAuth
	if kind == Client {
		usage = x509.ExtKeyUsageClientAuth
	}

	notAfter := now.AddDate(0, leafValidityMonths, 0)
	if a.Certificate.NotAfter.Before(notAfter) {
		notAfter = a.Certificate.NotAfter
	}

	var unique []string
	for _, name := range names {
		if !slices.Contains(unique, name) {
			unique = append(unique, name)
		}
	}

	var crlURLs []string
	if a.CRLURL != "" {
		crlURLs = []string{a.CRLURL}
	}

	return &x509.Certificate{
		Subject:               subject,
		DNSNames:              unique,
		CRLDistributionPoints: crlURLs,
		NotBefore:             now,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{usage},
		BasicConstraintsValid: true,
	}
}

// issue signs tmpl with the key of a, or self-signs it when a has no certificate.
func (a Authority) issue(tmpl *x509.Certificate, pub crypto.PublicKey) (*x509.Certificate, error) {
	serial, err := serialNumber()
	if err != nil {
		return nil, err
	}

	tmpl.SerialNumber = serial

	parent := a.Certificate
	if parent == nil {
		parent = tmpl
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, parent, pub, a.Key)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(der)
}

// serialNumber gives 126 random bits (the CA/Browser Forum requires 64) in a positive number of a constant length.
func serialNumber() (*big.Int, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	b[0] = b[0]&0x7f | 0x40

	return new(big.Int).SetBytes(b), nil
}
