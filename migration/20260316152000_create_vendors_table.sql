-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS vendors (
    uuid VARCHAR(36) PRIMARY KEY,
    merchant_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    beneficial_owner VARCHAR(255) NOT NULL,
    business_category VARCHAR(100) NOT NULL,
    avg_monthly_tpv_amount DECIMAL(20, 2) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(3) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    documents JSON NULL,
    status VARCHAR(8) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_vendors_status (status),
    INDEX idx_vendors_name (name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vendors;
-- +goose StatementEnd
