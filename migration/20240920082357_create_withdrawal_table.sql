-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `withdrawals` (
  `id` VARCHAR(36) NOT NULL PRIMARY KEY,
  `merchant_id` VARCHAR(36) NOT NULL,
  `beneficiary_bank_code` VARCHAR(20) NOT NULL,
  `beneficiary_bank_name` VARCHAR(60) NOT NULL,
  `beneficiary_account_no` VARCHAR(60) NOT NULL,
  `beneficiary_account_name` VARCHAR(60) NOT NULL,
  `currency` VARCHAR(3) NOT NULL COMMENT 'IDR, USD, Or Etc',
  `amount` DECIMAL(65,2) NOT NULL,
  `metadata` JSON DEFAULT NULL,
  `created_by` VARCHAR(36) NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  KEY `withdrawals_merchant_created_at_idx`(`merchant_id`, `created_at`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `withdrawals`;
-- +goose StatementEnd
