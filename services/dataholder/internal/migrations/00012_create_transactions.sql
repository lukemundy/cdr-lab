-- +goose Up
CREATE TABLE transactions (
    -- Gives the sequence of creation. It is not an identifier of the transaction for a client.
    transaction_sequence bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    transaction_id text NOT NULL,
    account_id text NOT NULL,
    type text NOT NULL,
    status text NOT NULL,
    description text NOT NULL,
    amount_cents bigint NOT NULL,
    reference text NOT NULL,
    posting_date_time timestamptz,
    execution_date_time timestamptz,
    value_date_time timestamptz,
    merchant_name text NOT NULL,
    merchant_category_code text NOT NULL,
    biller_code text NOT NULL,
    biller_name text NOT NULL,
    crn text NOT NULL,
    apca_number text NOT NULL,
    extended_data jsonb,
    effective_time timestamptz NOT NULL GENERATED ALWAYS AS (COALESCE(posting_date_time, execution_date_time)) STORED,
    CONSTRAINT transactions_transaction_id_key UNIQUE (transaction_id),
    CONSTRAINT transactions_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts (account_id)
);

CREATE INDEX transactions_account_id_effective_time_idx
    ON transactions (account_id, effective_time DESC, transaction_sequence DESC);
