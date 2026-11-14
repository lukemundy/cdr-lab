-- +goose Up
-- The time of the last change to the customer. A customer that exists before this migration gets the migration time.
ALTER TABLE customers ADD COLUMN last_updated timestamptz NOT NULL DEFAULT now();
ALTER TABLE customers ALTER COLUMN last_updated DROP DEFAULT;
