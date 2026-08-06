-- +goose Up
-- +goose StatementBegin
ALTER TABLE `users` MODIFY `deactivate_at` timestamp NULL;
ALTER TABLE `users` MODIFY `last_login_at` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `users` MODIFY COLUMN `deactivate_at` DATETIME NULL;
ALTER TABLE `users` MODIFY COLUMN `last_login_at` DATETIME NULL;
-- +goose StatementEnd
