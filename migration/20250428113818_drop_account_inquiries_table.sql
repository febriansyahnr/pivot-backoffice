-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS account_inquiries;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS account_inquiries (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    payment_id VARCHAR(36) NOT NULL,
    account_number VARCHAR(255) NOT NULL,
    bank_code VARCHAR(15) NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    payment_method_id VARCHAR(36) NOT NULL,
    payment_method_name VARCHAR(255) NOT NULL,
    payload TEXT NOT NULL,
    response TEXT NOT NULL,
    status VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
-- +goose StatementEnd 