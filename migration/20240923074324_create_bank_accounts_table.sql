-- +goose Up
-- +goose StatementBegin
CREATE TABLE `bank_accounts` (
    `id` varchar(36) NOT NULL PRIMARY KEY,
    `merchant_id` varchar(36) NOT NULL,
    `beneficiary_bank_code` varchar(20) NOT NULL,
    `beneficiary_bank_name` varchar(60) NOT NULL,
    `beneficiary_account_no` varchar(60) NOT NULL,
    `beneficiary_account_name` varchar(60) NOT NULL,
    `created_by` varchar(36) NOT NULL,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` varchar(36) NOT NULL,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted` boolean NOT NULL DEFAULT false,
    `deleted_at` datetime DEFAULT NULL,
    UNIQUE KEY `bank_accounts_merchant_bank_code_account_no_comp_uniq_idx` (`merchant_id`,`beneficiary_bank_code`,`beneficiary_account_no`,`deleted`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `bank_accounts`;
-- +goose StatementEnd
