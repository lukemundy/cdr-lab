-- +goose Up
CREATE TABLE products (
    product_id text PRIMARY KEY,
    product_category text NOT NULL,
    brand text NOT NULL,
    last_updated timestamptz NOT NULL,
    effective_from timestamptz,
    effective_to timestamptz,
    -- The domain product. No query reads its content: move a field to a column or a table before a query needs it.
    product jsonb NOT NULL
);
