-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants ADD jit_api_key VARCHAR(32) DEFAULT NULL AFTER callback_api_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants DROP COLUMN jit_api_key;
-- +goose StatementEnd
