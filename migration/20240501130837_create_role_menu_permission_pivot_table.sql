-- +goose Up
-- +goose StatementBegin
CREATE TABLE `role_menu_permission` (
     `role_id` varchar(100) NOT NULL,
     `menu_id` varchar(100) NOT NULL,
     `permission_id` varchar(100) NOT NULL,
     UNIQUE KEY `role_menu_permission_UNIQUE` (`role_id`, `menu_id`, `permission_id`),
     KEY `role_menu_permission_role_id_IDX` (`role_id`) USING BTREE,
     KEY `role_menu_permission_permission_id_IDX` (`permission_id`) USING BTREE,
     KEY `role_menu_permission_menu_id_IDX` (`menu_id`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `role_menu_permission`;
-- +goose StatementEnd
