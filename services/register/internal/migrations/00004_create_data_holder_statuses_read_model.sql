-- +goose Up
CREATE TABLE data_holder_statuses (
    legal_entity_id text PRIMARY KEY,
    status text NOT NULL,
    industries text[] NOT NULL
);
