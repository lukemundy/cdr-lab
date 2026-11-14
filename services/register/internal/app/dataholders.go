package app

import (
	"context"
	"time"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type DataHolderBrandItem struct {
	Brand domain.DataHolderBrand
	// LastUpdated is the time of the last event of the data holder of the brand.
	LastUpdated time.Time
}

type DataHolderLegalEntity struct {
	LegalEntityID   string                  `json:"legalEntityId"`
	LegalEntityName string                  `json:"legalEntityName"`
	LogoURI         string                  `json:"logoUri"`
	Status          domain.DataHolderStatus `json:"status"`
}

type RegisteredDataHolderBrandItem struct {
	Brand       domain.DataHolderBrand
	LegalEntity DataHolderLegalEntity
	// LastUpdated is the time of the last event of the data holder of the brand.
	LastUpdated time.Time
}

type DataHolderBrandQuery struct {
	// Industry is nil for all industries.
	Industry *domain.Industry
	// UpdatedSince is nil for no limit.
	UpdatedSince *time.Time
	// Page starts at 1.
	Page     int
	PageSize int
}

// DataHolderBrandReader gives the brands in the order of their IDs.
type DataHolderBrandReader interface {
	// DataHolderBrands gives the brands that are not REMOVED with the
	// industry, or with any industry when industry is nil.
	DataHolderBrands(ctx context.Context, industry *domain.Industry) ([]DataHolderBrandItem, error)
}

// RegisteredDataHolderBrandReader gives the brands in the order of their IDs.
type RegisteredDataHolderBrandReader interface {
	// RegisteredDataHolderBrands gives one page of the brands that have a dataHolderBrandId,
	// in each status, and the number of brands in all pages.
	RegisteredDataHolderBrands(
		ctx context.Context,
		query DataHolderBrandQuery,
	) (page []RegisteredDataHolderBrandItem, total int, err error)
}

type DataHolderItem struct {
	DataHolder domain.DataHolder
	// LastUpdated is the time of the last event of the data holder, which includes changes to its brands.
	LastUpdated time.Time
	// Version is the stream version of the data holder.
	Version int64
}

type DataHolderPageQuery struct {
	// Status is nil for each status.
	Status *domain.DataHolderStatus
	// NameContains matches a part of the legal entity name or of a brand name, with no difference for case.
	// It is empty for each name.
	NameContains string
	Sort         ParticipantSort
	Descending   bool
	// Page starts at 1.
	Page     int
	PageSize int
}

// DataHolderPageReader sorts by the legal entity ID after the sort of the query, so that each page is stable.
type DataHolderPageReader interface {
	DataHolderPage(ctx context.Context, query DataHolderPageQuery) (page []DataHolderItem, total int, err error)
}

// DataHolderDetailReader gives an error that wraps domain.ErrNotFound when no data holder has the legal entity ID.
type DataHolderDetailReader interface {
	DataHolder(ctx context.Context, legalEntityID string) (DataHolderItem, error)
}
