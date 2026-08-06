-- +goose Up
-- +goose StatementBegin
ALTER TABLE `disbursements` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `disbursements` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `disbursements` MODIFY `approved_at` timestamp NULL;
ALTER TABLE `disbursements` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `disbursements` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `disbursements` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `disbursements` MODIFY COLUMN `approved_at` DATETIME NULL;
ALTER TABLE `disbursements` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd