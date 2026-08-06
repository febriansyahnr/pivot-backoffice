-- +goose Up
-- +goose StatementBegin
ALTER TABLE backend_portal.merchant_documents CHANGE `number` identifier varchar(50) DEFAULT '' NOT NULL COMMENT 'store document number, letter number, etc';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backend_portal.merchant_documents CHANGE identifier `number` varchar(32) DEFAULT '' NOT NULL;
-- +goose StatementEnd
