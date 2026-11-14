package app

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var ErrProductNotFound = errors.New("the data holder has no product with this ID")

type Effective string

const (
	EffectiveCurrent Effective = "CURRENT"
	EffectiveFuture  Effective = "FUTURE"
	EffectiveAll     Effective = "ALL"
)

func ParseEffective(s string) (Effective, bool) {
	e := Effective(s)
	return e, slices.Contains([]Effective{EffectiveCurrent, EffectiveFuture, EffectiveAll}, e)
}

type ProductQuery struct {
	Effective Effective
	// At is the time that decides which products are current or future.
	At time.Time
	// UpdatedSince is nil for no limit.
	UpdatedSince *time.Time
	// Brand is empty for each brand.
	Brand string
	// Category is empty for each category.
	Category domain.ProductCategory
	// ProductID is empty for each product.
	ProductID string
	// Page starts at 1.
	Page int
	// PageSize is 0 for all products on one page.
	PageSize int
}

type ProductReader interface {
	// Products gives one page, newest lastUpdated first and then in the order of productId, and the total of all pages.
	Products(ctx context.Context, query ProductQuery) (page []domain.Product, total int, err error)
	// Product returns ErrProductNotFound when no product has the ID.
	Product(ctx context.Context, productID string) (domain.Product, error)
}
