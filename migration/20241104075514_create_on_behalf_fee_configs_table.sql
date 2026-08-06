-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `on_behalf_fee_configs`(
	`id`				VARCHAR(36) NOT NULL PRIMARY KEY,
	`merchant_id` 		VARCHAR(36) NOT NULL,
	`type`				VARCHAR(20) NOT NULL,
	`sub_merchant_id` 	VARCHAR(36) NULL,
	`reference`			VARCHAR(20) NOT NULL,
	`payment_method`	VARCHAR(20) NULL,
	`amount_type`		VARCHAR(20) NOT NULL,
	`amount`			DECIMAL(10,2) NOT NULL DEFAULT 0,
	`percentage`		DECIMAL(5,2) NOT NULL DEFAULT 0,
	`created_at`		TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
	`updated_at`		TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
	`deleted_at`		TIMESTAMP NULL,
	KEY `on_behalf_fee_configs_merchant_id_idx`(`merchant_id`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `on_behalf_fee_configs`;
-- +goose StatementEnd
