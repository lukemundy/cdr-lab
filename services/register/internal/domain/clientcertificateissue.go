package domain

import (
	"crypto/x509/pkix"
	"fmt"
	"unicode/utf8"
)

// IssueClientCertificate records and registers a certificate that the mock CDR intermediate issued.
type IssueClientCertificate struct {
	SoftwareProductID string
	// Thumbprint has the format of the RFC 8705 x5t#S256 value.
	Thumbprint string
}

func (IssueClientCertificate) isDataRecipientCommand() {}

type ClientCertificateIssued struct {
	SoftwareProductID string `json:"softwareProductId"`
	Thumbprint        string `json:"thumbprint"`
}

func (ClientCertificateIssued) isDataRecipientEvent() {}

func handleIssueClientCertificate(state DataRecipientState, c IssueClientCertificate) ([]DataRecipientEvent, error) {
	product := record{kind: "software product", id: c.SoftwareProductID}

	sp, ok := state.DataRecipient.softwareProduct(c.SoftwareProductID)
	if !ok {
		return nil, fmt.Errorf("%s %w", product, ErrNotFound)
	}

	if !validThumbprint(c.Thumbprint) {
		return nil, thumbprintError(product)
	}

	if err := registrationRule(sp); err != nil {
		return nil, err
	}

	return []DataRecipientEvent{ClientCertificateIssued(c)}, nil
}

// ClientCertificateSubjectRules gives the CSR rules that only the Register can check, or an error when the software
// product cannot get a certificate.
func ClientCertificateSubjectRules(
	state DataRecipientState,
	softwareProductID string,
	subject pkix.Name,
) ([]string, error) {
	brand, sp, ok := state.DataRecipient.brandAndSoftwareProduct(softwareProductID)
	if !ok {
		return nil, fmt.Errorf("%s %w", record{kind: "software product", id: softwareProductID}, ErrNotFound)
	}

	if err := registrationRule(sp); err != nil {
		return nil, err
	}

	var rules []string

	switch {
	case utf8.RuneCountInString(sp.SoftwareProductName) > maxCertificateNameLength:
		rules = append(rules, fmt.Sprintf(
			"the software product name has more than %d characters, and a certificate CN can have %[1]d or fewer",
			maxCertificateNameLength,
		))
	case subject.CommonName != sp.SoftwareProductName:
		rules = append(rules, fmt.Sprintf("subject CN must be the software product name %q", sp.SoftwareProductName))
	}

	switch {
	case utf8.RuneCountInString(brand.BrandName) > maxCertificateNameLength:
		rules = append(rules, fmt.Sprintf(
			"the brand name has more than %d characters, and a certificate O can have %[1]d or fewer",
			maxCertificateNameLength,
		))
	case len(subject.Organization) != 1 || subject.Organization[0] != brand.BrandName:
		rules = append(rules, fmt.Sprintf("subject O must be the brand name %q", brand.BrandName))
	}

	return rules, nil
}
