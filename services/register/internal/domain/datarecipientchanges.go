package domain

import (
	"fmt"
	"slices"
)

type CreateDataRecipient struct{ DataRecipient DataRecipient }

// DataRecipientDetails has the fields of a data recipient legal entity that a change replaces.
type DataRecipientDetails struct {
	LegalEntityName     string             `json:"legalEntityName"`
	AccreditationNumber string             `json:"accreditationNumber"`
	AccreditationLevel  AccreditationLevel `json:"accreditationLevel"`
	LogoURI             string             `json:"logoUri"`
}

type ChangeDataRecipientDetails struct{ Details DataRecipientDetails }

type AddDataRecipientBrand struct{ Brand DataRecipientBrand }

// DataRecipientBrandDetails has the fields of a data recipient brand that a change replaces.
type DataRecipientBrandDetails struct {
	BrandName string `json:"brandName"`
	LogoURI   string `json:"logoUri"`
}

type ChangeDataRecipientBrandDetails struct {
	BrandID string
	Details DataRecipientBrandDetails
}

type ActivateDataRecipientBrand struct{ BrandID string }

type DeactivateDataRecipientBrand struct{ BrandID string }

type RemoveDataRecipientBrand struct{ BrandID string }

type AddSoftwareProduct struct {
	BrandID         string
	SoftwareProduct SoftwareProduct
}

// SoftwareProductDetails has the fields of a software product that a change replaces: each field except the ID, the
// status and the client certificates.
type SoftwareProductDetails struct {
	SoftwareProductName        string   `json:"softwareProductName"`
	SoftwareProductDescription string   `json:"softwareProductDescription"`
	LogoURI                    string   `json:"logoUri"`
	ClientURI                  string   `json:"clientUri"`
	SectorIdentifierURI        string   `json:"sectorIdentifierUri,omitempty"`
	RedirectURIs               []string `json:"redirectUris"`
	TOSURI                     string   `json:"tosUri,omitempty"`
	PolicyURI                  string   `json:"policyUri,omitempty"`
	JWKSURI                    string   `json:"jwksUri"`
	RevocationURI              string   `json:"revocationUri"`
	RecipientBaseURI           string   `json:"recipientBaseUri"`
	Scopes                     []string `json:"scopes"`
}

func (d SoftwareProductDetails) SoftwareProduct(id string, status Status) SoftwareProduct {
	sp := SoftwareProduct{SoftwareProductID: id, Status: status}
	d.applyTo(&sp)

	return sp
}

func (d SoftwareProductDetails) applyTo(sp *SoftwareProduct) {
	sp.SoftwareProductName, sp.SoftwareProductDescription, sp.LogoURI = d.SoftwareProductName,
		d.SoftwareProductDescription, d.LogoURI
	sp.ClientURI, sp.SectorIdentifierURI, sp.RedirectURIs = d.ClientURI, d.SectorIdentifierURI, d.RedirectURIs
	sp.TOSURI, sp.PolicyURI, sp.JWKSURI = d.TOSURI, d.PolicyURI, d.JWKSURI
	sp.RevocationURI, sp.RecipientBaseURI, sp.Scopes = d.RevocationURI, d.RecipientBaseURI, d.Scopes
}

type ChangeSoftwareProductDetails struct {
	SoftwareProductID string
	Details           SoftwareProductDetails
}

func (CreateDataRecipient) isDataRecipientCommand()             {}
func (ChangeDataRecipientDetails) isDataRecipientCommand()      {}
func (AddDataRecipientBrand) isDataRecipientCommand()           {}
func (ChangeDataRecipientBrandDetails) isDataRecipientCommand() {}
func (ActivateDataRecipientBrand) isDataRecipientCommand()      {}
func (DeactivateDataRecipientBrand) isDataRecipientCommand()    {}
func (RemoveDataRecipientBrand) isDataRecipientCommand()        {}
func (AddSoftwareProduct) isDataRecipientCommand()              {}
func (ChangeSoftwareProductDetails) isDataRecipientCommand()    {}

