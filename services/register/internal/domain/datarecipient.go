package domain

import (
	"errors"
	"fmt"
	"slices"
)

type DataRecipientStatus string

const (
	DataRecipientActive      DataRecipientStatus = "ACTIVE"
	DataRecipientSuspended   DataRecipientStatus = "SUSPENDED"
	DataRecipientRevoked     DataRecipientStatus = "REVOKED"
	DataRecipientSurrendered DataRecipientStatus = "SURRENDERED"
)

var dataRecipientStatuses = []DataRecipientStatus{
	DataRecipientActive,
	DataRecipientSuspended,
	DataRecipientRevoked,
	DataRecipientSurrendered,
}

type AccreditationLevel string

const (
	Unrestricted AccreditationLevel = "UNRESTRICTED"
	Sponsored    AccreditationLevel = "SPONSORED"
)

type DataRecipient struct {
	LegalEntityID       string               `json:"legalEntityId"`
	LegalEntityName     string               `json:"legalEntityName"`
	AccreditationNumber string               `json:"accreditationNumber"`
	AccreditationLevel  AccreditationLevel   `json:"accreditationLevel"`
	LogoURI             string               `json:"logoUri"`
	Status              DataRecipientStatus  `json:"status"`
	Brands              []DataRecipientBrand `json:"dataRecipientBrands"`
}

func (dr DataRecipient) SoftwareProductCount() int {
	count := 0
	for _, b := range dr.Brands {
		count += len(b.SoftwareProducts)
	}

	return count
}

func (dr DataRecipient) IDs() []string {
	ids := []string{dr.LegalEntityID}
	for _, b := range dr.Brands {
		ids = append(ids, b.DataRecipientBrandID)
		for _, sp := range b.SoftwareProducts {
			ids = append(ids, sp.SoftwareProductID)
		}
	}

	return ids
}

type DataRecipientBrand struct {
	DataRecipientBrandID string            `json:"dataRecipientBrandId"`
	BrandName            string            `json:"brandName"`
	LogoURI              string            `json:"logoUri"`
	Status               Status            `json:"status"`
	SoftwareProducts     []SoftwareProduct `json:"softwareProducts"`
}

type SoftwareProduct struct {
	SoftwareProductID          string   `json:"softwareProductId"`
	SoftwareProductName        string   `json:"softwareProductName"`
	SoftwareProductDescription string   `json:"softwareProductDescription"`
	LogoURI                    string   `json:"logoUri"`
	Status                     Status   `json:"status"`
	ClientURI                  string   `json:"clientUri"`
	SectorIdentifierURI        string   `json:"sectorIdentifierUri,omitempty"`
	RedirectURIs               []string `json:"redirectUris"`
	TOSURI                     string   `json:"tosUri,omitempty"`
	PolicyURI                  string   `json:"policyUri,omitempty"`
	JWKSURI                    string   `json:"jwksUri"`
	RevocationURI              string   `json:"revocationUri"`
	RecipientBaseURI           string   `json:"recipientBaseUri"`
	Scopes                     []string `json:"scopes"`
	// ClientCertificates has the thumbprints of the client certificates
	// that the Register trusts for the software product.
	ClientCertificates []string `json:"clientCertificates,omitempty"`
}

type DataRecipientState struct {
	Imported      bool
	DataRecipient DataRecipient
}

type DataRecipientCommand interface{ isDataRecipientCommand() }

type ImportDataRecipient struct{ DataRecipient DataRecipient }

func (ImportDataRecipient) isDataRecipientCommand() {}

type DataRecipientEvent interface{ isDataRecipientEvent() }

type DataRecipientImported struct {
	DataRecipient DataRecipient `json:"dataRecipient"`
}

func (DataRecipientImported) isDataRecipientEvent() {}

