-- +goose Up
-- +goose StatementBegin
ALTER TABLE `beneficiary_accounts` ADD COLUMN `metadata` JSON NULL AFTER `beneficiary_account_name`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `beneficiary_accounts` DROP COLUMN `metadata`;
-- +goose StatementEnd
