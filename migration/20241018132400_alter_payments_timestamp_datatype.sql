-- +goose Up
-- +goose StatementBegin
ALTER TABLE `payments` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `payments` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `payments` MODIFY `deleted_at` timestamp NULL;
ALTER TABLE `payments` MODIFY `expired_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `payments` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `payments` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `payments` MODIFY COLUMN `deleted_at` DATETIME NULL;
ALTER TABLE `payments` MODIFY COLUMN `expired_at` DATETIME NULL;
-- +goose StatementEnd
