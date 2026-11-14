-- +goose Up
ALTER TABLE arrangements ADD COLUMN software_id text;

UPDATE arrangements SET software_id = state -> 'arrangement' ->> 'softwareId';

ALTER TABLE arrangements ALTER COLUMN software_id SET NOT NULL;

CREATE INDEX arrangements_software_id_idx ON arrangements (software_id, confirmed_at);
