-- +goose Up
-- +goose StatementBegin
ALTER TABLE `payment_methods`
    ADD COLUMN `activation_method` VARCHAR(12) NOT NULL DEFAULT 'MANUAL' AFTER `bank_name`,
    ADD COLUMN `required_document` JSON NULL AFTER `activation_method`,
    ADD COLUMN `country_of_operation` VARCHAR(2) NOT NULL DEFAULT 'ID' AFTER `required_document`,
    ADD COLUMN `supported_currency` VARCHAR(3) NOT NULL DEFAULT 'IDR' AFTER `country_of_operation`,
    ADD COLUMN `processor` VARCHAR(40) NOT NULL DEFAULT '' AFTER `supported_currency`,
    ADD COLUMN `config` JSON NULL AFTER `processor`;

ALTER TABLE `payment_method_merchant`
    ADD COLUMN `activation_status` VARCHAR(16) NOT NULL DEFAULT '' AFTER `is_active`,
    ADD COLUMN `channel_type` VARCHAR(20) NOT NULL DEFAULT 'AGGREGATOR' AFTER `activation_status`,
    ADD COLUMN `config` JSON NULL AFTER `channel_type`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `payment_method_merchant`
    DROP COLUMN `activation_status`,
    DROP COLUMN `channel_type`,
    DROP COLUMN `config`;

ALTER TABLE `payment_methods`
    DROP COLUMN `activation_method`,
    DROP COLUMN `required_document`,
    DROP COLUMN `country_of_operation`,
    DROP COLUMN `supported_currency`,
    DROP COLUMN `processor`,
    DROP COLUMN `config`;
-- +goose StatementEnd