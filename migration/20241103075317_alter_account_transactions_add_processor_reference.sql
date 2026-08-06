-- +goose Up
-- +goose StatementBegin
ALTER TABLE `account_transactions` ADD COLUMN `processor_reference` VARCHAR(30) NOT NULL DEFAULT '' COMMENT 'processor name';
ALTER TABLE `account_transactions` ADD COLUMN `processor_reference_id` VARCHAR(36) NOT NULL DEFAULT '' COMMENT 'processor transaction id';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE account_transactions DROP COLUMN processor_reference;
ALTER TABLE account_transactions DROP COLUMN processor_reference_id;
-- +goose StatementEnd
