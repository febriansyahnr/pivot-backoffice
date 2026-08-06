-- +goose Up
-- +goose StatementBegin
ALTER TABLE `payment_methods` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `payment_methods` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `payment_methods` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `payment_methods` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `payment_methods` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `payment_methods` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
