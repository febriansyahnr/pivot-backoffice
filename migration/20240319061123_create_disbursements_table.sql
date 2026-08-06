-- +goose Up
-- +goose StatementBegin
CREATE TABLE `disbursements` (
    `uuid` varchar(100) NOT NULL,
    `merchant_id` varchar(100) NOT NULL,
    `bulk_id` varchar(100) NULL,
    `purpose_id` varchar(100) NULL,
    `sender_name` varchar(60) NOT NULL,
    `beneficiary_bank_code` varchar(20) NOT NULL,
    `beneficiary_bank_name` varchar(60) NULL,
    `beneficiary_account_no` varchar(60) NOT NULL,
    `beneficiary_account_name` varchar(60) NOT NULL,
    `processor_reference_id` varchar(100) NULL,
    `currency` varchar(3) NOT NULL COMMENT "IDR, USD",
    `amount` DECIMAL(18, 2) NOT NULL,
    `fee` DECIMAL(18, 2) NULL,
    `total_amount` DECIMAL(18, 2) NOT NULL,
    `status` VARCHAR(20) NOT NULL COMMENT "PENDING, APPROVED, REJECTED",
    `reason_type` VARCHAR(100) NULL,
    `reason_description` VARCHAR(255) NULL,
    `created_from` VARCHAR(60) NULL,
    `created_by` VARCHAR(100) NULL,
    `approved_by` VARCHAR(100) NULL,
    `approved_at` datetime NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime NULL,
    PRIMARY KEY (`uuid`),
    KEY `disbursements_merchant_id_IDX` (`merchant_id`) USING BTREE,
    KEY `disbursements_created_at_IDX` (`created_at`) USING BTREE,
    KEY `disbursements_bulk_id_IDX` (`bulk_id`) USING BTREE,
    FOREIGN KEY (`merchant_id`) REFERENCES `merchants`(`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `disbursements`;
-- +goose StatementEnd
