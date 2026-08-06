-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS `creditcards`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE `creditcards` (
  `uuid` varchar(36) NOT NULL,
  `reference_id` varchar(100) NOT NULL,
  `processor_reference_number` varchar(100) NOT NULL,
  `bank_merchant_id` varchar(50) DEFAULT NULL,
  `amount` decimal(18,2) NOT NULL,
  `fee` decimal(18,2) NOT NULL,
  `total_amount` decimal(18,2) NOT NULL,
  `currency` char(3) NOT NULL,
  `authentication_method` varchar(13) NOT NULL DEFAULT 'CHALLENGE',
  `status` varchar(50) NOT NULL DEFAULT 'WAITING_FOR_PAYMENT',
  `payment_url` varchar(255) NOT NULL DEFAULT '',
  `authentication_result` json DEFAULT NULL,
  `card_data` json DEFAULT NULL,
  `merchant_id` varchar(36) NOT NULL DEFAULT '',
  `expired_at` timestamp NOT NULL DEFAULT ((now() + interval 15 minute)),
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`uuid`),
  KEY `idx_creditcards_reference_id` (`reference_id`),
  KEY `idx_creditcards_processor_reference_number` (`processor_reference_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
-- +goose StatementEnd