type DataRecipientCreated struct {
	DataRecipient DataRecipient `json:"dataRecipient"`
}

type DataRecipientDetailsChanged struct {
	DataRecipientDetails
}

type DataRecipientBrandAdded struct {
	Brand DataRecipientBrand `json:"brand"`
}

type DataRecipientBrandDetailsChanged struct {
	BrandID string `json:"brandId"`
	DataRecipientBrandDetails
}

type DataRecipientBrandActivated struct {
	BrandID string `json:"brandId"`
}

type DataRecipientBrandDeactivated struct {
	BrandID string `json:"brandId"`
}

type DataRecipientBrandRemoved struct {
	BrandID string `json:"brandId"`
}

type SoftwareProductAdded struct {
	BrandID         string          `json:"brandId"`
	SoftwareProduct SoftwareProduct `json:"softwareProduct"`
}

type SoftwareProductDetailsChanged struct {
	SoftwareProductID string `json:"softwareProductId"`
	SoftwareProductDetails
}

func (DataRecipientCreated) isDataRecipientEvent()             {}
func (DataRecipientDetailsChanged) isDataRecipientEvent()      {}
func (DataRecipientBrandAdded) isDataRecipientEvent()          {}
func (DataRecipientBrandDetailsChanged) isDataRecipientEvent() {}
func (DataRecipientBrandActivated) isDataRecipientEvent()      {}
func (DataRecipientBrandDeactivated) isDataRecipientEvent()    {}
func (DataRecipientBrandRemoved) isDataRecipientEvent()        {}
func (SoftwareProductAdded) isDataRecipientEvent()             {}
func (SoftwareProductDetailsChanged) isDataRecipientEvent()    {}

func handleNewDataRecipient(
	state DataRecipientState,
	dr DataRecipient,
	event DataRecipientEvent,
) ([]DataRecipientEvent, error) {
	if state.Imported {
		return nil, fmt.Errorf(
			"%s %w",
			record{kind: "data recipient", id: state.DataRecipient.LegalEntityID},
			ErrExists,
		)
	}

	if err := dr.validate(); err != nil {
		return nil, err
	}

	return []DataRecipientEvent{event}, nil
}

// handleDataRecipientEvent checks the full data recipient after the events, the same as a load.
func handleDataRecipientEvent(state DataRecipientState, events ...DataRecipientEvent) ([]DataRecipientEvent, error) {
	after := state
	for _, e := range events {
		after = ApplyDataRecipient(after, e)
	}

	if err := after.DataRecipient.validate(); err != nil {
		return nil, err
	}

	return events, nil
}

func handleAddDataRecipientBrand(state DataRecipientState, brand DataRecipientBrand) ([]DataRecipientEvent, error) {
	added := DataRecipient{Brands: []DataRecipientBrand{brand}}.IDs()[1:]
	if err := firstIDInUse(state.DataRecipient.IDs(), added); err != nil {
		return nil, err
	}

	return handleDataRecipientEvent(state, DataRecipientBrandAdded{Brand: brand})
}

func handleDataRecipientBrandCommand(
	state DataRecipientState,
	id string,
	handle func(DataRecipientBrand) ([]DataRecipientEvent, error),
) ([]DataRecipientEvent, error) {
	i := slices.IndexFunc(state.DataRecipient.Brands, func(b DataRecipientBrand) bool {
		return b.DataRecipientBrandID == id
	})
	if i < 0 {
		return nil, fmt.Errorf("%s %w", record{kind: "data recipient brand", id: id}, ErrNotFound)
	}

	return handle(state.DataRecipient.Brands[i])
}

