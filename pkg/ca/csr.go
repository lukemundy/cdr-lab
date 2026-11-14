// Package ca creates the mock CAs and issues certificates in the profiles of the CDR PKI.
package ca

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"net/netip"
	"regexp"
)

type Kind int

const (
	Server Kind = iota + 1
	Client
)

const cdrOrganisationalUnit = "Consumer Data Right"

// LeafKeyBits is the RSA key size that the CDS Certificate Signing Request Profile requires.
const LeafKeyBits = 2048

// maxCommonNameLength is the upper bound of a CN in RFC 5280, Appendix A.
const maxCommonNameLength = 64

// maxHostNameLength is the upper bound of a name in RFC 1035, section 2.3.4, in its text form with no last dot.
const maxHostNameLength = 253

var (
	hostLabels  = regexp.MustCompile(`^(?i)[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*$`)
	countryCode = regexp.MustCompile(`^[A-Z]{2}$`)
)

// hostNameRules gives an error for each host name rule of the PKI DESIGN.md that name breaks, with field in the text.
func hostNameRules(field, name string, maxLength int) []error {
	var errs []error

	switch {
	case !hostLabels.MatchString(name):
		errs = append(errs, fmt.Errorf(
			"%s %q is not a host name: use only letters, digits and hyphens in labels of 1 to 63 characters, "+
				"with a dot between two labels and no hyphen at the start or end of a label",
			field, name,
		))
	case isIPAddress(name):
		errs = append(errs, fmt.Errorf("%s %q is an IP address, not a host name", field, name))
	}

	if len(name) > maxLength {
		errs = append(errs, fmt.Errorf("%s %q has more than %d characters", field, name, maxLength))
	}

	return errs
}

func isIPAddress(name string) bool {
	_, err := netip.ParseAddr(name)
	return err == nil
}

// ValidateCSR applies the CDS Certificate Signing Request Profile, and
// gives an error for each rule that the CSR breaks.
func ValidateCSR(csr *x509.CertificateRequest, kind Kind) error {
	var errs []error

	fail := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	if k, ok := csr.PublicKey.(*rsa.PublicKey); !ok || k.N.BitLen() != LeafKeyBits {
		fail("the key must be RSA %d, not %s", LeafKeyBits, keyName(csr.PublicKey))
	}

	if csr.SignatureAlgorithm != x509.SHA256WithRSA {
		fail("the signature algorithm must be %s, not %s", x509.SHA256WithRSA, csr.SignatureAlgorithm)
	}

	if err := csr.CheckSignature(); err != nil {
		fail("the CSR signature is not valid: %w", err)
	}

	subject := csr.Subject
	if subject.CommonName == "" {
		fail("subject CN is required")
	}

	if len(subject.Organization) == 0 || subject.Organization[0] == "" {
		fail("subject O is required")
	}

	if len(subject.OrganizationalUnit) != 1 || subject.OrganizationalUnit[0] != cdrOrganisationalUnit {
		fail("subject OU must be %q", cdrOrganisationalUnit)
	}

	switch {
	case len(subject.Country) == 0:
		fail("subject C is required")
	case !countryCode.MatchString(subject.Country[0]):
		fail("subject C %q is not a two-letter country code", subject.Country[0])
	}

	switch kind {
	case Server:
		if subject.CommonName != "" {
			errs = append(errs, hostNameRules("subject CN", subject.CommonName, maxCommonNameLength)...)
		}

		for _, name := range csr.DNSNames {
			errs = append(errs, hostNameRules("DNS name", name, maxHostNameLength)...)
		}

		if len(csr.IPAddresses)+len(csr.EmailAddresses)+len(csr.URIs) > 0 {
			fail("a server CSR can have only DNS names as subject alternative names")
		}
	case Client:
		if len(csr.DNSNames)+len(csr.IPAddresses)+len(csr.EmailAddresses)+len(csr.URIs) > 0 {
			fail("a client CSR must not have subject alternative names")
		}
	}

	return errors.Join(errs...)
}

func keyName(k any) string {
	switch k := k.(type) {
	case *rsa.PublicKey:
		return fmt.Sprintf("RSA %d", k.N.BitLen())
	case *ecdsa.PublicKey:
		return "ECDSA"
	case ed25519.PublicKey:
		return "Ed25519"
	default:
		return fmt.Sprintf("%T", k)
	}
}
