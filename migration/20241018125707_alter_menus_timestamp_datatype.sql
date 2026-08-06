-- +goose Up
-- +goose StatementBegin
ALTER TABLE `menus` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `menus` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `menus` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `menus` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `menus` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `menus` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
