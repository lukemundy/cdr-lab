-- +goose Up
ALTER TABLE registrations ADD COLUMN registered_at timestamptz, ADD COLUMN updated_at timestamptz;

UPDATE registrations r SET
    registered_at = (
        SELECT e.recorded_at FROM events e
        WHERE e.stream_type = 'client-registration' AND e.stream_id = r.software_id
            AND e.event_type = 'ClientRegistered'
        ORDER BY e.stream_version DESC LIMIT 1
    ),
    updated_at = (
        SELECT e.recorded_at FROM events e
        WHERE e.stream_type = 'client-registration' AND e.stream_id = r.software_id
        ORDER BY e.stream_version DESC LIMIT 1
    );

ALTER TABLE registrations ALTER COLUMN registered_at SET NOT NULL, ALTER COLUMN updated_at SET NOT NULL;
