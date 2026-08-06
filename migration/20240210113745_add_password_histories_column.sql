-- +goose Up
-- +goose StatementBegin
CREATE TABLE `password_histories` (
    `uuid` varchar(255) NOT NULL,
    `user_id` varchar(255) NOT NULL,
    `password_hash` varchar(255) NOT NULL,
    `created_at` datetime NOT NULL,
    PRIMARY KEY (`uuid`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`uuid`),
    KEY `password_histories_user_id_IDX` (`user_id`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `password_histories`;
-- +goose StatementEnd
