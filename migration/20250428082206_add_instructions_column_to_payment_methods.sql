-- +goose Up
-- Add instructions column to payment_methods table
ALTER TABLE payment_methods ADD COLUMN instructions TEXT DEFAULT NULL AFTER bank_name;

-- +goose Down
ALTER TABLE payment_methods DROP COLUMN instructions;

