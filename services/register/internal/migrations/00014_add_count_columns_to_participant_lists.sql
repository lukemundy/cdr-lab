-- +goose Up
-- The defaults are only for rows that exist before this migration: the migrate command then projects each stream again.
ALTER TABLE data_recipients
    ADD COLUMN brand_count integer NOT NULL DEFAULT 0,
    ADD COLUMN software_product_count integer NOT NULL DEFAULT 0;
ALTER TABLE data_recipients
    ALTER COLUMN brand_count DROP DEFAULT,
    ALTER COLUMN software_product_count DROP DEFAULT;
ALTER TABLE data_holders ADD COLUMN brand_count integer NOT NULL DEFAULT 0;
ALTER TABLE data_holders ALTER COLUMN brand_count DROP DEFAULT;
