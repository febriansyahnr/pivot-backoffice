-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchant_fees RENAME COLUMN `payment_method_id` TO `payment_method`;
ALTER TABLE merchant_fees MODIFY COLUMN `payment_method` VARCHAR(36) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_fees MODIFY COLUMN `payment_method` VARCHAR(100) NULL;
ALTER TABLE merchant_fees RENAME COLUMN `payment_method` TO `payment_method_id`;
-- +goose StatementEnd
