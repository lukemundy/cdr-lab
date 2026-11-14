package domain

import "slices"

type DataRecipientAction string

const (
	Suspend   DataRecipientAction = "suspend"
	Revoke    DataRecipientAction = "revoke"
	Surrender DataRecipientAction = "surrender"
	Reinstate DataRecipientAction = "reinstate"
)

var DataRecipientActions = []DataRecipientAction{Suspend, Revoke, Surrender, Reinstate}

func (a DataRecipientAction) Command() DataRecipientCommand {
	switch a {
	case Suspend:
		return SuspendAccreditation{}
	case Revoke:
		return RevokeAccreditation{}
	case Surrender:
		return SurrenderAccreditation{}
	case Reinstate:
		return ReinstateAccreditation{}
	default:
		return nil
	}
}

// RecordAction is a status action of a brand or a software product: a record whose status is a Status.
type RecordAction string

const (
	Activate   RecordAction = "activate"
	Deactivate RecordAction = "deactivate"
	Remove     RecordAction = "remove"
)

var RecordActions = []RecordAction{Activate, Deactivate, Remove}

func (a RecordAction) SoftwareProductCommand(softwareProductID string) DataRecipientCommand {
	switch a {
	case Activate:
		return ActivateSoftwareProduct{SoftwareProductID: softwareProductID}
	case Deactivate:
		return DeactivateSoftwareProduct{SoftwareProductID: softwareProductID}
	case Remove:
		return RemoveSoftwareProduct{SoftwareProductID: softwareProductID}
	default:
		return nil
	}
}

func (a RecordAction) DataRecipientBrandCommand(brandID string) DataRecipientCommand {
	switch a {
	case Activate:
		return ActivateDataRecipientBrand{BrandID: brandID}
	case Deactivate:
		return DeactivateDataRecipientBrand{BrandID: brandID}
	case Remove:
		return RemoveDataRecipientBrand{BrandID: brandID}
	default:
		return nil
	}
}

// StatusAction gives the status after the action, and each other record that the action changes.
type StatusAction[A ~string, S ~string, C any] struct {
	Action       A
	Status       S
	Irreversible bool
	Changes      []C
}

type SoftwareProductChange struct {
	SoftwareProductID   string
	SoftwareProductName string
	From                Status
	To                  Status
}

type StatusActions struct {
	DataRecipient []StatusAction[DataRecipientAction, DataRecipientStatus, SoftwareProductChange]
	// Brands has an entry for each brand ID, also when no action is permitted.
	Brands map[string][]StatusAction[RecordAction, Status, SoftwareProductChange]
	// SoftwareProducts has an entry for each software product ID, also when no action is permitted.
	SoftwareProducts map[string][]StatusAction[RecordAction, Status, SoftwareProductChange]
}

// PermittedStatusActions runs each status command on the state with the rules of the change, and stores nothing.
func PermittedStatusActions(state DataRecipientState) StatusActions {
	actions := StatusActions{
		Brands:           map[string][]StatusAction[RecordAction, Status, SoftwareProductChange]{},
		SoftwareProducts: map[string][]StatusAction[RecordAction, Status, SoftwareProductChange]{},
	}

	for _, a := range DataRecipientActions {
		after, ok := tryCommand(state, a.Command())
		if !ok {
			continue
		}

		changes := productChanges(state.DataRecipient, after)
		actions.DataRecipient = append(
			actions.DataRecipient,
			StatusAction[DataRecipientAction, DataRecipientStatus, SoftwareProductChange]{
				Action:       a,
				Status:       after.Status,
				Irreversible: isFinal(after.Status, dataRecipientTransitions) || anyFinal(changes),
				Changes:      changes,
			},
		)
	}

	for _, b := range state.DataRecipient.Brands {
		actions.Brands[b.DataRecipientBrandID] = recordActions(state, b.DataRecipientBrandID,
			RecordAction.DataRecipientBrandCommand,
			func(after DataRecipient) Status {
				brand, _ := after.brand(b.DataRecipientBrandID)
				return brand.Status
			})

		for _, sp := range b.SoftwareProducts {
			actions.SoftwareProducts[sp.SoftwareProductID] = recordActions(state, sp.SoftwareProductID,
				RecordAction.SoftwareProductCommand,
				func(after DataRecipient) Status {
					product, _ := after.softwareProduct(sp.SoftwareProductID)
					return product.Status
				})
		}
	}

	return actions
}

func recordActions(
	state DataRecipientState,
	id string,
	command func(RecordAction, string) DataRecipientCommand,
	statusOf func(DataRecipient) Status,
) []StatusAction[RecordAction, Status, SoftwareProductChange] {
	permitted := []StatusAction[RecordAction, Status, SoftwareProductChange]{}

	for _, a := range RecordActions {
		after, ok := tryCommand(state, command(a, id))
		if !ok {
			continue
		}

		status := statusOf(after)
		// The action gives the status of its own record, so a software product is not in its own changes.
		changes := slices.DeleteFunc(productChanges(state.DataRecipient, after), func(c SoftwareProductChange) bool {
			return c.SoftwareProductID == id
		})
		permitted = append(permitted, StatusAction[RecordAction, Status, SoftwareProductChange]{
			Action:       a,
			Status:       status,
			Irreversible: isFinal(status, statusTransitions) || anyFinal(changes),
			Changes:      changes,
		})
	}

	return permitted
}

func tryCommand(state DataRecipientState, command DataRecipientCommand) (DataRecipient, bool) {
	events, err := HandleDataRecipient(state, command)
	if err != nil {
		return DataRecipient{}, false
	}

	for _, e := range events {
		state = ApplyDataRecipient(state, e)
	}

	return state.DataRecipient, true
}

func productChanges(before, after DataRecipient) []SoftwareProductChange {
	changes := []SoftwareProductChange{}

	for _, b := range before.Brands {
		for _, sp := range b.SoftwareProducts {
			next, _ := after.softwareProduct(sp.SoftwareProductID)
			if next.Status == sp.Status {
				continue
			}

			changes = append(changes, SoftwareProductChange{
				SoftwareProductID:   sp.SoftwareProductID,
				SoftwareProductName: sp.SoftwareProductName,
				From:                sp.Status,
				To:                  next.Status,
			})
		}
	}

	return changes
}

func anyFinal(changes []SoftwareProductChange) bool {
	for _, c := range changes {
		if isFinal(c.To, statusTransitions) {
			return true
		}
	}

	return false
}

func isFinal[S comparable](status S, transitions map[S][]S) bool {
	return len(transitions[status]) == 0
}
