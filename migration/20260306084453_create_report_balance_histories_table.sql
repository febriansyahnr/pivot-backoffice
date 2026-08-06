-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `report_balance_histories` (
  `transaction_id` CHAR(36) NOT NULL PRIMARY KEY,
  `merchant_id` CHAR(36) NOT NULL,
  `type` VARCHAR(50) NOT NULL,
  `balance_type` VARCHAR(100) NOT NULL,
  `transaction_type` VARCHAR(255) NOT NULL,
  `channel` VARCHAR(100) NOT NULL,
  `reference_id` VARCHAR(255) NOT NULL,
  `currency` CHAR(3) NOT NULL,
  `amount` DECIMAL(19,4) NOT NULL,
  `fee` DECIMAL(19,4) NOT NULL,
  `remarks` TEXT NOT NULL,
  `status` VARCHAR(100) NOT NULL,
  `reason_type` VARCHAR(100) NULL,
  `reason_description` VARCHAR(255) NULL,
  `settlement_model` VARCHAR(50) NOT NULL,
  `settlement_status` VARCHAR(50) NOT NULL,
  `settlement_at` TIMESTAMP(6) NOT NULL,
  `additional_info` JSON NULL,
  `created_at` TIMESTAMP NOT NULL,
  `status_updated_at` TIMESTAMP(6) NOT NULL,
  `source_id` CHAR(36) NOT NULL,
  `source_account_id` CHAR(36) NOT NULL,
  `source_created_at` TIMESTAMP NULL DEFAULT NULL,
  `source_created_by` VARCHAR(200) NOT NULL,
  `_created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `_updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `_is_deleted` TINYINT NOT NULL DEFAULT '0',
  `_deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `_ingested_at` TIMESTAMP NULL DEFAULT NULL,
  INDEX `merchant_status_updated_at_comp_idx` (`merchant_id`, `status_updated_at`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `report_balance_histories`;
-- +goose StatementEnd