func HandleDataRecipient(state DataRecipientState, command DataRecipientCommand) ([]DataRecipientEvent, error) {
	switch c := command.(type) {
	case ImportDataRecipient:
		return handleNewDataRecipient(state, c.DataRecipient, DataRecipientImported(c))
	case CreateDataRecipient:
		return handleNewDataRecipient(state, c.DataRecipient, DataRecipientCreated(c))
	}

	if !state.Imported {
		return nil, fmt.Errorf("data recipient %w", ErrNotFound)
	}

	switch c := command.(type) {
	case SuspendAccreditation:
		return handleDataRecipientStatus(state, DataRecipientSuspended, AccreditationSuspended{})
	case RevokeAccreditation:
		return handleDataRecipientStatus(state, DataRecipientRevoked, AccreditationRevoked{})
	case SurrenderAccreditation:
		return handleDataRecipientStatus(state, DataRecipientSurrendered, AccreditationSurrendered{})
	case ReinstateAccreditation:
		return handleDataRecipientStatus(state, DataRecipientActive, AccreditationReinstated{})
	case ActivateSoftwareProduct:
		return handleSoftwareProductStatus(state, c.SoftwareProductID, Active, SoftwareProductActivated(c))
	case DeactivateSoftwareProduct:
		return handleSoftwareProductStatus(state, c.SoftwareProductID, Inactive, SoftwareProductDeactivated(c))
	case RemoveSoftwareProduct:
		return handleSoftwareProductStatus(state, c.SoftwareProductID, Removed, SoftwareProductRemoved(c))
	case RegisterClientCertificate:
		return handleRegisterClientCertificate(state, c)
	case RemoveClientCertificate:
		return handleRemoveClientCertificate(state, c)
	case IssueClientCertificate:
		return handleIssueClientCertificate(state, c)
	default:
		return handleDataRecipientChange(state, command)
	}
}

func handleDataRecipientChange(state DataRecipientState, command DataRecipientCommand) ([]DataRecipientEvent, error) {
	switch c := command.(type) {
	case ChangeDataRecipientDetails:
		return handleDataRecipientEvent(state, DataRecipientDetailsChanged{DataRecipientDetails: c.Details})
	case AddDataRecipientBrand:
		return handleAddDataRecipientBrand(state, c.Brand)
	case ChangeDataRecipientBrandDetails:
		return handleDataRecipientBrandCommand(
			state,
			c.BrandID,
			func(DataRecipientBrand) ([]DataRecipientEvent, error) {
				return handleDataRecipientEvent(
					state,
					DataRecipientBrandDetailsChanged{BrandID: c.BrandID, DataRecipientBrandDetails: c.Details},
				)
			},
		)
	case ActivateDataRecipientBrand:
		return handleDataRecipientBrandStatus(state, c.BrandID, Active)
	case DeactivateDataRecipientBrand:
		return handleDataRecipientBrandStatus(state, c.BrandID, Inactive)
	case RemoveDataRecipientBrand:
		return handleDataRecipientBrandStatus(state, c.BrandID, Removed)
	case AddSoftwareProduct:
		return handleAddSoftwareProduct(state, c)
	case ChangeSoftwareProductDetails:
		return handleChangeSoftwareProductDetails(state, c)
	default:
		return nil, errors.New("unknown data recipient command")
	}
}

func ApplyDataRecipient(state DataRecipientState, event DataRecipientEvent) DataRecipientState {
	switch e := event.(type) {
	case DataRecipientImported:
		return DataRecipientState{Imported: true, DataRecipient: e.DataRecipient}
	case DataRecipientCreated:
		return DataRecipientState{Imported: true, DataRecipient: e.DataRecipient}
	case AccreditationSuspended:
		state.DataRecipient.Status = DataRecipientSuspended
	case AccreditationRevoked:
		state.DataRecipient.Status = DataRecipientRevoked
	case AccreditationSurrendered:
		state.DataRecipient.Status = DataRecipientSurrendered
	case AccreditationReinstated:
		state.DataRecipient.Status = DataRecipientActive
	case SoftwareProductActivated:
		state.DataRecipient = state.DataRecipient.withSoftwareProductStatus(e.SoftwareProductID, Active)
	case SoftwareProductDeactivated:
		state.DataRecipient = state.DataRecipient.withSoftwareProductStatus(e.SoftwareProductID, Inactive)
	case SoftwareProductRemoved:
		state.DataRecipient = state.DataRecipient.withSoftwareProductStatus(e.SoftwareProductID, Removed)
	case ClientCertificateRegistered:
		state.DataRecipient = state.DataRecipient.withClientCertificate(e.SoftwareProductID, e.Thumbprint)
	case ClientCertificateRemoved:
		state.DataRecipient = state.DataRecipient.withoutClientCertificate(e.SoftwareProductID, e.Thumbprint)
	case ClientCertificateIssued:
		state.DataRecipient = state.DataRecipient.withClientCertificate(e.SoftwareProductID, e.Thumbprint)
	default:
		state.DataRecipient = applyDataRecipientChange(state.DataRecipient, event)
	}

	return state
}

