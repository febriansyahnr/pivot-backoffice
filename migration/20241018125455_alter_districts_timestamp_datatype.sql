-- +goose Up
-- +goose StatementBegin
ALTER TABLE `districts` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `districts` MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd