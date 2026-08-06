-- +goose Up
-- +goose StatementBegin

ALTER TABLE `account_transactions`
ADD COLUMN `settlement_at` TIMESTAMP NULL AFTER `additional_info`,
ADD COLUMN `settlement_status` VARCHAR(20) NULL AFTER `settlement_at`;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE `account_transactions`
DROP COLUMN `settlement_at`,
DROP COLUMN `settlement_status`;

-- +goose StatementEnd