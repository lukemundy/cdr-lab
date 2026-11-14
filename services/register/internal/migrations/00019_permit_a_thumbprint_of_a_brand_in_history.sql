-- +goose Up
-- A server certificate issue gives a thumbprint with a data holder brand, and no software product.
ALTER TABLE participant_history
    DROP CONSTRAINT participant_history_check1,
    ADD CHECK (thumbprint IS NULL OR software_product_id IS NOT NULL OR brand_id IS NOT NULL);