func (dr DataRecipient) validate() error {
	var p problems

	recipient := record{kind: "data recipient", id: dr.LegalEntityID}
	p.id(recipient, "legalEntityId", dr.LegalEntityID)
	p.requiredText(recipient, "legalEntityName", dr.LegalEntityName, MaxNameLength)
	p.requiredText(recipient, "accreditationNumber", dr.AccreditationNumber, MaxAccreditationNumberLength)

	if dr.AccreditationLevel != Unrestricted && dr.AccreditationLevel != Sponsored {
		p.add(
			recipient,
			"accreditationLevel",
			"accreditationLevel %q is not UNRESTRICTED or SPONSORED",
			dr.AccreditationLevel,
		)
	}

	p.requiredURI(recipient, "logoUri", dr.LogoURI)

	statusValid := slices.Contains(dataRecipientStatuses, dr.Status)
	if !statusValid {
		p.add(recipient, "status", "status %q is not ACTIVE, SUSPENDED, REVOKED or SURRENDERED", dr.Status)
	}

	for _, b := range dr.Brands {
		brand := record{kind: "data recipient brand", id: b.DataRecipientBrandID}
		p.id(brand, "dataRecipientBrandId", b.DataRecipientBrandID)
		p.requiredText(brand, "brandName", b.BrandName, MaxNameLength)
		p.requiredURI(brand, "logoUri", b.LogoURI)
		brandStatusValid := p.status(brand, b.Status)

		for _, sp := range b.SoftwareProducts {
			product := record{kind: "software product", id: sp.SoftwareProductID}
			p.softwareProductFields(product, sp)

			if !p.status(product, sp.Status) {
				continue
			}

			if statusValid {
				p.recipientStatusRule(product, sp.Status, recipient, dr.Status)
			}

			if brandStatusValid {
				p.brandStatusRule(product, sp.Status, brand, b.Status)
			}
		}
	}

	return p.err()
}

func (p *problems) softwareProductFields(product record, sp SoftwareProduct) {
	p.id(product, "softwareProductId", sp.SoftwareProductID)
	p.requiredText(product, "softwareProductName", sp.SoftwareProductName, MaxNameLength)
	p.requiredText(product, "softwareProductDescription", sp.SoftwareProductDescription, MaxDescriptionLength)
	p.requiredURI(product, "logoUri", sp.LogoURI)
	p.requiredURI(product, "clientUri", sp.ClientURI)
	p.optionalURI(product, "sectorIdentifierUri", sp.SectorIdentifierURI)
	p.requiredURIList(product, "redirectUris", sp.RedirectURIs)
	p.optionalURI(product, "tosUri", sp.TOSURI)
	p.optionalURI(product, "policyUri", sp.PolicyURI)
	p.requiredURI(product, "jwksUri", sp.JWKSURI)
	p.requiredURI(product, "revocationUri", sp.RevocationURI)
	p.requiredURI(product, "recipientBaseUri", sp.RecipientBaseURI)
	p.scopes(product, sp.Scopes)
	p.clientCertificates(product, sp.ClientCertificates)
}

func (p *problems) scopes(product record, scopes []string) {
	p.requiredList(product, "scopes", scopes)

	if len(scopes) == 0 {
		return
	}

	if slices.Contains(scopes, "") {
		p.add(product, "scopes", "scopes has an empty value")
	}

	if !slices.Contains(scopes, "openid") || !slices.Contains(scopes, "profile") {
		p.add(product, "scopes", "scopes does not contain openid and profile")
	}
}

func (p *problems) recipientStatusRule(
	product record,
	productStatus Status,
	recipient record,
	recipientStatus DataRecipientStatus,
) {
	final := recipientStatus == DataRecipientRevoked || recipientStatus == DataRecipientSurrendered

	switch {
	case recipientStatus == DataRecipientSuspended && productStatus == Active:
		p.statusRuleBroken(
			product,
			"is %s, but %s is %s: a software product of a %s data recipient must be INACTIVE or REMOVED",
			productStatus,
			recipient,
			recipientStatus,
			recipientStatus,
		)
	case final && productStatus != Removed:
		p.statusRuleBroken(
			product,
			"is %s, but %s is %s: a software product of a %s data recipient must be REMOVED",
			productStatus,
			recipient,
			recipientStatus,
			recipientStatus,
		)
	}
}

func (p *problems) brandStatusRule(product record, productStatus Status, brand record, brandStatus Status) {
	switch {
	case brandStatus == Inactive && productStatus == Active:
		p.statusRuleBroken(
			product,
			"is %s, but %s is %s: a software product of an INACTIVE data recipient brand must be INACTIVE or REMOVED",
			productStatus,
			brand,
			brandStatus,
		)
	case brandStatus == Removed && productStatus != Removed:
		p.statusRuleBroken(
			product,
			"is %s, but %s is %s: a software product of a REMOVED data recipient brand must be REMOVED",
			productStatus,
			brand,
			brandStatus,
		)
	}
}
