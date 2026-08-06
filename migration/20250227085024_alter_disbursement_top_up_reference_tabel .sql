-- +goose Up
-- +goose StatementBegin
RENAME TABLE `disbursement_top_up_references` TO `merchant_top_up_references`;
ALTER TABLE `merchant_top_up_references` ADD COLUMN `account_name` VARCHAR(30) NOT NULL DEFAULT 'DISBURSEMENT' AFTER `merchant_id`;

UPDATE payment_methods SET category = 'MERCHANT_TOP_UP' WHERE category = 'DISBURSEMENT_TOP_UP';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
RENAME TABLE `merchant_top_up_references` TO `disbursement_top_up_references`;
ALTER TABLE `disbursement_top_up_references` DROP COLUMN `account_name`;

UPDATE payment_methods SET category = 'DISBURSEMENT_TOP_UP' WHERE category = 'MERCHANT_TOP_UP';
-- +goose StatementEnd
