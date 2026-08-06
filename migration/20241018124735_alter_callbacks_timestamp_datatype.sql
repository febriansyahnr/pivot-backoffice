-- +goose Up
-- +goose StatementBegin
ALTER TABLE `callbacks` MODIFY `created_at` timestamp NULL;
ALTER TABLE `callbacks` MODIFY `updated_at` timestamp NULL;
ALTER TABLE `callbacks` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `callbacks` MODIFY COLUMN `created_at` DATETIME NULL;
ALTER TABLE `callbacks` MODIFY COLUMN `updated_at` DATETIME NULL;
ALTER TABLE `callbacks` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd