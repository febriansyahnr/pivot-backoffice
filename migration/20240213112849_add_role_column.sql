-- +goose Up
-- +goose StatementBegin
CREATE TABLE `roles` (
    `uuid` varchar(255) NOT NULL,
    `name` varchar(255) NOT NULL,
    `slug` varchar(255) NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime,
    PRIMARY KEY (`uuid`),
    UNIQUE KEY `roles_slug_UNIQUE` (`slug`),
    KEY `roles_slug_IDX` (`slug`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `roles`;
-- +goose StatementEnd
