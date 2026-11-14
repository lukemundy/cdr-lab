-- +goose Up
-- The time at which the arrangement ends or ended: its end, or else the end of its sharing.
ALTER TABLE arrangements ADD COLUMN ends_at timestamptz;

UPDATE arrangements SET ends_at = coalesce(
    (state ->> 'endedAt')::timestamptz,
    confirmed_at + make_interval(secs => (state -> 'arrangement' ->> 'sharingDurationSeconds')::bigint)
);

ALTER TABLE arrangements ALTER COLUMN ends_at SET NOT NULL;
