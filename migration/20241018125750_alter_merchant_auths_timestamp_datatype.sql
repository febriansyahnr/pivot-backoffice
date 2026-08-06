-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_auths` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_auths` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `merchant_auths` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_auths` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `merchant_auths` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `merchant_auths` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
