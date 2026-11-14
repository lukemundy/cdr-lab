-- +goose Up
-- The revocations of the mock CDR intermediate: a record of the CA, not an event of a participant. The serial number
-- is hexadecimal, and the subject is in the RFC 4514 format.
CREATE TABLE cdr_revocations (
    serial text PRIMARY KEY,
    subject text NOT NULL,
    thumbprint text NOT NULL,
    revoked_at timestamptz NOT NULL,
    reason smallint NOT NULL
);

-- The newest CRL of the mock CDR intermediate. The table has one row at most.
CREATE TABLE cdr_crl (
    one_row boolean PRIMARY KEY DEFAULT true CHECK (one_row),
    crl_number bigint NOT NULL,
    this_update timestamptz NOT NULL,
    next_update timestamptz NOT NULL,
    crl bytea NOT NULL
);
