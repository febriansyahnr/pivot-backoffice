-- +goose Up
-- +goose StatementBegin
CREATE TABLE `user_role` (
    `uuid` varchar(255) NOT NULL,
    `user_id` varchar(255) NOT NULL,
    `role_id` varchar(255) NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `deleted_at` datetime,
    PRIMARY KEY (`uuid`),
    KEY `user_role_user_id_IDX` (`user_id`) USING BTREE,
    KEY `user_role_role_id_IDX` (`role_id`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `user_role`;
-- +goose StatementEnd
