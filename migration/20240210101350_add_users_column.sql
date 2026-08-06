-- +goose Up
-- +goose StatementBegin
CREATE TABLE `users` (
    `uuid` varchar(255) NOT NULL,
    `name` varchar(255) NOT NULL,
    `email` varchar(255) NOT NULL,
    `password` varchar(150) NOT NULL,
    `blocked_at` datetime,
    `merchant_id` varchar(255) NOT NULL,
    `refresh_token` varchar(255),
    `is_change_password` tinyint(1) NOT NULL DEFAULT '0',
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime,
    PRIMARY KEY (`uuid`),
    UNIQUE KEY `users_email_UNIQUE` (`email`),
    KEY `users_email_IDX` (`email`) USING BTREE,
    KEY `users_merchant_id_IDX` (`merchant_id`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `users`;
-- +goose StatementEnd
