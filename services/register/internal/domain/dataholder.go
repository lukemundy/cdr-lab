package domain

import (
	"errors"
	"fmt"
	"slices"
)

type DataHolderStatus string

const (
	DataHolderActive  DataHolderStatus = "ACTIVE"
	DataHolderRemoved DataHolderStatus = "REMOVED"
)

type DataHolder struct {
	LegalEntityID   string            `json:"legalEntityId"`
	LegalEntityName string            `json:"legalEntityName"`
	LogoURI         string            `json:"logoUri"`
	Status          DataHolderStatus  `json:"status"`
	Brands          []DataHolderBrand `json:"dataHolderBrands"`
}

type DataHolderBrand struct {
	DataHolderBrandID string         `json:"dataHolderBrandId,omitempty"`
	InterimID         string         `json:"interimId,omitempty"`
	BrandName         string         `json:"brandName"`
	BrandGroup        string         `json:"brandGroup,omitempty"`
	Industries        []Industry     `json:"industries"`
	LogoURI           string         `json:"logoUri"`
	Status            Status         `json:"status"`
	EndpointDetail    EndpointDetail `json:"endpointDetail"`
	JWKSEndpoint      string         `json:"jwksEndpoint"`
}

// ID is the brand ID, or the interim ID of a brand that shares only product data.
func (b DataHolderBrand) ID() string {
	if b.DataHolderBrandID != "" {
		return b.DataHolderBrandID
	}

	return b.InterimID
}

type EndpointDetail struct {
	Version          string `json:"version"`
	PublicBaseURI    string `json:"publicBaseUri"`
	ProductBaseURI   string `json:"productBaseUri,omitempty"`
	ResourceBaseURI  string `json:"resourceBaseUri"`
	InfosecBaseURI   string `json:"infosecBaseUri"`
	ExtensionBaseURI string `json:"extensionBaseUri,omitempty"`
	WebsiteURI       string `json:"websiteUri"`
}

type DataHolderState struct {
	Imported   bool
	DataHolder DataHolder
}

type DataHolderCommand interface{ isDataHolderCommand() }

type ImportDataHolder struct{ DataHolder DataHolder }

func (ImportDataHolder) isDataHolderCommand() {}

type DataHolderEvent interface{ isDataHolderEvent() }

type DataHolderImported struct {
	DataHolder DataHolder `json:"dataHolder"`
}

func (DataHolderImported) isDataHolderEvent() {}

func HandleDataHolder(state DataHolderState, command DataHolderCommand) ([]DataHolderEvent, error) {
	switch c := command.(type) {
	case ImportDataHolder:
		return handleNewDataHolder(state, c.DataHolder, DataHolderImported(c))
	case CreateDataHolder:
		return handleNewDataHolder(state, c.DataHolder, DataHolderCreated(c))
	}

	if !state.Imported {
		return nil, fmt.Errorf("data holder %w", ErrNotFound)
	}

	switch c := command.(type) {
	case ChangeDataHolderDetails:
		return handleDataHolderChange(state, DataHolderDetailsChanged{DataHolderDetails: c.Details})
	case RemoveDataHolder:
		return handleRemoveDataHolder(state)
	case AddDataHolderBrand:
		return handleAddDataHolderBrand(state, c.Brand)
	case ChangeDataHolderBrandDetails:
		return handleDataHolderBrandCommand(state, c.BrandID, func(id string) ([]DataHolderEvent, error) {
			return handleDataHolderChange(
				state,
				DataHolderBrandDetailsChanged{BrandID: id, DataHolderBrandDetails: c.Details},
			)
		})
	case ActivateDataHolderBrand:
		return handleDataHolderBrandStatus(state, c.BrandID, Active, func(id string) DataHolderEvent {
			return DataHolderBrandActivated{BrandID: id}
		})
	case DeactivateDataHolderBrand:
		return handleDataHolderBrandStatus(state, c.BrandID, Inactive, func(id string) DataHolderEvent {
			return DataHolderBrandDeactivated{BrandID: id}
		})
	case RemoveDataHolderBrand:
		return handleDataHolderBrandStatus(state, c.BrandID, Removed, func(id string) DataHolderEvent {
			return DataHolderBrandRemoved{BrandID: id}
		})
	case IssueServerCertificate:
		return handleIssueServerCertificate(state, c)
	default:
		return nil, errors.New("unknown data holder command")
	}
}

