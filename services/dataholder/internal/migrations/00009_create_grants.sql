-- +goose Up
-- The authorisation code and the tokens of each consent. The code and the refresh token are secrets.
CREATE TABLE grants (
    grant_id text PRIMARY KEY,
    client_id text NOT NULL,
    subject text NOT NULL,
    scopes text[] NOT NULL,
    arrangement_id text NOT NULL,
    sharing_duration_seconds bigint NOT NULL,
    authenticated_at timestamptz NOT NULL,
    -- Also the expiry of the refresh token.
    sharing_ends_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    revoked_at timestamptz,
    code text NOT NULL,
    code_expires_at timestamptz NOT NULL,
    code_consumed_at timestamptz,
    redirect_uri text NOT NULL,
    nonce text NOT NULL,
    code_challenge text NOT NULL,
    -- NULL for a consent with no sharing duration.
    refresh_token text,
    -- Empty until the client exchanges the code.
    certificate_thumbprint text NOT NULL,
    CONSTRAINT grants_code_key UNIQUE (code),
    CONSTRAINT grants_refresh_token_key UNIQUE (refresh_token)
);
