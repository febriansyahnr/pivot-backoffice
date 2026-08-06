-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `merchant_bods`(
	`id`				VARCHAR(36)  NOT NULL PRIMARY KEY,
	`merchant_id` 		VARCHAR(36)  NOT NULL,
	`position` 			VARCHAR(20)  NOT NULL,
	`name`				VARCHAR(255) NOT NULL,
	`identity_number`	VARCHAR(32)  NOT NULL,
	`identity_file`		JSON		 NOT NULL,
	`position_long`		VARCHAR(100) NOT NULL DEFAULT '',
	`status`			VARCHAR(30)  NOT NULL,
	`created_by`		VARCHAR(50)  NOT NULL,
	`created_at`		DATETIME	 NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`approved_by`		VARCHAR(50)  NOT NULL,
	`approved_at`		DATETIME     NULL,
	`updated_at`		DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`deleted`			BOOLEAN      NOT NULL DEFAULT false,
	`deleted_at`		DATETIME     NULL,
	UNIQUE KEY `merchant_bods_merchant_position_identity_comp_uniq_idx`(`merchant_id`,`position`,`identity_number`,`deleted`)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `merchant_bods`;
-- +goose StatementEnd
