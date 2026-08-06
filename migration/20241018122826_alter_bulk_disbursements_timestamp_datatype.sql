-- +goose Up
-- +goose StatementBegin
ALTER TABLE `bulk_disbursements` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `bulk_disbursements` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `bulk_disbursements` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `bulk_disbursements` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `bulk_disbursements` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `bulk_disbursements` MODIFY COLUMN `deleted_at` DATETIME  NULL;
-- +goose StatementEnd