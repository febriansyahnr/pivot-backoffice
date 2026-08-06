-- +goose Up
-- +goose StatementBegin
ALTER TABLE `provinces` MODIFY `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `provinces` MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
