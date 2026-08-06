-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS merchant_fees (
    `uuid`					VARCHAR(36) NOT NULL PRIMARY KEY,
    `merchant_id`			VARCHAR(36) NOT NULL,
    `amount`                DECIMAL(18, 2) NOT NULL,
    `fee_type`              VARCHAR(100) NOT NULL,
    `created_at`			TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at`			TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    `deleted_at`			TIMESTAMP NULL,

    KEY `merchant_id_IDX` (`merchant_id`) USING BTREE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS merchant_fees;
-- +goose StatementEnd
