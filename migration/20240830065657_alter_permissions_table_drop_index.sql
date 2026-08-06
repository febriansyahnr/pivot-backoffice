-- +goose Up
-- +goose StatementBegin
ALTER TABLE `permissions` DROP INDEX `permissions_slug_IDX`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `permissions` ADD INDEX `permissions_slug_IDX`(`slug`) USING BTREE;
-- +goose StatementEnd
