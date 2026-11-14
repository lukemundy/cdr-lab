-- +goose Up
-- The Register adds the row of a key set at its next start, also for a key set that exists before this migration.
CREATE TABLE key_sets (
    key_set text PRIMARY KEY,
    version bigint NOT NULL
);
