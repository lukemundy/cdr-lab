// Package productstore keeps the products of the data holder in PostgreSQL.
package productstore

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

type Store struct {
	pool *pgxpool.Pool
}

var (
	_ app.ProductStore  = (*Store)(nil)
	_ app.ProductReader = (*Store)(nil)
)

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) ImportProducts(ctx context.Context, products []domain.Product) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		// The lock makes a second import wait, so that it sees the products of the first.
		if _, err := tx.Exec(ctx, "LOCK TABLE products IN EXCLUSIVE MODE"); err != nil {
			return err
		}

		var empty bool
		if err := tx.QueryRow(ctx, "SELECT NOT EXISTS (SELECT 1 FROM products)").Scan(&empty); err != nil {
			return err
		}

		if !empty {
			return app.ErrNotEmpty
		}

		rows := make([][]any, 0, len(products))
		for _, p := range products {
			encoded, err := json.Marshal(p)
			if err != nil {
				return err
			}

			rows = append(
				rows,
				[]any{p.ProductID, p.ProductCategory, p.Brand, p.LastUpdated, p.EffectiveFrom, p.EffectiveTo, encoded},
			)
		}

		_, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"products"},
			[]string{
				"product_id", "product_category", "brand", "last_updated", "effective_from", "effective_to", "product",
			},
			pgx.CopyFromRows(rows),
		)

		return err
	})
}

// The effective period includes effective_from and excludes effective_to.
const matchingProducts = `FROM products
	WHERE ($1::timestamptz IS NULL OR last_updated > $1)
	AND ($2 = '' OR brand = $2)
	AND ($3 = '' OR product_category = $3)
	AND ($6 = '' OR product_id = $6)
	AND CASE $4::text
		WHEN 'CURRENT' THEN (effective_from IS NULL OR effective_from <= $5::timestamptz)
			AND (effective_to IS NULL OR effective_to > $5)
		WHEN 'FUTURE' THEN effective_from > $5
		ELSE true
	END`

func (s *Store) Products(ctx context.Context, query app.ProductQuery) ([]domain.Product, int, error) {
	filters := []any{
		query.UpdatedSince, query.Brand, string(query.Category), string(query.Effective), query.At, query.ProductID,
	}

	// LIMIT NULL is no limit.
	var limit *int
	if query.PageSize > 0 {
		limit = &query.PageSize
	}

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) "+matchingProducts, filters...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(
		ctx,
		"SELECT product "+matchingProducts+" ORDER BY last_updated DESC, product_id LIMIT $7 OFFSET $8",
		append(filters, limit, (query.Page-1)*query.PageSize)...,
	)
	if err != nil {
		return nil, 0, err
	}

	page, err := pgx.CollectRows(rows, pgx.RowTo[domain.Product])

	return page, total, err
}

func (s *Store) Product(ctx context.Context, productID string) (domain.Product, error) {
	rows, err := s.pool.Query(ctx, "SELECT product FROM products WHERE product_id = $1", productID)
	if err != nil {
		return domain.Product{}, err
	}

	product, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[domain.Product])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Product{}, app.ErrProductNotFound
	}

	return product, err
}
