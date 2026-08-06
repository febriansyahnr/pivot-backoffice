-- +goose Up
-- +goose StatementBegin
ALTER TABLE `transfers` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `transfers` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `transfers` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `transfers` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `transfers` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `transfers` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
