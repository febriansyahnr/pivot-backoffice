-- +goose Up
-- +goose StatementBegin
ALTER TABLE `menus` DROP INDEX `menus_slug_IDX`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `menus` ADD INDEX `menus_slug_IDX`(`slug`) USING BTREE;
-- +goose StatementEnd
