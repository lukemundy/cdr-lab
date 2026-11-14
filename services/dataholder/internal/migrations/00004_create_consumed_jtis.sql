-- +goose Up
CREATE TABLE consumed_jtis (
    jti text PRIMARY KEY,
    consumed_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX consumed_jtis_consumed_at_idx ON consumed_jtis (consumed_at);
