-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `balances` (
  `uuid` char(36) NOT NULL,
  `merchant_id` char(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `amount` decimal(19,4) NOT NULL DEFAULT '0',
  `currency` varchar(3) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `account_transaction`;
-- +goose StatementEnd
