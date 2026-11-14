-- +goose Up
-- One row for each event of a participant, with only the fields that the console history gives. No payload.
CREATE TABLE participant_history (
    stream_type text NOT NULL,
    legal_entity_id text NOT NULL,
    stream_version bigint NOT NULL,
    event_type text NOT NULL,
    recorded_at timestamptz NOT NULL,
    software_product_id text,
    software_product_name text,
    thumbprint text,
    PRIMARY KEY (stream_type, legal_entity_id, stream_version),
    CHECK ((software_product_id IS NULL) = (software_product_name IS NULL)),
    CHECK (thumbprint IS NULL OR software_product_id IS NOT NULL)
);
