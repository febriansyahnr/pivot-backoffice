-- +goose Up
-- +goose StatementBegin
ALTER TABLE `account_inquiries` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `account_inquiries` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `account_inquiries` MODIFY `deleted_at` timestamp NULL ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `account_inquiries` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `account_inquiries` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `account_inquiries` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
