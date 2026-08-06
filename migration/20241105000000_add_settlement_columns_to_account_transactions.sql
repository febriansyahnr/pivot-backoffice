-- +goose Up
-- +goose StatementBegin

ALTER TABLE `account_transactions`
ADD COLUMN `settlement_status` VARCHAR(50) NULL AFTER `additional_info`,
ADD COLUMN `settlement_at` TIMESTAMP NULL AFTER `settlement_status`;

-- Drop the columns
ALTER TABLE `account_transactions`
DROP COLUMN `settlement_at`,
DROP COLUMN `settlement_status`;

-- +goose StatementEnd
