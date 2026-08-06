-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchants` ADD COLUMN `transaction_configs` JSON NULL AFTER `callback_api_key`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchants` DROP COLUMN `transaction_configs`;
-- +goose StatementEnd
