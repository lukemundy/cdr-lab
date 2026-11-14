package domain

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"slices"
)

// ErrNotValid matches an error for a value that is not valid in a command, for example a thumbprint.
var ErrNotValid = errors.New("is not valid")

type notValidError struct{ err error }

func (e notValidError) Error() string { return e.err.Error() }

func (e notValidError) Unwrap() error { return e.err }

func (notValidError) Is(target error) bool { return target == ErrNotValid }

const (
	thumbprintNotValid  = "a client certificate thumbprint is not a base64url SHA-256 value"
	thumbprintNotUnique = "a client certificate thumbprint is not unique"
)

type RegisterClientCertificate struct {
	SoftwareProductID string
	// Thumbprint has the format of the RFC 8705 x5t#S256 value.
	Thumbprint string
}

func (RegisterClientCertificate) isDataRecipientCommand() {}

type ClientCertificateRegistered struct {
	SoftwareProductID string `json:"softwareProductId"`
	Thumbprint        string `json:"thumbprint"`
}

func (ClientCertificateRegistered) isDataRecipientEvent() {}

func handleRegisterClientCertificate(
	state DataRecipientState,
	c RegisterClientCertificate,
) ([]DataRecipientEvent, error) {
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

	if slices.Contains(sp.ClientCertificates, c.Thumbprint) {
		return nil, nil
	}

	return []DataRecipientEvent{ClientCertificateRegistered(c)}, nil
}

func registrationRule(sp SoftwareProduct) error {
	if sp.Status != Removed {
		return nil
	}

	product := record{kind: "software product", id: sp.SoftwareProductID}

	return StatusRuleError{fmt.Errorf("%s cannot get a client certificate: %s is final", product, Removed)}
}

type ClientCertificateAction string

const (
	RegisterCertificate ClientCertificateAction = "register"
	IssueCertificate    ClientCertificateAction = "issue"
	RemoveCertificate   ClientCertificateAction = "remove"
)

// PermittedClientCertificateActions gives an issue with the rule of a registration, and a removal in each status.
func PermittedClientCertificateActions(sp SoftwareProduct) []ClientCertificateAction {
	if registrationRule(sp) != nil {
		return []ClientCertificateAction{RemoveCertificate}
	}

	return []ClientCertificateAction{RegisterCertificate, IssueCertificate, RemoveCertificate}
}

type RemoveClientCertificate struct {
	SoftwareProductID string
	// Thumbprint has the format of the RFC 8705 x5t#S256 value.
	Thumbprint string
}

func (RemoveClientCertificate) isDataRecipientCommand() {}

type ClientCertificateRemoved struct {
	SoftwareProductID string `json:"softwareProductId"`
	Thumbprint        string `json:"thumbprint"`
}

func (ClientCertificateRemoved) isDataRecipientEvent() {}

// A removal is permitted in each status, because it only takes trust away.
func handleRemoveClientCertificate(
	state DataRecipientState,
	c RemoveClientCertificate,
) ([]DataRecipientEvent, error) {
	product := record{kind: "software product", id: c.SoftwareProductID}

	sp, ok := state.DataRecipient.softwareProduct(c.SoftwareProductID)
	if !ok {
		return nil, fmt.Errorf("%s %w", product, ErrNotFound)
	}

	if !validThumbprint(c.Thumbprint) {
		return nil, thumbprintError(product)
	}

	if !slices.Contains(sp.ClientCertificates, c.Thumbprint) {
		return nil, nil
	}

	return []DataRecipientEvent{ClientCertificateRemoved(c)}, nil
}

func thumbprintError(r record) error {
	return notValidError{fmt.Errorf("%s: %s", r, thumbprintNotValid)}
}

// ClientCertificateThumbprint gives the RFC 8705 x5t#S256 value of data, which must be one PEM certificate and no
// other text. The error does not contain data.
func ClientCertificateThumbprint(data []byte) (string, error) {
	notPEM := notValidError{errors.New("the text is not one PEM certificate")}

	block, rest := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" || len(bytes.TrimSpace(rest)) != 0 {
		return "", notPEM
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", notPEM
	}

	return CertificateThumbprint(cert), nil
}

// CertificateThumbprint gives the RFC 8705 x5t#S256 value of cert.
func CertificateThumbprint(cert *x509.Certificate) string {
	sum := sha256.Sum256(cert.Raw)

	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func validThumbprint(s string) bool {
	sum, err := base64.RawURLEncoding.Strict().DecodeString(s)
	return err == nil && len(sum) == sha256.Size
}

func (p *problems) clientCertificates(product record, thumbprints []string) {
	seen := map[string]bool{}

	for _, thumbprint := range thumbprints {
		switch {
		case !validThumbprint(thumbprint):
			p.add(product, "clientCertificates", "%s", thumbprintNotValid)
		case seen[thumbprint]:
			p.add(product, "clientCertificates", "%s", thumbprintNotUnique)
		}

		seen[thumbprint] = true
	}
}

func (dr DataRecipient) withClientCertificate(id, thumbprint string) DataRecipient {
	return dr.withSoftwareProduct(id, func(sp *SoftwareProduct) {
		sp.ClientCertificates = append(slices.Clone(sp.ClientCertificates), thumbprint)
	})
}

func (dr DataRecipient) withoutClientCertificate(id, thumbprint string) DataRecipient {
	return dr.withSoftwareProduct(id, func(sp *SoftwareProduct) {
		sp.ClientCertificates = slices.DeleteFunc(slices.Clone(sp.ClientCertificates), func(t string) bool {
			return t == thumbprint
		})
	})
}
