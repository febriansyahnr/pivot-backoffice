-- +goose Up
-- +goose StatementBegin
ALTER TABLE `disbursement_top_up_references` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `disbursement_top_up_references` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `disbursement_top_up_references` MODIFY `deleted_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `disbursement_top_up_references` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `disbursement_top_up_references` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
ALTER TABLE `disbursement_top_up_references` MODIFY COLUMN `deleted_at` DATETIME NULL;
-- +goose StatementEnd