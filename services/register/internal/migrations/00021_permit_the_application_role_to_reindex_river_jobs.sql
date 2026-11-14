-- +goose Up
-- The River reindexer runs REINDEX INDEX CONCURRENTLY on the indexes of river_job, which needs MAINTAIN.
GRANT MAINTAIN ON river_job TO register_app;
