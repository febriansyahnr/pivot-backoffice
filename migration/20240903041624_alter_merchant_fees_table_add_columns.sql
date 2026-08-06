-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_fees` 
	ADD COLUMN `max_fee_amount` DECIMAL(18, 2) NULL AFTER `amount`,
	ADD COLUMN `deduction_day` TINYINT UNSIGNED NULL AFTER `deduction_type`,
	ADD COLUMN `deduction_last_date` TIMESTAMP NULL AFTER `deduction_day`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_fees`
	DROP COLUMN `max_fee_amount`, DROP COLUMN `deduction_day`, DROP COLUMN `deduction_last_date`;
-- +goose StatementEnd
