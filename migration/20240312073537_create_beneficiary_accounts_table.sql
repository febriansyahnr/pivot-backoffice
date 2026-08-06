-- +goose Up
-- +goose StatementBegin
CREATE TABLE `beneficiary_accounts` (
     `uuid` varchar(100) NOT NULL,
     `merchant_id` varchar(100) NOT NULL,
     `beneficiary_bank_code` varchar(20) NOT NULL,
     `beneficiary_bank_name` varchar(60) NOT NULL,
     `beneficiary_account_no` varchar(60) NOT NULL,
     `beneficiary_account_name` varchar(60) NOT NULL,
     `created_at` datetime NOT NULL,
     `updated_at` datetime NOT NULL,
     `deleted_at` datetime NULL,
     PRIMARY KEY (`uuid`),
     KEY `beneficiary_accounts_merchant_id_IDX` (`merchant_id`) USING BTREE,
     FOREIGN KEY (`merchant_id`) REFERENCES `merchants`(`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `beneficiary_accounts`;
-- +goose StatementEnd
