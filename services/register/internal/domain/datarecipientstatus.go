package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var ErrNotFound = errors.New("does not exist")

// StatusRuleError messages name the rule.
type StatusRuleError struct{ err error }

func (e StatusRuleError) Error() string { return e.err.Error() }

func (e StatusRuleError) Unwrap() error { return e.err }

// Source: the state diagrams in CDS Participant Statuses. The CDS gives no diagram for data holders and brands: they
// use the software product diagram, and REMOVED is final (Register specification, design/participants.md).
var (
	dataRecipientTransitions = map[DataRecipientStatus][]DataRecipientStatus{
		DataRecipientActive:    {DataRecipientSuspended, DataRecipientRevoked, DataRecipientSurrendered},
		DataRecipientSuspended: {DataRecipientActive, DataRecipientRevoked, DataRecipientSurrendered},
		DataRecipientRevoked:   {DataRecipientActive},
	}
	// statusTransitions is for software products and brands.
	statusTransitions = map[Status][]Status{
		Active:   {Inactive, Removed},
		Inactive: {Active, Removed},
	}
	dataHolderTransitions = map[DataHolderStatus][]DataHolderStatus{
		DataHolderActive: {DataHolderRemoved},
	}
)

type SuspendAccreditation struct{}

type RevokeAccreditation struct{}

type SurrenderAccreditation struct{}

type ReinstateAccreditation struct{}

type ActivateSoftwareProduct struct{ SoftwareProductID string }

type DeactivateSoftwareProduct struct{ SoftwareProductID string }

type RemoveSoftwareProduct struct{ SoftwareProductID string }

func (SuspendAccreditation) isDataRecipientCommand()      {}
func (RevokeAccreditation) isDataRecipientCommand()       {}
func (SurrenderAccreditation) isDataRecipientCommand()    {}
func (ReinstateAccreditation) isDataRecipientCommand()    {}
func (ActivateSoftwareProduct) isDataRecipientCommand()   {}
func (DeactivateSoftwareProduct) isDataRecipientCommand() {}
func (RemoveSoftwareProduct) isDataRecipientCommand()     {}

type AccreditationSuspended struct{}

type AccreditationRevoked struct{}

type AccreditationSurrendered struct{}

type AccreditationReinstated struct{}

type SoftwareProductActivated struct {
	SoftwareProductID string `json:"softwareProductId"`
}

type SoftwareProductDeactivated struct {
	SoftwareProductID string `json:"softwareProductId"`
}

type SoftwareProductRemoved struct {
	SoftwareProductID string `json:"softwareProductId"`
}

func (AccreditationSuspended) isDataRecipientEvent()     {}
func (AccreditationRevoked) isDataRecipientEvent()       {}
func (AccreditationSurrendered) isDataRecipientEvent()   {}
func (AccreditationReinstated) isDataRecipientEvent()    {}
func (SoftwareProductActivated) isDataRecipientEvent()   {}
func (SoftwareProductDeactivated) isDataRecipientEvent() {}
func (SoftwareProductRemoved) isDataRecipientEvent()     {}

func handleDataRecipientStatus(
	state DataRecipientState,
	to DataRecipientStatus,
	event DataRecipientEvent,
) ([]DataRecipientEvent, error) {
	dr := state.DataRecipient
	if err := transition(
		record{kind: "data recipient", id: dr.LegalEntityID},
		dr.Status,
		to,
		dataRecipientTransitions,
	); err != nil {
		return nil, err
	}
	// The cascades are events, so that a later change to the cascade rules does not change the history.
	events := []DataRecipientEvent{event}

	for _, b := range dr.Brands {
		for _, sp := range b.SoftwareProducts {
			switch {
			case to == DataRecipientSuspended && sp.Status == Active:
				events = append(events, SoftwareProductDeactivated{SoftwareProductID: sp.SoftwareProductID})
			case (to == DataRecipientRevoked || to == DataRecipientSurrendered) && sp.Status != Removed:
				events = append(events, SoftwareProductRemoved{SoftwareProductID: sp.SoftwareProductID})
			}
		}
	}

	return events, nil
}

func handleSoftwareProductStatus(
	state DataRecipientState,
	id string,
	to Status,
	event DataRecipientEvent,
) ([]DataRecipientEvent, error) {
	sp, ok := state.DataRecipient.softwareProduct(id)
	if !ok {
		return nil, fmt.Errorf("%s %w", record{kind: "software product", id: id}, ErrNotFound)
	}

	if err := transition(
		record{kind: "software product", id: id},
		sp.Status,
		to,
		statusTransitions,
	); err != nil {
		return nil, err
	}

	if err := ApplyDataRecipient(state, event).DataRecipient.validate(); err != nil {
		return nil, StatusRuleError{err}
	}

	return []DataRecipientEvent{event}, nil
}

func transition[S ~string](r record, from, to S, allowed map[S][]S) error {
	next := allowed[from]
	if slices.Contains(next, to) {
		return nil
	}

	if len(next) == 0 {
		return StatusRuleError{fmt.Errorf("%s cannot change from %s to %s: %s is final", r, from, to, from)}
	}

	names := make([]string, len(next))
	for i, s := range next {
		names[i] = string(s)
	}

	list := names[0]
	if len(names) > 1 {
		list = strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
	}

	return StatusRuleError{
		fmt.Errorf("%s cannot change from %s to %s: from %s, the allowed statuses are %s", r, from, to, from, list),
	}
}

func (dr DataRecipient) brand(id string) (DataRecipientBrand, bool) {
	i := slices.IndexFunc(dr.Brands, func(b DataRecipientBrand) bool { return b.DataRecipientBrandID == id })
	if i < 0 {
		return DataRecipientBrand{}, false
	}

	return dr.Brands[i], true
}

func (dr DataRecipient) softwareProduct(id string) (SoftwareProduct, bool) {
	_, sp, ok := dr.brandAndSoftwareProduct(id)

	return sp, ok
}

func (dr DataRecipient) brandAndSoftwareProduct(id string) (DataRecipientBrand, SoftwareProduct, bool) {
	for _, b := range dr.Brands {
		for _, sp := range b.SoftwareProducts {
			if sp.SoftwareProductID == id {
				return b, sp, true
			}
		}
	}

	return DataRecipientBrand{}, SoftwareProduct{}, false
}

func (dr DataRecipient) withSoftwareProductStatus(id string, status Status) DataRecipient {
	return dr.withSoftwareProduct(id, func(sp *SoftwareProduct) { sp.Status = status })
}

func (dr DataRecipient) withSoftwareProduct(id string, change func(*SoftwareProduct)) DataRecipient {
	dr.Brands = cloneBrands(dr.Brands)
	for i := range dr.Brands {
		for j := range dr.Brands[i].SoftwareProducts {
			if sp := &dr.Brands[i].SoftwareProducts[j]; sp.SoftwareProductID == id {
				change(sp)
			}
		}
	}

	return dr
}

// cloneBrands copies the slices, so that apply does not change the state of the caller.
func cloneBrands(brands []DataRecipientBrand) []DataRecipientBrand {
	out := make([]DataRecipientBrand, len(brands))
	for i, b := range brands {
		b.SoftwareProducts = append([]SoftwareProduct(nil), b.SoftwareProducts...)
		out[i] = b
	}

	return out
}
