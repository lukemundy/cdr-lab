-- +goose Up
CREATE TABLE arrangements (
    arrangement_id text PRIMARY KEY,
    customer_number text NOT NULL,
    confirmed_at timestamptz NOT NULL,
    -- The arrangement and its end. No query reads its content: move a field to a column before a query needs it.
    state jsonb NOT NULL
);

CREATE INDEX arrangements_customer_number_idx ON arrangements (customer_number, confirmed_at);
