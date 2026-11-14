package app

import (
	"crypto/x509"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

const NotACertificate = "The certificate must be one PEM certificate, or a PEM chain with the certificate first."

// RevocationFieldError has the Field certificate or reason, and its Detail never contains the certificate.
type RevocationFieldError struct {
	Field  string
	Detail string
}

// ParseRevocationRequest gives the first certificate of the PEM text, and a nil value is a field that is not there.
func ParseRevocationRequest(certificate, reason *string) (*x509.Certificate, ca.Reason, []RevocationFieldError) {
	var (
		errs   []RevocationFieldError
		cert   *x509.Certificate
		parsed ca.Reason
	)

	if certificate == nil {
		errs = append(errs, RevocationFieldError{Field: "certificate", Detail: "The certificate is required."})
	} else if certs, err := ca.ParseCertificates([]byte(*certificate)); err != nil {
		errs = append(errs, RevocationFieldError{Field: "certificate", Detail: NotACertificate})
	} else {
		cert = certs[0]
	}

	if reason == nil {
		errs = append(errs, RevocationFieldError{Field: "reason", Detail: "The reason is required."})
	} else if r, err := ca.ParseReason(*reason); err != nil {
		errs = append(errs, RevocationFieldError{
			Field: "reason",
			Detail: "The reason must be unspecified, keyCompromise, affiliationChanged, superseded, " +
				"cessationOfOperation or privilegeWithdrawn.",
		})
	} else {
		parsed = r
	}

	if len(errs) > 0 {
		return nil, 0, errs
	}

	return cert, parsed, nil
}
