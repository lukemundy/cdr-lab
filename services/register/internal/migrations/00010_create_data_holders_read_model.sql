-- +goose Up
-- One row for each data holder legal entity, also for a data holder with no brands.
CREATE TABLE data_holders (
    legal_entity_id text PRIMARY KEY,
    legal_entity_name text NOT NULL,
    status text NOT NULL,
    brand_names text[] NOT NULL,
    data_holder jsonb NOT NULL,
    last_updated timestamptz NOT NULL
);
