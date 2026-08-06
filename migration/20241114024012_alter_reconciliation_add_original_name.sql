-- +goose Up
-- +goose StatementBegin
ALTER TABLE `reconciliations` ADD COLUMN `original_name` VARCHAR(100) NOT NULL DEFAULT '' AFTER `uuid`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `reconciliations` DROP COLUMN `original_name`;
-- +goose StatementEnd
