package app

import (
	"context"
	"time"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type DataRecipientItem struct {
	DataRecipient domain.DataRecipient
	// LastUpdated is the time of the last event of the data recipient, which
	// includes changes to its brands and software products.
	LastUpdated time.Time
	// Version is the stream version of the data recipient.
	Version int64
}

// DataRecipientReader gives each data recipient, in the order of their legal entity IDs.
type DataRecipientReader interface {
	DataRecipients(ctx context.Context) ([]DataRecipientItem, error)
}

type DataRecipientPageQuery struct {
	// Status is nil for each status.
	Status *domain.DataRecipientStatus
	// NameContains matches a part of the name with no difference for case. It is empty for each name.
	NameContains string
	Sort         ParticipantSort
	Descending   bool
	// Page starts at 1.
	Page     int
	PageSize int
}

// DataRecipientPageReader sorts by the legal entity ID after the sort of the query, so that each page is stable.
type DataRecipientPageReader interface {
	DataRecipientPage(
		ctx context.Context,
		query DataRecipientPageQuery,
	) (page []DataRecipientItem, total int, err error)
}

// DataRecipientDetailReader gives an error that wraps domain.ErrNotFound when no data recipient has the legal entity
// ID.
type DataRecipientDetailReader interface {
	DataRecipient(ctx context.Context, legalEntityID string) (DataRecipientItem, error)
}
