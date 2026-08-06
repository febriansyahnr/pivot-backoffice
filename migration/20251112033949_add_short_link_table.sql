-- +goose Up
-- +goose StatementBegin
CREATE TABLE short_links (
    uuid varchar(36),
    reference varchar(36),
    code varchar(50),
    destination_url varchar(512),
    created_at DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expired_at DATETIME  NOT NULL,
    PRIMARY KEY (`uuid`),
    INDEX idx_short_links_code (`code`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE short_links;
-- +goose StatementEnd