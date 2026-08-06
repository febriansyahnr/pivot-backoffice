-- +goose Up
-- +goose StatementBegin
ALTER TABLE `accounts` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `accounts` MODIFY `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
-- +goose StatementEnd
