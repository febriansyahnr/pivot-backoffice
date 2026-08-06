-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS request_account_inquiries (
    `uuid` VARCHAR(36) NOT NULL PRIMARY KEY,
    `merchant_id` VARCHAR(36) NOT NULL,
    `account_inquiry_id` VARCHAR(36) NULL,
    `beneficiary_bank_code` VARCHAR(5) NOT NULL,
    `beneficiary_bank_name` VARCHAR(50) NULL,
    `beneficiary_account_no` VARCHAR(50) NULL,
    `beneficiary_account_name` VARCHAR(100) NULL,
    `status` VARCHAR(10) NULL,
    `metadata` JSON NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    KEY `merchant_id_IDX` (`merchant_id`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS request_account_inquiries;
-- +goose StatementEnd
