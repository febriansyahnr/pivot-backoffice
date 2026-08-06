-- +goose Up
-- +goose StatementBegin
ALTER TABLE `callback_masters` MODIFY `created_at` timestamp NULL;
ALTER TABLE `callback_masters` MODIFY `updated_at` timestamp NULL;
ALTER TABLE `callback_masters` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `callback_masters` MODIFY COLUMN `created_at` DATETIME  NULL;
ALTER TABLE `callback_masters` MODIFY COLUMN `updated_at` DATETIME  NULL;
ALTER TABLE `callback_masters` MODIFY COLUMN `deleted_at` DATETIME  NULL;
-- +goose StatementEnd