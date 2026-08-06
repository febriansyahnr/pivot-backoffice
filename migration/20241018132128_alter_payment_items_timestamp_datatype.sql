-- +goose Up
-- +goose StatementBegin
ALTER TABLE `payment_items` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `payment_items` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `payment_items` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `payment_items` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `payment_items` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `payment_items` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
