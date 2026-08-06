-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `disbursement_top_up_references` (
    `uuid` varchar(100) NOT NULL,
    `merchant_id` varchar(100) NOT NULL,
    `payment_method_id` varchar(100) NOT NULL,
    `reference_number` varchar(255) NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime NULL,
    PRIMARY KEY (`uuid`),
    KEY `disbursement_top_up_reference_merchant_id_IDX` (`merchant_id`) USING BTREE,
    KEY `disbursement_top_up_reference_payment_method_id_IDX` (`payment_method_id`) USING BTREE,
    KEY `disbursement_top_up_reference_reference_number_IDX` (`reference_number`) USING BTREE,
    FOREIGN KEY (`merchant_id`) REFERENCES `merchants`(`uuid`),
    FOREIGN KEY (`payment_method_id`) REFERENCES `payment_methods`(`uuid`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `disbursement_top_up_references`;
-- +goose StatementEnd
