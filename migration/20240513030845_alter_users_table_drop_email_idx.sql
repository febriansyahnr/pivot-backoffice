-- +goose Up
-- +goose StatementBegin
ALTER TABLE `users` DROP INDEX `users_email_IDX`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `users` ADD INDEX `users_email_IDX`(`email`);
-- +goose StatementEnd
