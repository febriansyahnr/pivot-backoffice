-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_fees`
    ADD COLUMN `payment_method_id` VARCHAR(100) NULL AFTER `merchant_id`,
    MODIFY COLUMN `amount_type` VARCHAR(20) NOT NULL DEFAULT 'AMOUNT' AFTER `payment_method_id`,
    MODIFY COLUMN `amount` DECIMAL(18,2) NOT NULL DEFAULT 0 AFTER `amount_type`,
    ADD COLUMN `percentage` DECIMAL(5,2) NOT NULL DEFAULT 0 AFTER `amount`,
    ADD COLUMN `deduction_type` VARCHAR(20) NOT NULL DEFAULT 'DIRECT' AFTER `reference`,
    ADD COLUMN `tax_type` VARCHAR(20) NOT NULL DEFAULT '' AFTER `deduction_type`,
    ADD COLUMN `tax_percentage` DECIMAL(5,2) NOT NULL DEFAULT 0 AFTER `tax_type`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_fees`
    DROP COLUMN `payment_method_id`,
    DROP COLUMN `percentage`,
    DROP COLUMN `deduction_type`,
    DROP COLUMN `tax_type`,
    DROP COLUMN `tax_percentage`;
-- +goose StatementEnd
