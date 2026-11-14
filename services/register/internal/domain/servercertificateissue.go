package domain

import (
	"crypto/x509/pkix"
	"fmt"
	"unicode/utf8"
)

// IssueServerCertificate records a server certificate that the mock CDR intermediate issued for a data holder brand.
type IssueServerCertificate struct {
	BrandID string
	// Thumbprint has the format of the RFC 8705 x5t#S256 value.
	Thumbprint string
}

func (IssueServerCertificate) isDataHolderCommand() {}

type ServerCertificateIssued struct {
	BrandID    string `json:"brandId"`
	Thumbprint string `json:"thumbprint"`
}

func (ServerCertificateIssued) isDataHolderEvent() {}

func handleIssueServerCertificate(state DataHolderState, c IssueServerCertificate) ([]DataHolderEvent, error) {
	return handleDataHolderBrandCommand(state, c.BrandID, func(id string) ([]DataHolderEvent, error) {
		b, _ := state.DataHolder.brand(id)
		if err := serverCertificateRule(b); err != nil {
			return nil, err
		}

		if !validThumbprint(c.Thumbprint) {
			return nil, thumbprintError(record{kind: "data holder brand", id: id})
		}

		return []DataHolderEvent{ServerCertificateIssued{BrandID: id, Thumbprint: c.Thumbprint}}, nil
	})
}

type ServerCertificateAction string

const IssueServerCertificateAction ServerCertificateAction = "issue"

func PermittedServerCertificateActions(b DataHolderBrand) []ServerCertificateAction {
	if serverCertificateRule(b) != nil {
		return []ServerCertificateAction{}
	}

	return []ServerCertificateAction{IssueServerCertificateAction}
}

// serverCertificateRule refuses a brand with no dataHolderBrandId, because it shares only product data and has no
// mTLS endpoint.
func serverCertificateRule(b DataHolderBrand) error {
	brand := record{kind: "data holder brand", id: b.ID()}

	switch {
	case b.Status == Removed:
		return StatusRuleError{fmt.Errorf("%s cannot get a server certificate: %s is final", brand, Removed)}
	case b.DataHolderBrandID == "":
		return StatusRuleError{fmt.Errorf(
			"%s cannot get a server certificate: it has no dataHolderBrandId, so it shares only product data",
			brand,
		)}
	default:
		return nil
	}
}

// ServerCertificateSubjectRules gives the CSR rules that only the Register can check, or an error when the brand
// cannot get a certificate.
func ServerCertificateSubjectRules(state DataHolderState, brandID string, subject pkix.Name) ([]string, error) {
	if !state.Imported {
		return nil, fmt.Errorf("data holder %w", ErrNotFound)
	}

	b, ok := state.DataHolder.brand(brandID)
	if !ok {
		return nil, fmt.Errorf("%s %w", record{kind: "data holder brand", id: brandID}, ErrNotFound)
	}

	if err := serverCertificateRule(b); err != nil {
		return nil, err
	}

	legalEntityName := state.DataHolder.LegalEntityName
	o := ""

	if len(subject.Organization) == 1 {
		o = subject.Organization[0]
	}

	// A legal entity name has 36 characters or fewer, so it can always be the O.
	if utf8.RuneCountInString(b.BrandName) > maxCertificateNameLength {
		if o == legalEntityName {
			return nil, nil
		}

		return []string{fmt.Sprintf(
			"subject O must be the legal entity name %q: the brand name has more than %d characters, and a "+
				"certificate O can have %[2]d or fewer",
			legalEntityName, maxCertificateNameLength,
		)}, nil
	}

	if o == b.BrandName || o == legalEntityName {
		return nil, nil
	}

	return []string{fmt.Sprintf(
		"subject O must be the brand name %q or the legal entity name %q",
		b.BrandName, legalEntityName,
	)}, nil
}
