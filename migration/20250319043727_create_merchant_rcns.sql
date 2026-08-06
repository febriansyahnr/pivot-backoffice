-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `merchant_rcns`
(
    `uuid`                VARCHAR(36)  NOT NULL PRIMARY KEY,
    `merchant_id`         VARCHAR(36)  NOT NULL,
    `principal_issuer`    VARCHAR(255) NOT NULL,
    `real_card_number`    VARCHAR(255) NOT NULL,
    `encrypt_kms_version` VARCHAR(255) NOT NULL,
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp NULL DEFAULT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `merchant_rcns`;
-- +goose StatementEnd