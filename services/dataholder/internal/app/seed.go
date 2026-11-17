package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var ErrNotEmpty = errors.New("the data holder already has products")

type ProductStore interface {
	// ImportProducts adds all the products, or adds none and returns ErrNotEmpty when the store has products.
	ImportProducts(ctx context.Context, products []domain.Product) error
}

// Seed loads a full set of products of the brand into a store that has none, or makes no changes.
func Seed(ctx context.Context, store ProductStore, details [][]byte) error {
	products, err := CheckProducts(details)
	if err != nil {
		return err
	}

	return store.ImportProducts(ctx, products)
}

// CheckProducts gives the products of a full set of the brand, or an error that names each product and rule that fails.
func CheckProducts(details [][]byte) ([]domain.Product, error) {
	var errs []error

	products := make([]domain.Product, 0, len(details))

	// The IDs come from all the products, so that a product that is not valid can also be a duplicate.
	var ids []string

	for i, detail := range details {
		if id := domain.ProductID(detail); id != "" {
			ids = append(ids, id)
		}

		product, err := domain.ParseProduct(detail)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s:\n%w", productName(i, detail), err))
			continue
		}

		products = append(products, product)
	}

	for _, id := range domain.DuplicateProductIDs(ids) {
		errs = append(errs, fmt.Errorf("productId %q is not unique", id))
	}

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("products are not valid:\n%w", err)
	}

	return products, nil
}

// productName gives the position also, because a product that is not valid can have no ID.
func productName(position int, detail []byte) string {
	if id := domain.ProductID(detail); id != "" {
		return fmt.Sprintf("product %d (%q)", position+1, id)
	}

	return fmt.Sprintf("product %d", position+1)
}
