-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payout_manual_processing_accounts (
    uuid VARCHAR(36) PRIMARY KEY,
    merchant_id VARCHAR(36) NOT NULL,
    bank_code VARCHAR(3) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    status VARCHAR(8) NOT NULL DEFAULT 'ACTIVE',
    updated_by VARCHAR(50) NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_payout_manual_processing_accounts_status (status)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payout_manual_processing_accounts;
-- +goose StatementEnd
