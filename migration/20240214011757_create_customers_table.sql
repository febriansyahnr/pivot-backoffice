-- +goose Up
-- +goose StatementBegin
CREATE TABLE `customers` (
    `uuid` varchar(100) NOT NULL,
    `merchant_id` varchar(100) NOT NULL,
    `full_name` varchar(255) NOT NULL,
    `email` varchar(60) NULL,
    `phone_number` VARCHAR(20) NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime NULL,
    PRIMARY KEY (`uuid`),
    KEY `customers_merchant_id_IDX` (`merchant_id`) USING BTREE,
    FOREIGN KEY (`merchant_id`) REFERENCES `merchants`(`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `customers`;
-- +goose StatementEnd
