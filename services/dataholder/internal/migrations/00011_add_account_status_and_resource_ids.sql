-- +goose Up
-- The defaults only let the migration run on a lab database that has accounts.
ALTER TABLE accounts
    ADD COLUMN credit_limit_cents bigint NOT NULL DEFAULT 0,
    ADD COLUMN open_status text NOT NULL DEFAULT 'OPEN',
    ADD COLUMN created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE accounts
    ALTER COLUMN credit_limit_cents DROP DEFAULT,
    ALTER COLUMN open_status DROP DEFAULT,
    ALTER COLUMN created_at DROP DEFAULT;

-- The accountId that each software product gets for an account, which does not change.
CREATE TABLE account_resource_ids (
    software_id text NOT NULL,
    account_id text NOT NULL,
    resource_id text NOT NULL,
    CONSTRAINT account_resource_ids_pkey PRIMARY KEY (software_id, account_id),
    CONSTRAINT account_resource_ids_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts (account_id),
    CONSTRAINT account_resource_ids_resource_id_key UNIQUE (resource_id)
);
