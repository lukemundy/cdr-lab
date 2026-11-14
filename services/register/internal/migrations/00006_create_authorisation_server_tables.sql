-- +goose Up
CREATE TABLE software_product_clients (
    software_product_id text PRIMARY KEY,
    legal_entity_id text NOT NULL REFERENCES data_recipient_statuses ON DELETE CASCADE,
    status text NOT NULL,
    jwks_uri text NOT NULL,
    client_certificates text[] NOT NULL
);

CREATE INDEX software_product_clients_legal_entity_id_idx ON software_product_clients (legal_entity_id);

CREATE TABLE consumed_jtis (
    jti text PRIMARY KEY,
    consumed_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX consumed_jtis_consumed_at_idx ON consumed_jtis (consumed_at);

CREATE TABLE signing_keys (
    key_id text PRIMARY KEY,
    key_set text NOT NULL,
    -- PKCS #8, DER.
    private_key bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX signing_keys_key_set_idx ON signing_keys (key_set, created_at);
