-- +goose Up
-- +goose StatementBegin
CREATE TABLE `menus` (
     `uuid` varchar(100) NOT NULL,
     `slug` varchar(120) NOT NULL,
     `name` varchar(120) NOT NULL,
     `icon` varchar(255) NOT NULL,
     `path` varchar(255) NOT NULL,
     `level` int NOT NULL DEFAULT 0,
     `order` int NOT NULL DEFAULT 0,
     `parent_id` varchar(100) NULL,
     `created_at` datetime NOT NULL,
     `updated_at` datetime NOT NULL,
     `deleted_at` datetime NULL,
     PRIMARY KEY (`uuid`),
     UNIQUE KEY `menus_slug_UNIQUE` (`slug`),
     KEY `menus_slug_IDX` (`slug`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `menus`;
-- +goose StatementEnd
