-- +goose Up
-- The default is only for rows that exist before this migration: the migrate command then projects each stream again.
ALTER TABLE data_holders ADD COLUMN version bigint NOT NULL DEFAULT 0;
ALTER TABLE data_holders ALTER COLUMN version DROP DEFAULT;
