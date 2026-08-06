-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_fees` MODIFY COLUMN `reference` VARCHAR(100) NOT NULL AFTER `merchant_id`,
	ADD COLUMN `channel` VARCHAR(75) NULL AFTER `payment_method`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_fees` MODIFY COLUMN `reference` VARCHAR(100) NOT NULL AFTER `percentage`,
	DROP COLUMN `channel`;
-- +goose StatementEnd
