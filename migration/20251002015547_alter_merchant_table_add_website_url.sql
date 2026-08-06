-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchants` ADD COLUMN `website` VARCHAR(255) NULL DEFAULT '' AFTER `description`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchants` DROP COLUMN `website`;
-- +goose StatementEnd
