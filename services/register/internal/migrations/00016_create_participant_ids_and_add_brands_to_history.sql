-- +goose Up
-- One row for each ID of each participant record, also of a REMOVED record, so that the Register never issues an ID
-- again. The primary key makes each ID unique across all participants. The migrate command projects each stream
-- again, so that the participants that exist get their rows.
CREATE TABLE participant_ids (
    id text PRIMARY KEY,
    stream_type text NOT NULL,
    legal_entity_id text NOT NULL
);
CREATE INDEX participant_ids_stream ON participant_ids (stream_type, legal_entity_id);

ALTER TABLE participant_history
    ADD COLUMN brand_id text,
    ADD COLUMN brand_name text,
    ADD CHECK ((brand_id IS NULL) = (brand_name IS NULL));
