-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_logged_in_devices (
    `uuid`					VARCHAR(36) NOT NULL PRIMARY KEY,
    `user_id`	    		VARCHAR(36) NOT NULL,
    `device_identifier`     VARCHAR(255) NOT NULL,
    `additional_info`		JSON NULL,
    `status`				VARCHAR(30) NOT NULL COMMENT "ACTIVE, BLOCKED, etc",
    `created_at`			TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at`			TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    `deleted_at`			TIMESTAMP NULL,
    KEY `user_logged_in_devices_user_id_IDX` (`user_id`) USING BTREE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_logged_in_devices;
-- +goose StatementEnd
