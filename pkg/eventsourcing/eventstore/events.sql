CREATE TABLE events (
    -- Not in commit order: a transaction that starts earlier can commit later.
    global_position bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    stream_type text NOT NULL,
    stream_id text NOT NULL,
    stream_version bigint NOT NULL CHECK (stream_version > 0),
    event_type text NOT NULL,
    schema_version integer NOT NULL CHECK (schema_version > 0),
    payload jsonb NOT NULL,
    metadata jsonb NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT events_stream_version_key UNIQUE (stream_type, stream_id, stream_version)
);
