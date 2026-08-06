-- +goose Up
-- +goose StatementBegin
CREATE TABLE `account_inquiries` (
    `uuid` varchar(100) NOT NULL,
    `beneficiary_bank_code` varchar(100) NOT NULL,
    `beneficiary_bank_name` varchar(100) NOT NULL,
    `beneficiary_account_no` varchar(100) NOT NULL,
    `beneficiary_account_name` varchar(100) NOT NULL,
    `response` json NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime NULL,
    PRIMARY KEY (`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `account_inquiries`;
-- +goose StatementEnd
