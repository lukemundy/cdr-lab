-- +goose Up
CREATE TABLE software_product_statuses (
    software_product_id text PRIMARY KEY,
    status text NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE', 'REMOVED'))
);

-- One row: the last response of Get Software Products Statuses that had the statuses.
CREATE TABLE software_product_statuses_response (
    only_row boolean PRIMARY KEY DEFAULT true CHECK (only_row),
    etag text NOT NULL,
    saved_at timestamptz NOT NULL
);
