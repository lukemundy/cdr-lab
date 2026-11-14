-- +goose Up
-- The default is only for rows that exist before this migration, because the projection writes the legal entity of each new row.
ALTER TABLE data_holder_brands ADD COLUMN legal_entity jsonb NOT NULL DEFAULT '{}';
ALTER TABLE data_holder_brands ALTER COLUMN legal_entity DROP DEFAULT;
