package domain

import (
	"errors"
	"fmt"
)

type Participants struct {
	DataHolders    []DataHolder    `json:"dataHolders"`
	DataRecipients []DataRecipient `json:"dataRecipients"`
}

// Validate applies the rules between aggregates. The import of each aggregate applies the rules in it.
func (p Participants) Validate() error {
	var errs []error

	owners := map[string]record{}

	check := func(owner record, ids []string) {
		seen := map[string]bool{}
		for _, id := range ids {
			if seen[id] {
				continue
			}

			seen[id] = true
			if first, ok := owners[id]; ok {
				errs = append(errs, fmt.Errorf("ID %q is not unique: it is in %s and %s", id, first, owner))
				continue
			}

			owners[id] = owner
		}
	}
	for _, dh := range p.DataHolders {
		ids := []string{dh.LegalEntityID}
		for _, b := range dh.Brands {
			for _, id := range []string{b.DataHolderBrandID, b.InterimID} {
				if id != "" {
					ids = append(ids, id)
				}
			}
		}

		check(record{kind: "data holder", id: dh.LegalEntityID}, ids)
	}

	for _, dr := range p.DataRecipients {
		ids := []string{dr.LegalEntityID}
		for _, b := range dr.Brands {
			ids = append(ids, b.DataRecipientBrandID)
			for _, sp := range b.SoftwareProducts {
				ids = append(ids, sp.SoftwareProductID)
			}
		}

		check(record{kind: "data recipient", id: dr.LegalEntityID}, ids)
	}

	return errors.Join(errs...)
}
