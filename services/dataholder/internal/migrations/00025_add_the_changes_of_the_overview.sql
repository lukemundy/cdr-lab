-- +goose Up
-- The overview of the console gives the newest changes, so each change keeps its time.
ALTER TABLE customers ADD COLUMN created_at timestamptz;
UPDATE customers SET created_at = last_updated;
ALTER TABLE customers ALTER COLUMN created_at SET NOT NULL;

CREATE INDEX customers_created_at_idx ON customers (created_at);

ALTER TABLE accounts ADD COLUMN closed_at timestamptz;

CREATE INDEX accounts_created_at_idx ON accounts (created_at);
CREATE INDEX accounts_closed_at_idx ON accounts (closed_at) WHERE closed_at IS NOT NULL;

-- The migrate command projects each stream again after the migrations, which gives these columns their values.
ALTER TABLE arrangements
    ADD COLUMN client_name text NOT NULL DEFAULT '',
    ADD COLUMN sharing_ends_at timestamptz NOT NULL DEFAULT 'epoch',
    ADD COLUMN end_reason text,
    ADD COLUMN withdrawal_method text;
ALTER TABLE arrangements
    ALTER COLUMN client_name DROP DEFAULT,
    ALTER COLUMN sharing_ends_at DROP DEFAULT;

CREATE INDEX arrangements_confirmed_at_idx ON arrangements (confirmed_at);
CREATE INDEX arrangements_ends_at_idx ON arrangements (ends_at);

-- One row for each event of a client registration, with the client_name of the registration at that event.
CREATE TABLE registration_changes (
    software_id text NOT NULL,
    stream_version bigint NOT NULL,
    client_id text NOT NULL,
    client_name text NOT NULL,
    kind text NOT NULL CHECK (kind IN ('CLIENT_REGISTERED', 'CLIENT_UPDATED', 'CLIENT_DELETED')),
    changed_at timestamptz NOT NULL,
    PRIMARY KEY (software_id, stream_version)
);

CREATE INDEX registration_changes_changed_at_idx ON registration_changes (changed_at);
