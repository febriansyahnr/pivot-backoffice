-- +goose Up
-- +goose StatementBegin
CREATE TABLE `bulk_disbursements` (
     `uuid` varchar(100) NOT NULL,
     `merchant_id` varchar(100) NOT NULL,
     `file` varchar(255) NOT NULL,
     `file_failed` varchar(255) NULL,
     `status` VARCHAR(20) NOT NULL COMMENT "WAITING, IN_PROGRESS, DONE, PENDING",
     `created_by` VARCHAR(100) NULL,
     `created_at` datetime NOT NULL,
     `updated_at` datetime NOT NULL,
     `deleted_at` datetime NULL,
     PRIMARY KEY (`uuid`),
     KEY `bulk_disbursements_merchant_id_status_created_at_IDX` (`merchant_id`,`status`,`created_at`) USING BTREE,
     FOREIGN KEY (`merchant_id`) REFERENCES `merchants`(`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `bulk_disbursements`;
-- +goose StatementEnd
