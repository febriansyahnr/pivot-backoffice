-- +goose Up
-- +goose StatementBegin
ALTER TABLE `products` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `products` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `products` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `products` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
-- +goose StatementEnd
