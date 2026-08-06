-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS refunds (
    uuid                 VARCHAR(36) NOT NULL,
    merchant_id          VARCHAR(36) NOT NULL,
    client_reference_id  VARCHAR(100) NOT NULL,
    payment_id           VARCHAR(36) NOT NULL,
    payment_charge_id    VARCHAR(36) NOT NULL,
    currency             VARCHAR(10) NOT NULL,
    amount               DECIMAL(18, 2) NOT NULL,
    status               VARCHAR(36) NOT NULL,
    reason               VARCHAR(100) NOT NULL,
    description          VARCHAR(255) NOT NULL DEFAULT "",
    destination_type     VARCHAR(20) NOT NULL,
    `method`             VARCHAR(20) NOT NULL,
    created_at           timestamp NOT NULL,
    updated_at           timestamp NOT NULL,
    metadata             JSON,

    PRIMARY KEY (uuid),
    KEY refunds_merchant_id_client_reference_id_idx (merchant_id, client_reference_id) USING BTREE,
    KEY refunds_created_at_idx (created_at) USING BTREE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refunds;
-- +goose StatementEnd
