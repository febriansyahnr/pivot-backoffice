-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `fraud_rules` (
    `uuid` VARCHAR(36)  NOT NULL,
    `rule_name` VARCHAR(255) NOT NULL,
    `condition` VARCHAR(255) DEFAULT 'default' NOT NULL,
    `priority` INT DEFAULT 0 NOT NULL,
    `weight` DOUBLE DEFAULT 0.0 NOT NULL,
    `is_active` BOOL DEFAULT 1 NOT NULL,
    `reference_type` VARCHAR(100) NOT NULL,
    `provider` VARCHAR(255) NULL DEFAULT NULL,
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`uuid`),
    CONSTRAINT fraud_rules_unique UNIQUE KEY (`priority`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `fraud_rules`;
-- +goose StatementEnd
