-- +goose Up
-- The JWT ID of each revoked access token, until the token expires. The table holds no token.
CREATE TABLE revoked_access_tokens (
    jti text PRIMARY KEY,
    expires_at timestamptz NOT NULL
);

CREATE INDEX revoked_access_tokens_expires_at_idx ON revoked_access_tokens (expires_at);
