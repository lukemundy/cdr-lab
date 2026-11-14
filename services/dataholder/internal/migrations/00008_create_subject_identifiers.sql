-- +goose Up
-- The pairwise sub of each customer for each sector identifier, which does not change.
CREATE TABLE subject_identifiers (
    customer_number text NOT NULL,
    sector_identifier text NOT NULL,
    subject text NOT NULL,
    CONSTRAINT subject_identifiers_pkey PRIMARY KEY (customer_number, sector_identifier),
    CONSTRAINT subject_identifiers_customer_number_fkey FOREIGN KEY (customer_number)
        REFERENCES customers (customer_number),
    CONSTRAINT subject_identifiers_subject_key UNIQUE (subject)
);
