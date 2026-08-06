-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchant_tnc_signing_histories
    RENAME COLUMN document_url TO document_path;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_tnc_signing_histories
    RENAME COLUMN document_path TO document_url;
-- +goose StatementEnd
