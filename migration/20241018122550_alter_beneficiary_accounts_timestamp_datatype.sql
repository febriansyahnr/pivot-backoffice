-- +goose Up
-- +goose StatementBegin
ALTER TABLE `beneficiary_accounts` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `beneficiary_accounts` MODIFY `updated_at` timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);
ALTER TABLE `beneficiary_accounts` MODIFY `deleted_at` timestamp NULL ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `beneficiary_accounts` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `beneficiary_accounts` MODIFY COLUMN `updated_at` DATETIME(3) NOT NULL;
ALTER TABLE `beneficiary_accounts` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd
