-- +goose Up
-- +goose StatementBegin
CREATE TABLE `activity_logs` (
    `id` varchar(100) NOT NULL,
    `merchant_id` varchar(100) NOT NULL,
    `user_id` varchar(100) NULL,
    `tag` varchar(60) NOT NULL,
    `activity` varchar(255) NOT NULL,
    `service_name` varchar(60) NOT NULL,
    `parameter` json NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    PRIMARY KEY (`id`),
    KEY `activity_logs_merchant_id_IDX` (`merchant_id`) USING BTREE,
    KEY `activity_logs_created_at_IDX` (`created_at`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `activity_logs`;
-- +goose StatementEnd
