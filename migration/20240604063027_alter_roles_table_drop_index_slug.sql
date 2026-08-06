-- +goose Up
-- +goose StatementBegin
ALTER TABLE `roles` DROP INDEX `roles_slug_IDX`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `roles` ADD INDEX `roles_slug_IDX`(slug);
-- +goose StatementEnd
