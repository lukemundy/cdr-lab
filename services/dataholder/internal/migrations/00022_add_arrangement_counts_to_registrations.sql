-- +goose Up
ALTER TABLE arrangements ADD COLUMN client_id text;

UPDATE arrangements SET client_id = state -> 'arrangement' ->> 'clientId';

ALTER TABLE arrangements ALTER COLUMN client_id SET NOT NULL;

CREATE INDEX arrangements_client_id_idx ON arrangements (client_id, ends_at);

-- The migrate subcommand projects all streams after the migrations, which sets the count of each registration.
ALTER TABLE registrations ADD COLUMN arrangement_count integer NOT NULL DEFAULT 0;

ALTER TABLE registrations ALTER COLUMN arrangement_count DROP DEFAULT;
