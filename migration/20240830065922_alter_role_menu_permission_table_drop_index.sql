-- +goose Up
-- +goose StatementBegin
ALTER TABLE `role_menu_permission` DROP INDEX `role_menu_permission_role_id_IDX`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `role_menu_permission` ADD INDEX `role_menu_permission_role_id_IDX`(`role_id`) USING BTREE;
-- +goose StatementEnd
