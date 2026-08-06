-- +goose Up
-- +goose StatementBegin
CREATE TABLE `payment_methods` (
     `uuid` varchar(100) NOT NULL,
     `type` varchar(20) NOT NULL COMMENT "VIRTUAL_ACCOUNT, BANK_TRANSFER, QRIS, CREDIT_CARD",
     `name` varchar(60) NOT NULL COMMENT "e.g. VA Permata, VA Mandiri etc",
     `description` varchar(255) NULL,
     `logo` varchar(255) NULL,
     `acquirer` varchar(20) NOT NULL,
     `bank_name` varchar(60) NULL,
     `created_at` datetime NOT NULL,
     `updated_at` datetime NOT NULL,
     `deleted_at` datetime NULL,
     PRIMARY KEY (`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `payment_methods`;
-- +goose StatementEnd

-- logo
-- GCS Path URL
-- string
-- type
-- Virtual account, Bank transfer etc
-- timestamp
-- created_at
-- timestamp
-- updated_at
-- timestamp
-- deleted_at