func ApplyDataHolder(state DataHolderState, event DataHolderEvent) DataHolderState {
	switch e := event.(type) {
	case DataHolderImported:
		return DataHolderState{Imported: true, DataHolder: e.DataHolder}
	case DataHolderCreated:
		return DataHolderState{Imported: true, DataHolder: e.DataHolder}
	case DataHolderDetailsChanged:
		state.DataHolder.LegalEntityName = e.LegalEntityName
		state.DataHolder.LogoURI = e.LogoURI
	case DataHolderLegalEntityRemoved:
		state.DataHolder.Status = DataHolderRemoved
	case DataHolderBrandAdded:
		state.DataHolder.Brands = append(slices.Clone(state.DataHolder.Brands), e.Brand)
	case DataHolderBrandDetailsChanged:
		state.DataHolder = state.DataHolder.withBrand(e.BrandID, func(b *DataHolderBrand) {
			b.BrandName, b.BrandGroup, b.Industries = e.BrandName, e.BrandGroup, e.Industries
			b.LogoURI, b.EndpointDetail, b.JWKSEndpoint = e.LogoURI, e.EndpointDetail, e.JWKSEndpoint
		})
	case DataHolderBrandActivated:
		state.DataHolder = state.DataHolder.withBrandStatus(e.BrandID, Active)
	case DataHolderBrandDeactivated:
		state.DataHolder = state.DataHolder.withBrandStatus(e.BrandID, Inactive)
	case DataHolderBrandRemoved:
		state.DataHolder = state.DataHolder.withBrandStatus(e.BrandID, Removed)
	}

	return state
}

// Industries includes the industries of brands in all statuses.
func (dh DataHolder) Industries() []Industry {
	var got []Industry

	for _, industry := range industries {
		if slices.ContainsFunc(
			dh.Brands,
			func(b DataHolderBrand) bool { return slices.Contains(b.Industries, industry) },
		) {
			got = append(got, industry)
		}
	}

	return got
}

func (dh DataHolder) validate() error {
	var p problems

	holder := record{kind: "data holder", id: dh.LegalEntityID}
	p.id(holder, "legalEntityId", dh.LegalEntityID)
	p.requiredText(holder, "legalEntityName", dh.LegalEntityName, MaxDataHolderLegalEntityNameLength)
	p.requiredURI(holder, "logoUri", dh.LogoURI)

	statusValid := dh.Status == DataHolderActive || dh.Status == DataHolderRemoved
	if !statusValid {
		p.add(holder, "status", "status %q is not ACTIVE or REMOVED", dh.Status)
	}

	for _, b := range dh.Brands {
		brand := record{kind: "data holder brand", id: b.ID()}
		if b.ID() == "" {
			p.add(brand, "dataHolderBrandId", "dataHolderBrandId or interimId is required")
		}

		if b.DataHolderBrandID != "" {
			p.id(brand, "dataHolderBrandId", b.DataHolderBrandID)
		}

		if b.InterimID != "" {
			p.id(brand, "interimId", b.InterimID)
		}

		p.required(brand, "brandName", b.BrandName)
		p.requiredURI(brand, "logoUri", b.LogoURI)
		p.requiredURI(brand, "endpointDetail.publicBaseUri", b.EndpointDetail.PublicBaseURI)
		p.optionalURI(brand, "endpointDetail.productBaseUri", b.EndpointDetail.ProductBaseURI)
		p.optionalURI(brand, "endpointDetail.extensionBaseUri", b.EndpointDetail.ExtensionBaseURI)

		// A brand with no dataHolderBrandId needs no secure endpoints, but a value that it has must be valid.
		secureURI := p.optionalURI

		if b.DataHolderBrandID != "" {
			p.required(brand, "endpointDetail.version", b.EndpointDetail.Version)

			secureURI = p.requiredURI
		}

		secureURI(brand, "endpointDetail.resourceBaseUri", b.EndpointDetail.ResourceBaseURI)
		secureURI(brand, "endpointDetail.infosecBaseUri", b.EndpointDetail.InfosecBaseURI)
		secureURI(brand, "endpointDetail.websiteUri", b.EndpointDetail.WebsiteURI)
		secureURI(brand, "jwksEndpoint", b.JWKSEndpoint)

		if len(b.Industries) == 0 {
			p.add(brand, "industries", "industries is required")
		}

		for i, industry := range b.Industries {
			switch {
			case !slices.Contains(industries, industry):
				p.add(brand, "industries", "industry %q is not banking, energy, non-bank-lending or telco", industry)
			case slices.Contains(b.Industries[:i], industry):
				p.add(brand, "industries", "industry %q is not unique", industry)
			}
		}

		if p.status(brand, b.Status) && dh.Status == DataHolderRemoved && b.Status != Removed {
			p.statusRuleBroken(
				brand,
				"is %s, but %s is REMOVED: a brand of a REMOVED data holder must be REMOVED",
				b.Status,
				holder,
			)
		}
	}

	return p.err()
}
