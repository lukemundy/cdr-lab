-- +goose Up
-- One row for each response that disclosed CDR data of an arrangement (CDR Rules, rule 9.3(1)(d)).
CREATE TABLE disclosures (
    -- Gives the sequence of the records. It is not an identifier for a client.
    disclosure_sequence bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    arrangement_id text NOT NULL,
    -- The scope of the API, which selects the data cluster of the arrangement.
    scope text NOT NULL,
    disclosed_at timestamptz NOT NULL
);

CREATE INDEX disclosures_arrangement_id_scope_idx ON disclosures (arrangement_id, scope, disclosed_at);

-- The data holder keeps each record: the service only reads and adds them.
REVOKE UPDATE, DELETE ON disclosures FROM dataholder_app;
