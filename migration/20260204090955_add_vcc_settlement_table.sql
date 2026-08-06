-- +goose Up
-- +goose StatementBegin
CREATE TABLE `vcc_settlements` (
    `uuid` varchar(36) NOT NULL,
    `rcn_id` varchar(36) NOT NULL,
    `acquirer_reference_number` varchar(30) NOT NULL,
    `status` varchar(30) NOT NULL,
    `reference_no` varchar(20) NOT NULL,
    `authorization_no` varchar(20) NOT NULL,
    `posting_date` datetime NOT NULL,
    `billing_cycle` tinyint NOT NULL,
    `source_amount` json NOT NULL,
    `billing_amount` json NOT NULL,
    `transaction_date` datetime NOT NULL,
    `settlement_date` datetime NOT NULL,
    `merchant_name` varchar(50) NOT NULL,
    `merchant_country` varchar(10) NOT NULL,
    `merchant_category` varchar(10) NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime NULL,
    PRIMARY KEY (`uuid`),
    KEY `idx_vcc_settlements_posting_date_rcn_id_acquirer_ref_number` (`posting_date`, `rcn_id`, `acquirer_reference_number`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `vcc_settlements`;
-- +goose StatementEnd
