-- +goose Up
CREATE TABLE customers (
    customer_number text PRIMARY KEY,
    first_name text NOT NULL,
    last_name text NOT NULL,
    email_address text NOT NULL
);

CREATE TABLE accounts (
    -- Gives the sequence of creation. It is not an identifier of the account for a client.
    account_sequence bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_number text NOT NULL,
    display_name text NOT NULL,
    product_category text NOT NULL,
    product_name text NOT NULL,
    bsb text NOT NULL,
    account_number text NOT NULL,
    CONSTRAINT accounts_customer_number_fkey FOREIGN KEY (customer_number) REFERENCES customers (customer_number),
    CONSTRAINT accounts_bsb_account_number_key UNIQUE (bsb, account_number)
);

CREATE INDEX accounts_customer_number_idx ON accounts (customer_number);
