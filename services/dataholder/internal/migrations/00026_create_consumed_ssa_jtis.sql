-- +goose Up
-- The JWT IDs of SSAs are separate from those of other JWTs, because a different issuer gives them. Each stays until
-- its SSA expires.
CREATE TABLE consumed_ssa_jtis (
    jti text PRIMARY KEY,
    expires_at timestamptz NOT NULL
);

CREATE INDEX consumed_ssa_jtis_expires_at_idx ON consumed_ssa_jtis (expires_at);
