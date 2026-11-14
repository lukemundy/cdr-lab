-- +goose Up
CREATE TABLE signing_keys (
    key_id text PRIMARY KEY,
    -- PKCS #8, DER.
    private_key bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
