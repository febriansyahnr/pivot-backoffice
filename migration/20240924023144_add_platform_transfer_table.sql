-- +goose Up
-- +goose StatementBegin

CREATE TABLE `transfers` (
  `uuid` varchar(36) NOT NULL,
  `reference_id` varchar(100) NOT NULL,
  `merchant_id` varchar(36) NOT NULL,
  `recipient_id` varchar(36) NOT NULL,
  `currency` varchar(3) NOT NULL COMMENT 'IDR, USD',
  `amount` decimal(18,2) NOT NULL,
  `remarks` varchar(100) NOT NULL DEFAULT '',
  `reason_description` varchar(100) NOT NULL DEFAULT '',
  `status` varchar(20) NOT NULL COMMENT 'PENDING, SUCCESS, FAILED',
  `transfer_type` varchar(20) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  KEY `payments_merchant_id_composite_IDX` (`merchant_id`, `uuid`, `reference_id`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `transfers`;
-- +goose StatementEnd
