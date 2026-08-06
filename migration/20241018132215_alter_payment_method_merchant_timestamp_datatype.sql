-- +goose Up
-- +goose StatementBegin
ALTER TABLE `payment_method_merchant` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `payment_method_merchant` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `payment_method_merchant` MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `payment_method_merchant` MODIFY COLUMN `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
