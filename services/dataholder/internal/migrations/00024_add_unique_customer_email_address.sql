-- +goose Up
-- The email address is a user identifier, so it identifies one customer, with no difference for upper and lower case.
CREATE UNIQUE INDEX customers_email_address_key ON customers (lower(email_address));
