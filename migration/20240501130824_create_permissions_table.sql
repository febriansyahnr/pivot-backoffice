-- +goose Up
-- +goose StatementBegin
CREATE TABLE `permissions` (
     `uuid` varchar(100) NOT NULL,
     `slug` varchar(120) NOT NULL,
     `name` varchar(120) NOT NULL,
     `description` varchar(255) NOT NULL,
     `group` varchar(120) NOT NULL,
     `created_at` datetime NOT NULL,
     `updated_at` datetime NOT NULL,
     `deleted_at` datetime NULL,
     PRIMARY KEY (`uuid`),
     UNIQUE KEY `permissions_slug_UNIQUE` (`slug`),
     KEY `permissions_slug_IDX` (`slug`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `permissions`;
-- +goose StatementEnd
