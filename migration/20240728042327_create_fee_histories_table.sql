-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `fee_histories` (
    id							VARCHAR(36) NOT NULL PRIMARY KEY,
    merchant_id 				VARCHAR(36) NOT NULL,
    merchant_fee_id				VARCHAR(36) NULL,
    reference   				VARCHAR(50) NOT NULL,
    reference_id   				VARCHAR(36) NOT NULL,
    currency                    VARCHAR(3) NOT NULL,
    amount                      DECIMAL(18, 2) NOT NULL,
    created_at					TIMESTAMP 	NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at					TIMESTAMP 	NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY `fee_histories_merchant_usecase_idx`(`merchant_id`, `reference_id`, `reference`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `fee_histories`;
-- +goose StatementEnd
