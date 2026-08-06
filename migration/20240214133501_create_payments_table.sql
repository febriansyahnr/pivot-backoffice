-- +goose Up
-- +goose StatementBegin
CREATE TABLE `payments` (
     `uuid` varchar(100) NOT NULL,
     `merchant_id` varchar(100) NOT NULL,
     `customer_id` varchar(100) NOT NULL,
     `payment_method_id` varchar(100) NOT NULL,
     `processor_reference_number` varchar(60) NULL,
     `currency` varchar(3) NOT NULL COMMENT "IDR, USD",
     `amount` DECIMAL(18, 2) NOT NULL,
     `fee` DECIMAL(18, 2) NULL,
     `discount` DECIMAL(18, 2) NULL,
     `total_amount` DECIMAL(18, 2) NOT NULL,
     `status` VARCHAR(8) NOT NULL COMMENT "PENDING, SUCCESS, FAILED, ACTIVE, INACTIVE",
     `metadata` JSON NULL,
     `created_at` datetime NOT NULL,
     `updated_at` datetime NOT NULL,
     `deleted_at` datetime NULL,
     PRIMARY KEY (`uuid`),
     KEY `payments_merchant_id_IDX` (`merchant_id`) USING BTREE,
     KEY `payments_customer_id_IDX` (`customer_id`) USING BTREE,
     KEY `payments_payment_method_id_IDX` (`payment_method_id`) USING BTREE,
     KEY `payments_created_at_IDX` (`created_at`) USING BTREE,
     KEY `payments_status_processor_reference_number_IDX` (`status`, `processor_reference_number`) USING BTREE,
     FOREIGN KEY (`merchant_id`) REFERENCES `merchants`(`uuid`),
     FOREIGN KEY (`customer_id`) REFERENCES `customers`(`uuid`),
     FOREIGN KEY (`payment_method_id`) REFERENCES `payment_methods`(`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `payments`;
-- +goose StatementEnd