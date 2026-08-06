-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS merchant_forbidden_usecases (
    `uuid` VARCHAR(36) NOT NULL PRIMARY KEY,
    `merchant_id` VARCHAR(36) NOT NULL,
    `use_case` VARCHAR(20) NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime NULL,

    KEY `merchant_forbidden_usecases_merchant_id_IDX` (`merchant_id`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS merchant_forbidden_usecases;
-- +goose StatementEnd