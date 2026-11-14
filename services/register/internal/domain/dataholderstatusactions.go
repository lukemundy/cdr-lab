package domain

type DataHolderAction string

const RemoveLegalEntity DataHolderAction = "remove"

var DataHolderActions = []DataHolderAction{RemoveLegalEntity}

func (a DataHolderAction) Command() DataHolderCommand {
	switch a {
	case RemoveLegalEntity:
		return RemoveDataHolder{}
	default:
		return nil
	}
}

func (a RecordAction) DataHolderBrandCommand(brandID string) DataHolderCommand {
	switch a {
	case Activate:
		return ActivateDataHolderBrand{BrandID: brandID}
	case Deactivate:
		return DeactivateDataHolderBrand{BrandID: brandID}
	case Remove:
		return RemoveDataHolderBrand{BrandID: brandID}
	default:
		return nil
	}
}

type DataHolderBrandChange struct {
	DataHolderBrandID string
	InterimID         string
	BrandName         string
	From              Status
	To                Status
}

type DataHolderStatusActions struct {
	DataHolder []StatusAction[DataHolderAction, DataHolderStatus, DataHolderBrandChange]
	// Brands has an entry for each DataHolderBrand.ID, also when no action is permitted.
	Brands map[string][]StatusAction[RecordAction, Status, DataHolderBrandChange]
}

// PermittedDataHolderStatusActions runs each status command on the state with the rules of the change, and stores
// nothing.
func PermittedDataHolderStatusActions(state DataHolderState) DataHolderStatusActions {
	actions := DataHolderStatusActions{
		DataHolder: []StatusAction[DataHolderAction, DataHolderStatus, DataHolderBrandChange]{},
		Brands:     map[string][]StatusAction[RecordAction, Status, DataHolderBrandChange]{},
	}

	for _, a := range DataHolderActions {
		after, ok := tryDataHolderCommand(state, a.Command())
		if !ok {
			continue
		}

		changes := brandChanges(state.DataHolder, after, "")
		actions.DataHolder = append(
			actions.DataHolder,
			StatusAction[DataHolderAction, DataHolderStatus, DataHolderBrandChange]{
				Action:       a,
				Status:       after.Status,
				Irreversible: isFinal(after.Status, dataHolderTransitions) || anyFinalBrand(changes),
				Changes:      changes,
			},
		)
	}

	for _, b := range state.DataHolder.Brands {
		id := b.ID()
		permitted := []StatusAction[RecordAction, Status, DataHolderBrandChange]{}

		for _, a := range RecordActions {
			after, ok := tryDataHolderCommand(state, a.DataHolderBrandCommand(id))
			if !ok {
				continue
			}

			brand, _ := after.brand(id)
			// The action gives the status of its own brand, so the brand is not in its own changes.
			changes := brandChanges(state.DataHolder, after, id)
			permitted = append(permitted, StatusAction[RecordAction, Status, DataHolderBrandChange]{
				Action:       a,
				Status:       brand.Status,
				Irreversible: isFinal(brand.Status, statusTransitions) || anyFinalBrand(changes),
				Changes:      changes,
			})
		}

		actions.Brands[id] = permitted
	}

	return actions
}

func tryDataHolderCommand(state DataHolderState, command DataHolderCommand) (DataHolder, bool) {
	events, err := HandleDataHolder(state, command)
	if err != nil {
		return DataHolder{}, false
	}

	for _, e := range events {
		state = ApplyDataHolder(state, e)
	}

	return state.DataHolder, true
}

func brandChanges(before, after DataHolder, exceptID string) []DataHolderBrandChange {
	changes := []DataHolderBrandChange{}

	for _, b := range before.Brands {
		next, _ := after.brand(b.ID())
		if b.ID() == exceptID || next.Status == b.Status {
			continue
		}

		changes = append(changes, DataHolderBrandChange{
			DataHolderBrandID: b.DataHolderBrandID,
			InterimID:         b.InterimID,
			BrandName:         b.BrandName,
			From:              b.Status,
			To:                next.Status,
		})
	}

	return changes
}

func anyFinalBrand(changes []DataHolderBrandChange) bool {
	for _, c := range changes {
		if isFinal(c.To, statusTransitions) {
			return true
		}
	}

	return false
}
