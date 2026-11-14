-- +goose Up
-- The identifier of an account in the data holder only. It is a random UUID, and a data recipient does not get it.
ALTER TABLE accounts ADD COLUMN account_id text NOT NULL DEFAULT gen_random_uuid()::text;
ALTER TABLE accounts ALTER COLUMN account_id DROP DEFAULT;
ALTER TABLE accounts ADD CONSTRAINT accounts_account_id_key UNIQUE (account_id);
