-- +goose Up
-- +goose StatementBegin
CREATE TABLE `services` (
    `id` varchar(36) NOT NULL PRIMARY KEY,
    `name` varchar(150) NOT NULL,
    `category` varchar(150) NOT NULL,
    `channel` varchar(150) NOT NULL,
    `additional_info` json DEFAULT NULL,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` datetime DEFAULT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `services`;
-- +goose StatementEnd
