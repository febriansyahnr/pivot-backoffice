-- +goose Up
-- +goose StatementBegin
ALTER TABLE `withdrawals` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `withdrawals` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `withdrawals` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `withdrawals` MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `withdrawals` MODIFY COLUMN `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `withdrawals` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
