package app

import (
	"context"
	"time"
)

// DisclosureStore keeps one record for each response that disclosed CDR data, with the scope of the API as the data
// cluster.
type DisclosureStore interface {
	DisclosureReader
	AddDisclosure(ctx context.Context, arrangementID, scope string, at time.Time) error
}

type DisclosureReader interface {
	// FirstAndLastDisclosures has no entry for a pair with no disclosure.
	FirstAndLastDisclosures(ctx context.Context, arrangementIDs []string) (map[DisclosureKey]DisclosurePeriod, error)
}

type DisclosureKey struct {
	ArrangementID string
	Scope         string
}

type DisclosurePeriod struct {
	First time.Time
	Last  time.Time
}

type Disclosures struct {
	Store DisclosureStore
	Now   func() time.Time
}

func (d Disclosures) RecordDisclosure(ctx context.Context, arrangementID, scope string) error {
	return d.Store.AddDisclosure(ctx, arrangementID, scope, d.Now())
}
