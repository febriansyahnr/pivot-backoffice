-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS creditcards (
    uuid VARCHAR(36) PRIMARY KEY,
    reference_id VARCHAR(100) NOT NULL,
    processor_reference_number VARCHAR(100) NOT NULL,
    bank_merchant_id VARCHAR(50),
    amount DECIMAL(18, 2) NOT NULL,
    fee DECIMAL(18, 2) NOT NULL,
    total_amount DECIMAL(18, 2) NOT NULL,
    currency CHAR(3) NOT NULL,
    authentication_method VARCHAR(13) NOT NULL DEFAULT 'CHALLENGE',
    status VARCHAR(50) NOT NULL DEFAULT 'WAITING_FOR_PAYMENT',
    payment_url VARCHAR(255) NOT NULL DEFAULT '',
    authentication_result JSON,
    card_data JSON,
    expired_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL 15 MINUTE),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS creditcards;
-- +goose StatementEnd
