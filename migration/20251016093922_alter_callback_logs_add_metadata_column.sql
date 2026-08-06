-- +goose Up
-- +goose StatementBegin
ALTER TABLE `callback_logs` ADD COLUMN `metadata` JSON NULL AFTER `retry`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `callback_logs` DROP COLUMN `metadata`;
-- +goose StatementEnd
