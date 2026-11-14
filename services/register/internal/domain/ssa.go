package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ssaIssuer        = "cdr-register"
	ssaSoftwareRoles = "data-recipient-software-product"
	// SSALifetime is the expiry time that the CDS gives for a Software Statement Assertion.
	SSALifetime = 10 * time.Minute
)

var ErrNotActive = errors.New("is not ACTIVE")

type SSAClaims struct {
	Issuer              string   `json:"iss"`
	IssuedAt            int64    `json:"iat"`
	ExpiresAt           int64    `json:"exp"`
	ID                  string   `json:"jti"`
	LegalEntityID       string   `json:"legal_entity_id"`
	LegalEntityName     string   `json:"legal_entity_name"`
	OrgID               string   `json:"org_id"`
	OrgName             string   `json:"org_name"`
	ClientName          string   `json:"client_name"`
	ClientDescription   string   `json:"client_description"`
	ClientURI           string   `json:"client_uri"`
	RedirectURIs        []string `json:"redirect_uris"`
	SectorIdentifierURI string   `json:"sector_identifier_uri,omitempty"`
	LogoURI             string   `json:"logo_uri"`
	TOSURI              string   `json:"tos_uri,omitempty"`
	PolicyURI           string   `json:"policy_uri,omitempty"`
	JWKSURI             string   `json:"jwks_uri"`
	RevocationURI       string   `json:"revocation_uri"`
	RecipientBaseURI    string   `json:"recipient_base_uri"`
	SoftwareID          string   `json:"software_id"`
	SoftwareRoles       string   `json:"software_roles"`
	Scope               string   `json:"scope"`
}

// SSA gives the claims only for an ACTIVE software product of an ACTIVE data recipient.
func (dr DataRecipient) SSA(softwareProductID, jti string, issuedAt time.Time) (SSAClaims, error) {
	b, sp, ok := dr.brandAndSoftwareProduct(softwareProductID)
	if !ok {
		return SSAClaims{}, fmt.Errorf("%s %w", record{kind: "software product", id: softwareProductID}, ErrNotFound)
	}

	if dr.Status != DataRecipientActive {
		return SSAClaims{}, fmt.Errorf("%s %w", record{kind: "data recipient", id: dr.LegalEntityID}, ErrNotActive)
	}

	if sp.Status != Active {
		return SSAClaims{}, fmt.Errorf("%s %w", record{kind: "software product", id: softwareProductID}, ErrNotActive)
	}

	return SSAClaims{
		Issuer:              ssaIssuer,
		IssuedAt:            issuedAt.Unix(),
		ExpiresAt:           issuedAt.Add(SSALifetime).Unix(),
		ID:                  jti,
		LegalEntityID:       dr.LegalEntityID,
		LegalEntityName:     dr.LegalEntityName,
		OrgID:               b.DataRecipientBrandID,
		OrgName:             b.BrandName,
		ClientName:          sp.SoftwareProductName,
		ClientDescription:   sp.SoftwareProductDescription,
		ClientURI:           sp.ClientURI,
		RedirectURIs:        sp.RedirectURIs,
		SectorIdentifierURI: sp.SectorIdentifierURI,
		LogoURI:             sp.LogoURI,
		TOSURI:              sp.TOSURI,
		PolicyURI:           sp.PolicyURI,
		JWKSURI:             sp.JWKSURI,
		RevocationURI:       sp.RevocationURI,
		RecipientBaseURI:    sp.RecipientBaseURI,
		SoftwareID:          sp.SoftwareProductID,
		SoftwareRoles:       ssaSoftwareRoles,
		Scope:               strings.Join(sp.Scopes, " "),
	}, nil
}
