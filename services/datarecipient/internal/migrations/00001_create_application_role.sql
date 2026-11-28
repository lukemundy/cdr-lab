-- +goose Up
-- Roles are global to the server: another database, or a transaction at the same time, can have created it.
-- +goose StatementBegin
DO $$
BEGIN
  CREATE ROLE datarecipient_app NOLOGIN;
EXCEPTION
  WHEN duplicate_object OR unique_violation THEN NULL;
END
$$;
-- +goose StatementEnd

-- The database has no table of the service yet. The default privileges give the role each table of a later migration.
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO datarecipient_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO datarecipient_app;
