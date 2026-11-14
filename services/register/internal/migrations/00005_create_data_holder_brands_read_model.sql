-- +goose Up
CREATE TABLE data_holder_brands (
    brand_id text PRIMARY KEY,
    legal_entity_id text NOT NULL,
    status text NOT NULL,
    industries text[] NOT NULL,
    brand jsonb NOT NULL,
    last_updated timestamptz NOT NULL
);

CREATE INDEX data_holder_brands_legal_entity_id_idx ON data_holder_brands (legal_entity_id);
