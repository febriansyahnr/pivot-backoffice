-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `ip_whitelist_configurations`(
	`id`				VARCHAR(36) NOT NULL PRIMARY KEY,
	`merchant_id` 		VARCHAR(36) NOT NULL,
	`ip`         		VARCHAR(50) NOT NULL,
    `subnet`        	VARCHAR(10) NOT NULL,
    `priority`          INT NOT NULL,
    `action`			VARCHAR(20) NOT NULL,
    `status`			VARCHAR(20) NOT NULL,
    `description`		VARCHAR(200) NOT NULL,
	`created_at`		TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
	`updated_at`		TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
	KEY `ip_whitelist_configurations_merchant_id_idx`(`merchant_id`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `ip_whitelist_configurations`;
-- +goose StatementEnd
