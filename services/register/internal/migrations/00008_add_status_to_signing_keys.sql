-- +goose Up
-- The default is only for the keys that exist before this migration, each of which is the one key of its key set.
ALTER TABLE signing_keys ADD COLUMN status text NOT NULL DEFAULT 'signing';
ALTER TABLE signing_keys ALTER COLUMN status DROP DEFAULT;

CREATE UNIQUE INDEX signing_keys_one_signing_idx ON signing_keys (key_set) WHERE status = 'signing';
CREATE UNIQUE INDEX signing_keys_one_next_idx ON signing_keys (key_set) WHERE status = 'next';
