-- +goose Up
CREATE TABLE data_recipients (
    legal_entity_id text PRIMARY KEY,
    data_recipient jsonb NOT NULL,
    last_updated timestamptz NOT NULL
);
