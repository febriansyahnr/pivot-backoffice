-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchant_documents ADD COLUMN notes TEXT NOT NULL AFTER status;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_documents DROP COLUMN notes;
-- +goose StatementEnd
