package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

// HistoryEntry has only the fields that the table of the history in services/register/design/operator-console.md
// gives, so that no payload gets to a reader.
type HistoryEntry struct {
	StreamVersion int64
	EventType     string
	RecordedAt    time.Time
	// SoftwareProduct is nil when the event does not change a software product.
	SoftwareProduct *HistorySoftwareProduct
	// Thumbprint is empty when the event does not issue, register or remove a certificate.
	Thumbprint string
	// Brand is nil when the event does not change a brand.
	Brand *HistoryBrand
}

type HistoryBrand struct {
	// ID is the ID of domain.DataHolderBrand.ID for a data holder brand.
	ID string
	// Name is the name at the time of the event.
	Name string
}

type HistorySoftwareProduct struct {
	ID string
	// Name is the name at the time of the event.
	Name string
}

type HistoryPageQuery struct {
	LegalEntityID string
	// Page starts at 1.
	Page     int
	PageSize int
}

// HistoryReader gives the entries of one stream, newest first. It gives an error that wraps domain.ErrNotFound when
// the stream has no events.
type HistoryReader interface {
	History(ctx context.Context, streamType string, query HistoryPageQuery) (page []HistoryEntry, total int, err error)
}

// HistoryOfDataRecipient gives an error for an event type that it does not know, so that a projection cannot skip an
// event.
func HistoryOfDataRecipient(streamID string, recorded []eventsourcing.RecordedEvent) ([]HistoryEntry, error) {
	t := DataRecipientType
	state := t.Initial
	entries := make([]HistoryEntry, 0, len(recorded))

	for _, r := range recorded {
		event, err := t.Codec.Decode(r.EventData)
		if err != nil {
			return nil, fmt.Errorf("%s %s version %d: %w", t.StreamType, streamID, r.StreamVersion, err)
		}

		state = t.Apply(state, event)
		entry := historyEntry(r)

		switch e := event.(type) {
		case domain.DataRecipientImported, domain.DataRecipientCreated, domain.DataRecipientDetailsChanged,
			domain.AccreditationSuspended, domain.AccreditationRevoked, domain.AccreditationSurrendered,
			domain.AccreditationReinstated:
		case domain.DataRecipientBrandAdded:
			entry.Brand, err = historyDataRecipientBrand(state, e.Brand.DataRecipientBrandID)
		case domain.DataRecipientBrandDetailsChanged:
			entry.Brand, err = historyDataRecipientBrand(state, e.BrandID)
		case domain.DataRecipientBrandActivated:
			entry.Brand, err = historyDataRecipientBrand(state, e.BrandID)
		case domain.DataRecipientBrandDeactivated:
			entry.Brand, err = historyDataRecipientBrand(state, e.BrandID)
		case domain.DataRecipientBrandRemoved:
			entry.Brand, err = historyDataRecipientBrand(state, e.BrandID)
		case domain.SoftwareProductAdded:
			entry.SoftwareProduct, err = historySoftwareProduct(state, e.SoftwareProduct.SoftwareProductID)
		case domain.SoftwareProductDetailsChanged:
			entry.SoftwareProduct, err = historySoftwareProduct(state, e.SoftwareProductID)
		case domain.SoftwareProductActivated:
			entry.SoftwareProduct, err = historySoftwareProduct(state, e.SoftwareProductID)
		case domain.SoftwareProductDeactivated:
			entry.SoftwareProduct, err = historySoftwareProduct(state, e.SoftwareProductID)
		case domain.SoftwareProductRemoved:
			entry.SoftwareProduct, err = historySoftwareProduct(state, e.SoftwareProductID)
		case domain.ClientCertificateRegistered:
			entry.SoftwareProduct, err = historySoftwareProduct(state, e.SoftwareProductID)
			entry.Thumbprint = e.Thumbprint
		case domain.ClientCertificateRemoved:
			entry.SoftwareProduct, err = historySoftwareProduct(state, e.SoftwareProductID)
			entry.Thumbprint = e.Thumbprint
		case domain.ClientCertificateIssued:
			entry.SoftwareProduct, err = historySoftwareProduct(state, e.SoftwareProductID)
			entry.Thumbprint = e.Thumbprint
		default:
			err = fmt.Errorf("event type %q has no history entry", r.EventType)
		}

		if err != nil {
			return nil, fmt.Errorf("%s %s version %d: %w", t.StreamType, streamID, r.StreamVersion, err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// HistoryOfDataHolder gives an error for an event type that it does not know, so that a projection cannot skip an
// event.
func HistoryOfDataHolder(streamID string, recorded []eventsourcing.RecordedEvent) ([]HistoryEntry, error) {
	t := DataHolderType
	state := t.Initial
	entries := make([]HistoryEntry, 0, len(recorded))

	for _, r := range recorded {
		event, err := t.Codec.Decode(r.EventData)
		if err != nil {
			return nil, fmt.Errorf("%s %s version %d: %w", t.StreamType, streamID, r.StreamVersion, err)
		}

		state = t.Apply(state, event)
		entry := historyEntry(r)

		switch e := event.(type) {
		case domain.DataHolderImported, domain.DataHolderCreated, domain.DataHolderDetailsChanged,
			domain.DataHolderLegalEntityRemoved:
		case domain.DataHolderBrandAdded:
			entry.Brand, err = historyDataHolderBrand(state, e.Brand.ID())
		case domain.DataHolderBrandDetailsChanged:
			entry.Brand, err = historyDataHolderBrand(state, e.BrandID)
		case domain.DataHolderBrandActivated:
			entry.Brand, err = historyDataHolderBrand(state, e.BrandID)
		case domain.DataHolderBrandDeactivated:
			entry.Brand, err = historyDataHolderBrand(state, e.BrandID)
		case domain.DataHolderBrandRemoved:
			entry.Brand, err = historyDataHolderBrand(state, e.BrandID)
		case domain.ServerCertificateIssued:
			entry.Brand, err = historyDataHolderBrand(state, e.BrandID)
			entry.Thumbprint = e.Thumbprint
		default:
			err = fmt.Errorf("event type %q has no history entry", r.EventType)
		}

		if err != nil {
			return nil, fmt.Errorf("%s %s version %d: %w", t.StreamType, streamID, r.StreamVersion, err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func historyDataHolderBrand(state domain.DataHolderState, id string) (*HistoryBrand, error) {
	for _, b := range state.DataHolder.Brands {
		if b.ID() == id {
			return &HistoryBrand{ID: id, Name: b.BrandName}, nil
		}
	}

	return nil, fmt.Errorf("data holder brand %w", domain.ErrNotFound)
}

func historyEntry(r eventsourcing.RecordedEvent) HistoryEntry {
	return HistoryEntry{StreamVersion: r.StreamVersion, EventType: r.EventType, RecordedAt: r.RecordedAt}
}

func historyDataRecipientBrand(state domain.DataRecipientState, id string) (*HistoryBrand, error) {
	for _, b := range state.DataRecipient.Brands {
		if b.DataRecipientBrandID == id {
			return &HistoryBrand{ID: id, Name: b.BrandName}, nil
		}
	}

	return nil, fmt.Errorf("data recipient brand %w", domain.ErrNotFound)
}

func historySoftwareProduct(state domain.DataRecipientState, id string) (*HistorySoftwareProduct, error) {
	for _, b := range state.DataRecipient.Brands {
		for _, sp := range b.SoftwareProducts {
			if sp.SoftwareProductID == id {
				return &HistorySoftwareProduct{ID: id, Name: sp.SoftwareProductName}, nil
			}
		}
	}

	return nil, fmt.Errorf("software product %w", domain.ErrNotFound)
}
