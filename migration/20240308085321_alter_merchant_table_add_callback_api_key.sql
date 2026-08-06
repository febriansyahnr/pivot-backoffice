-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants ADD COLUMN callback_api_key VARCHAR(32) AFTER mid
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants DROP COLUMN callback_api_key;
-- +goose StatementEnd
