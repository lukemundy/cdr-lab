-- +goose Up
-- The defaults are only for rows that exist before this migration: the migrate command then projects each stream again.
ALTER TABLE data_recipients
    ADD COLUMN legal_entity_name text NOT NULL DEFAULT '',
    ADD COLUMN accreditation_number text NOT NULL DEFAULT '',
    ADD COLUMN status text NOT NULL DEFAULT '';
ALTER TABLE data_recipients
    ALTER COLUMN legal_entity_name DROP DEFAULT,
    ALTER COLUMN accreditation_number DROP DEFAULT,
    ALTER COLUMN status DROP DEFAULT;
