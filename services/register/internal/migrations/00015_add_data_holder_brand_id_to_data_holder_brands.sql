-- +goose Up
-- NULL for a brand that shares only product data, because its interim ID is not a dataHolderBrandId.
-- The migrate command then projects each stream again, so that the rows that exist get a value.
ALTER TABLE data_holder_brands ADD COLUMN data_holder_brand_id text;
