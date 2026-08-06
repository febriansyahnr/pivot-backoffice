-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_documents` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_documents` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_documents` MODIFY `approved_at` timestamp NULL;
ALTER TABLE `merchant_documents` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_documents` MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_documents` MODIFY COLUMN `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_documents` MODIFY COLUMN `approved_at` DATETIME NULL;
ALTER TABLE `merchant_documents` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
