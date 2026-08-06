-- +goose Up
-- +goose StatementBegin
CREATE TABLE `payment_items` (
    `uuid` varchar(100) NOT NULL,
    `payment_id` varchar(100) NOT NULL,
    `name` varchar(255) NOT NULL,
    `description` varchar(255) NULL,
    `qty` int NOT NULL,
    `currency` varchar(3) NOT NULL COMMENT "IDR, USD",
    `amount` DECIMAL(18, 2) NOT NULL,
    `total_amount` DECIMAL(18, 2) NOT NULL,
    `metadata` JSON NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime NULL,
    PRIMARY KEY (`uuid`),
    KEY `payments_items_payment_id_IDX` (`payment_id`) USING BTREE,
    FOREIGN KEY (`payment_id`) REFERENCES `payments`(`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `payment_items`;
-- +goose StatementEnd