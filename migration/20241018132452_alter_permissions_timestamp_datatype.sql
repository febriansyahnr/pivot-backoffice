-- +goose Up
-- +goose StatementBegin
ALTER TABLE `permissions` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `permissions` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `permissions` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `permissions` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `permissions` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `permissions` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