// The cascade has the rules of a data recipient status (CDS Status Mapping), applied to the brand.
func handleDataRecipientBrandStatus(state DataRecipientState, id string, to Status) ([]DataRecipientEvent, error) {
	return handleDataRecipientBrandCommand(state, id, func(b DataRecipientBrand) ([]DataRecipientEvent, error) {
		if err := transition(
			record{kind: "data recipient brand", id: id},
			b.Status,
			to,
			statusTransitions,
		); err != nil {
			return nil, err
		}

		var events []DataRecipientEvent

		switch to {
		case Active:
			events = append(events, DataRecipientBrandActivated{BrandID: id})
		case Inactive:
			events = append(events, DataRecipientBrandDeactivated{BrandID: id})
		case Removed:
			events = append(events, DataRecipientBrandRemoved{BrandID: id})
		}

		for _, sp := range b.SoftwareProducts {
			switch {
			case to == Inactive && sp.Status == Active:
				events = append(events, SoftwareProductDeactivated{SoftwareProductID: sp.SoftwareProductID})
			case to == Removed && sp.Status != Removed:
				events = append(events, SoftwareProductRemoved{SoftwareProductID: sp.SoftwareProductID})
			}
		}

		return handleDataRecipientEvent(state, events...)
	})
}

func handleAddSoftwareProduct(state DataRecipientState, c AddSoftwareProduct) ([]DataRecipientEvent, error) {
	return handleDataRecipientBrandCommand(state, c.BrandID, func(DataRecipientBrand) ([]DataRecipientEvent, error) {
		if err := firstIDInUse(state.DataRecipient.IDs(), []string{c.SoftwareProduct.SoftwareProductID}); err != nil {
			return nil, err
		}

		return handleDataRecipientEvent(state, SoftwareProductAdded(c))
	})
}

func handleChangeSoftwareProductDetails(
	state DataRecipientState,
	c ChangeSoftwareProductDetails,
) ([]DataRecipientEvent, error) {
	if _, ok := state.DataRecipient.softwareProduct(c.SoftwareProductID); !ok {
		return nil, fmt.Errorf("%s %w", record{kind: "software product", id: c.SoftwareProductID}, ErrNotFound)
	}

	return handleDataRecipientEvent(
		state,
		SoftwareProductDetailsChanged{SoftwareProductID: c.SoftwareProductID, SoftwareProductDetails: c.Details},
	)
}

func applyDataRecipientChange(dr DataRecipient, event DataRecipientEvent) DataRecipient {
	switch e := event.(type) {
	case DataRecipientDetailsChanged:
		dr.LegalEntityName, dr.AccreditationNumber = e.LegalEntityName, e.AccreditationNumber
		dr.AccreditationLevel, dr.LogoURI = e.AccreditationLevel, e.LogoURI
	case DataRecipientBrandAdded:
		dr.Brands = append(cloneBrands(dr.Brands), e.Brand)
	case DataRecipientBrandDetailsChanged:
		dr = dr.withBrand(e.BrandID, func(b *DataRecipientBrand) { b.BrandName, b.LogoURI = e.BrandName, e.LogoURI })
	case DataRecipientBrandActivated:
		dr = dr.withBrand(e.BrandID, func(b *DataRecipientBrand) { b.Status = Active })
	case DataRecipientBrandDeactivated:
		dr = dr.withBrand(e.BrandID, func(b *DataRecipientBrand) { b.Status = Inactive })
	case DataRecipientBrandRemoved:
		dr = dr.withBrand(e.BrandID, func(b *DataRecipientBrand) { b.Status = Removed })
	case SoftwareProductAdded:
		dr = dr.withBrand(e.BrandID, func(b *DataRecipientBrand) {
			b.SoftwareProducts = append(b.SoftwareProducts, e.SoftwareProduct)
		})
	case SoftwareProductDetailsChanged:
		dr = dr.withSoftwareProduct(e.SoftwareProductID, e.applyTo)
	}

	return dr
}

// withBrand copies the slices, so that apply does not change the state of the caller.
func (dr DataRecipient) withBrand(id string, change func(*DataRecipientBrand)) DataRecipient {
	dr.Brands = cloneBrands(dr.Brands)
	for i := range dr.Brands {
		if dr.Brands[i].DataRecipientBrandID == id {
			change(&dr.Brands[i])
		}
	}

	return dr
}
