package domain

import "slices"

// ParticipantLoad gives the IDs of the records of a load file in the sequence of the file.
type ParticipantLoad struct {
	Added    []string
	Existing []string
}

func (l *ParticipantLoad) Append(other ParticipantLoad) {
	l.Added = append(l.Added, other.Added...)
	l.Existing = append(l.Existing, other.Existing...)
}

// DataHolderLoad gives the commands that add the records of dh that the data holder does not have. A brand exists when
// the data holder has a brand with its ID.
func DataHolderLoad(state DataHolderState, dh DataHolder) ([]DataHolderCommand, ParticipantLoad) {
	var load ParticipantLoad

	if !state.Imported {
		load.Added = append(load.Added, dh.LegalEntityID)
		for _, b := range dh.Brands {
			load.Added = append(load.Added, b.ID())
		}

		return []DataHolderCommand{CreateDataHolder{DataHolder: dh}}, load
	}

	var commands []DataHolderCommand

	load.Existing = append(load.Existing, dh.LegalEntityID)

	for _, b := range dh.Brands {
		if _, ok := state.DataHolder.brand(b.ID()); ok {
			load.Existing = append(load.Existing, b.ID())
			continue
		}

		load.Added = append(load.Added, b.ID())
		commands = append(commands, AddDataHolderBrand{Brand: b})
	}

	return commands, load
}

// DataRecipientLoad gives the commands that add the records of dr that the data recipient does not have. A software
// product exists when a brand of the data recipient has it, also a different brand than in dr.
func DataRecipientLoad(state DataRecipientState, dr DataRecipient) ([]DataRecipientCommand, ParticipantLoad) {
	var load ParticipantLoad

	if !state.Imported {
		load.Added = dr.IDs()
		return []DataRecipientCommand{CreateDataRecipient{DataRecipient: dr}}, load
	}

	var commands []DataRecipientCommand

	load.Existing = append(load.Existing, dr.LegalEntityID)
	registered := state.DataRecipient

	for _, b := range dr.Brands {
		brandExists := slices.ContainsFunc(registered.Brands, func(e DataRecipientBrand) bool {
			return e.DataRecipientBrandID == b.DataRecipientBrandID
		})

		var products []SoftwareProduct

		if brandExists {
			load.Existing = append(load.Existing, b.DataRecipientBrandID)
		} else {
			load.Added = append(load.Added, b.DataRecipientBrandID)
		}

		for _, sp := range b.SoftwareProducts {
			if _, ok := registered.softwareProduct(sp.SoftwareProductID); ok {
				load.Existing = append(load.Existing, sp.SoftwareProductID)
				continue
			}

			load.Added = append(load.Added, sp.SoftwareProductID)
			products = append(products, sp)
		}

		if !brandExists {
			b.SoftwareProducts = products
			commands = append(commands, AddDataRecipientBrand{Brand: b})

			continue
		}

		for _, sp := range products {
			commands = append(commands, AddSoftwareProduct{BrandID: b.DataRecipientBrandID, SoftwareProduct: sp})
		}
	}

	return commands, load
}
