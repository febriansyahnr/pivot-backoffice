-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tncs (
    uuid VARCHAR(36) PRIMARY KEY,
    version VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    markdown_content LONGTEXT NOT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 0,
    created_by VARCHAR(36) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

CREATE TABLE IF NOT EXISTS merchant_tnc_signing_histories (
    uuid VARCHAR(36) PRIMARY KEY,
    merchant_id VARCHAR(36) NOT NULL,
    tnc_id VARCHAR(36) NOT NULL,
    version VARCHAR(50) NOT NULL,
    signed_by VARCHAR(36) NOT NULL,
    signed_by_email VARCHAR(255) NOT NULL,
    signed_at DATETIME NOT NULL,
    document_url VARCHAR(512) NULL,
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(512) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_merchant_tnc_signing_histories_merchant_id_tnc_id (merchant_id, tnc_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS merchant_tnc_signing_histories;
DROP TABLE IF EXISTS tncs;
-- +goose StatementEnd
