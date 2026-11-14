package domain

import (
	"fmt"
	"slices"
)

type CreateDataHolder struct{ DataHolder DataHolder }

// DataHolderDetails has the fields of a data holder legal entity that a change replaces.
type DataHolderDetails struct {
	LegalEntityName string `json:"legalEntityName"`
	LogoURI         string `json:"logoUri"`
}

type ChangeDataHolderDetails struct{ Details DataHolderDetails }

type RemoveDataHolder struct{}

type AddDataHolderBrand struct{ Brand DataHolderBrand }

// DataHolderBrandDetails has the fields of a data holder brand that a change replaces.
type DataHolderBrandDetails struct {
	BrandName      string         `json:"brandName"`
	BrandGroup     string         `json:"brandGroup,omitempty"`
	Industries     []Industry     `json:"industries"`
	LogoURI        string         `json:"logoUri"`
	EndpointDetail EndpointDetail `json:"endpointDetail"`
	JWKSEndpoint   string         `json:"jwksEndpoint"`
}

// ChangeDataHolderBrandDetails and the other brand commands have a BrandID that is the dataHolderBrandId or the
// interimId of the brand.
type ChangeDataHolderBrandDetails struct {
	BrandID string
	Details DataHolderBrandDetails
}

type ActivateDataHolderBrand struct{ BrandID string }

type DeactivateDataHolderBrand struct{ BrandID string }

type RemoveDataHolderBrand struct{ BrandID string }

func (CreateDataHolder) isDataHolderCommand()             {}
func (ChangeDataHolderDetails) isDataHolderCommand()      {}
func (RemoveDataHolder) isDataHolderCommand()             {}
func (AddDataHolderBrand) isDataHolderCommand()           {}
func (ChangeDataHolderBrandDetails) isDataHolderCommand() {}
func (ActivateDataHolderBrand) isDataHolderCommand()      {}
func (DeactivateDataHolderBrand) isDataHolderCommand()    {}
func (RemoveDataHolderBrand) isDataHolderCommand()        {}

type DataHolderCreated struct {
	DataHolder DataHolder `json:"dataHolder"`
}

type DataHolderDetailsChanged struct {
	DataHolderDetails
}

type DataHolderLegalEntityRemoved struct{}

type DataHolderBrandAdded struct {
	Brand DataHolderBrand `json:"brand"`
}

// DataHolderBrandDetailsChanged and the other brand events have a BrandID that is the value of DataHolderBrand.ID.
type DataHolderBrandDetailsChanged struct {
	BrandID string `json:"brandId"`
	DataHolderBrandDetails
}

type DataHolderBrandActivated struct {
	BrandID string `json:"brandId"`
}

type DataHolderBrandDeactivated struct {
	BrandID string `json:"brandId"`
}

type DataHolderBrandRemoved struct {
	BrandID string `json:"brandId"`
}

func (DataHolderCreated) isDataHolderEvent()             {}
func (DataHolderDetailsChanged) isDataHolderEvent()      {}
func (DataHolderLegalEntityRemoved) isDataHolderEvent()  {}
func (DataHolderBrandAdded) isDataHolderEvent()          {}
func (DataHolderBrandDetailsChanged) isDataHolderEvent() {}
func (DataHolderBrandActivated) isDataHolderEvent()      {}
func (DataHolderBrandDeactivated) isDataHolderEvent()    {}
func (DataHolderBrandRemoved) isDataHolderEvent()        {}

func handleNewDataHolder(state DataHolderState, dh DataHolder, event DataHolderEvent) ([]DataHolderEvent, error) {
	if state.Imported {
		return nil, fmt.Errorf("%s %w", record{kind: "data holder", id: state.DataHolder.LegalEntityID}, ErrExists)
	}

	if err := dh.validate(); err != nil {
		return nil, err
	}

	return []DataHolderEvent{event}, nil
}

// handleDataHolderChange checks the full data holder after the event, the same as a load.
func handleDataHolderChange(state DataHolderState, event DataHolderEvent) ([]DataHolderEvent, error) {
	if err := ApplyDataHolder(state, event).DataHolder.validate(); err != nil {
		return nil, err
	}

	return []DataHolderEvent{event}, nil
}

func handleAddDataHolderBrand(state DataHolderState, brand DataHolderBrand) ([]DataHolderEvent, error) {
	if err := firstIDInUse(state.DataHolder.IDs(), DataHolder{Brands: []DataHolderBrand{brand}}.IDs()); err != nil {
		return nil, err
	}

	return handleDataHolderChange(state, DataHolderBrandAdded{Brand: brand})
}

func handleRemoveDataHolder(state DataHolderState) ([]DataHolderEvent, error) {
	dh := state.DataHolder
	if err := transition(
		record{kind: "data holder", id: dh.LegalEntityID},
		dh.Status,
		DataHolderRemoved,
		dataHolderTransitions,
	); err != nil {
		return nil, err
	}

	// The cascade is events, so that a later change to the cascade rules does not change the history.
	events := []DataHolderEvent{DataHolderLegalEntityRemoved{}}

	for _, b := range dh.Brands {
		if b.Status != Removed {
			events = append(events, DataHolderBrandRemoved{BrandID: b.ID()})
		}
	}

	return events, nil
}

// handleDataHolderBrandCommand gives handle the ID of DataHolderBrand.ID, also when the command has the interimId of a
// brand that has a dataHolderBrandId.
func handleDataHolderBrandCommand(
	state DataHolderState,
	id string,
	handle func(id string) ([]DataHolderEvent, error),
) ([]DataHolderEvent, error) {
	b, ok := state.DataHolder.brand(id)
	if !ok {
		return nil, fmt.Errorf("%s %w", record{kind: "data holder brand", id: id}, ErrNotFound)
	}

	return handle(b.ID())
}

func handleDataHolderBrandStatus(
	state DataHolderState,
	id string,
	to Status,
	event func(id string) DataHolderEvent,
) ([]DataHolderEvent, error) {
	return handleDataHolderBrandCommand(state, id, func(id string) ([]DataHolderEvent, error) {
		b, _ := state.DataHolder.brand(id)
		if err := transition(record{kind: "data holder brand", id: id}, b.Status, to, statusTransitions); err != nil {
			return nil, err
		}

		return handleDataHolderChange(state, event(id))
	})
}

func (dh DataHolder) brand(id string) (DataHolderBrand, bool) {
	i := slices.IndexFunc(dh.Brands, func(b DataHolderBrand) bool {
		return id != "" && (b.DataHolderBrandID == id || b.InterimID == id)
	})
	if i < 0 {
		return DataHolderBrand{}, false
	}

	return dh.Brands[i], true
}

func (dh DataHolder) IDs() []string {
	ids := []string{dh.LegalEntityID}
	for _, b := range dh.Brands {
		ids = append(ids, b.DataHolderBrandID, b.InterimID)
	}

	return slices.DeleteFunc(ids, func(id string) bool { return id == "" })
}

func (dh DataHolder) withBrand(id string, change func(*DataHolderBrand)) DataHolder {
	dh.Brands = slices.Clone(dh.Brands)
	for i := range dh.Brands {
		if dh.Brands[i].ID() == id {
			change(&dh.Brands[i])
		}
	}

	return dh
}

func (dh DataHolder) withBrandStatus(id string, status Status) DataHolder {
	return dh.withBrand(id, func(b *DataHolderBrand) { b.Status = status })
}
