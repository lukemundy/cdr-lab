package ca

import (
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"math/big"
	"time"
)

// crlValidityDays is the time from the issue of a CRL to its next update, the
// same as the production CRL of the CDR Intermediate CA 2025.
const crlValidityDays = 7

// Reason is an RFC 5280 CRL reason code.
type Reason int

const (
	Unspecified          Reason = 0
	KeyCompromise        Reason = 1
	AffiliationChanged   Reason = 3
	Superseded           Reason = 4
	CessationOfOperation Reason = 5
	PrivilegeWithdrawn   Reason = 9
)

// The other RFC 5280 reasons are for CA certificates, attribute certificates
// or suspension, which the CDR CA does not support.
var reasons = map[string]Reason{
	"unspecified":          Unspecified,
	"keyCompromise":        KeyCompromise,
	"affiliationChanged":   AffiliationChanged,
	"superseded":           Superseded,
	"cessationOfOperation": CessationOfOperation,
	"privilegeWithdrawn":   PrivilegeWithdrawn,
}

func (r Reason) String() string {
	for name, reason := range reasons {
		if reason == r {
			return name
		}
	}

	return fmt.Sprintf("reason %d", int(r))
}

func ParseReason(name string) (Reason, error) {
	r, ok := reasons[name]
	if !ok {
		return 0, fmt.Errorf(
			"reason %q is not unspecified, keyCompromise, affiliationChanged, superseded, cessationOfOperation "+
				"or privilegeWithdrawn",
			name,
		)
	}

	return r, nil
}

type Revocation struct {
	Serial *big.Int
	Time   time.Time
	Reason Reason
}

// CRL gives a DER CRL that lists the revoked certificates.
func (a Authority) CRL(revoked []Revocation, number *big.Int, now time.Time) ([]byte, error) {
	entries := make([]x509.RevocationListEntry, 0, len(revoked))
	for _, r := range revoked {
		entries = append(
			entries,
			x509.RevocationListEntry{SerialNumber: r.Serial, RevocationTime: r.Time, ReasonCode: int(r.Reason)},
		)
	}

	return x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:                    number,
		ThisUpdate:                now,
		NextUpdate:                now.AddDate(0, 0, crlValidityDays),
		RevokedCertificateEntries: entries,
	}, a.Certificate, a.Key)
}
