-- +goose Up
CREATE TABLE registrations (
    client_id text PRIMARY KEY,
    software_id text NOT NULL UNIQUE,
    -- The domain registration. No query reads its content: move a field to a column before a query needs it.
    registration jsonb NOT NULL
);
