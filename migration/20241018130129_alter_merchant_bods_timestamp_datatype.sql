-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_bods` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_bods` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_bods` MODIFY `approved_at` timestamp NULL;
ALTER TABLE `merchant_bods` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_bods` MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_bods` MODIFY COLUMN `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_bods` MODIFY COLUMN `approved_at` DATETIME NULL;
ALTER TABLE `merchant_bods` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
