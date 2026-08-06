-- +goose Up
-- +goose StatementBegin
ALTER TABLE `callback_logs` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `callback_logs` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `callback_logs` MODIFY COLUMN `created_at` DATETIME NOT NULL;
ALTER TABLE `callback_logs` MODIFY COLUMN `updated_at` DATETIME NOT NULL;
-- +goose StatementEnd