package app

import (
	"context"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type DataHolderStatusItem struct {
	LegalEntityID string
	Status        domain.DataHolderStatus
}

type DataRecipientStatusItem struct {
	LegalEntityID string
	Status        domain.DataRecipientStatus
}

type SoftwareProductStatusItem struct {
	SoftwareProductID string
	Status            domain.Status
}

// StatusReader gives the current status of each participant, in the order of their IDs.
type StatusReader interface {
	// DataHolderStatuses gives the data holders with a brand in the industry, or all data holders when industry is nil.
	DataHolderStatuses(ctx context.Context, industry *domain.Industry) ([]DataHolderStatusItem, error)
	DataRecipientStatuses(ctx context.Context) ([]DataRecipientStatusItem, error)
	SoftwareProductStatuses(ctx context.Context) ([]SoftwareProductStatusItem, error)
}
