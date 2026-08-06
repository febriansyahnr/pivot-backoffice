-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `merchant_documents` (
	`id`			VARCHAR(36) NOT NULL PRIMARY KEY,
	`merchant_id` 	VARCHAR(36) NOT NULL,
	`type`			VARCHAR(50) NOT NULL,
    `number`        VARCHAR(32) NOT NULL DEFAULT '',
	`location`		JSON		NOT NULL,
	`status`		VARCHAR(30) NOT NULL,
	`created_by`	VARCHAR(50) NOT NULL,
	`created_at`	DATETIME	NOT NULL DEFAULT NOW(),
	`approved_by`	VARCHAR(50) NOT NULL DEFAULT '',
	`approved_at`	DATETIME	NULL,
	`updated_at`	DATETIME	NOT NULL DEFAULT NOW(),
	`deleted_at`	DATETIME	NULL,
	UNIQUE KEY `merchant_documents_merchant_id_type_comp_uniq_idx`(`merchant_id`,`type`,`deleted_at`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `merchant_documents`;
-- +goose StatementEnd
