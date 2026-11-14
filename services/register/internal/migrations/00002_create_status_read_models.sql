-- +goose Up
CREATE TABLE data_recipient_statuses (
    legal_entity_id text PRIMARY KEY,
    status text NOT NULL
);

CREATE TABLE software_product_statuses (
    software_product_id text PRIMARY KEY,
    legal_entity_id text NOT NULL REFERENCES data_recipient_statuses ON DELETE CASCADE,
    status text NOT NULL
);

CREATE INDEX software_product_statuses_legal_entity_id_idx ON software_product_statuses (legal_entity_id);
