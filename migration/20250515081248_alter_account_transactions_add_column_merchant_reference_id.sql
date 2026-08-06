-- +goose Up
-- +goose StatementBegin
ALTER TABLE `account_transactions`
	ADD COLUMN `merchant_reference_id` VARCHAR(100) NULL DEFAULT NULL AFTER `account_id`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `account_transactions` DROP COLUMN `merchant_reference_id`;
-- +goose StatementEnd
