-- +goose Up
-- +goose StatementBegin
CREATE TABLE rate_limit_configurations (
    `uuid` VARCHAR(36) NOT NULL PRIMARY KEY,
    `merchant_id` VARCHAR(36) NOT NULL,
    `limit` INT NOT NULL,
    `order` INT NOT NULL,
    `time` VARCHAR(20) NOT NULL,
    `variable` VARCHAR(20) NOT NULL,
    `variable_value` VARCHAR(255) NOT NULL,
    `variable_match_type` VARCHAR(20) NOT NULL ,
    `status` VARCHAR(20) NOT NULL,
    `description` TEXT,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rate_limit_configurations;
-- +goose StatementEnd
