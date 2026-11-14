-- +goose Up
-- Roles are global to the server: another database, or a transaction at the same time, can have created it.
-- +goose StatementBegin
DO $$
BEGIN
  CREATE ROLE dataholder_app NOLOGIN;
EXCEPTION
  WHEN duplicate_object OR unique_violation THEN NULL;
END
$$;
-- +goose StatementEnd

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO dataholder_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO dataholder_app;
-- Only the owner role changes the goose version table.
REVOKE ALL ON goose_db_version, goose_db_version_id_seq FROM dataholder_app;
-- Events are a permanent record: the service only reads and appends them.
REVOKE UPDATE, DELETE ON events FROM dataholder_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO dataholder_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO dataholder_app;

-- The River reindexer runs REINDEX INDEX CONCURRENTLY on the indexes of river_job, which needs MAINTAIN.
GRANT MAINTAIN ON river_job TO dataholder_app;
