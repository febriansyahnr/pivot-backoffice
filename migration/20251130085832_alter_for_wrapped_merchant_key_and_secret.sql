-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants
	MODIFY COLUMN `callback_api_key` VARCHAR(255) DEFAULT NULL,
	MODIFY COLUMN `jit_api_key` VARCHAR(255) DEFAULT NULL,
	ADD COLUMN `callback_api_key_version` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `callback_api_key`,
	ADD COLUMN `jit_api_key_version` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `jit_api_key`;
ALTER TABLE merchant_auths ADD COLUMN `secret_version` SMALLINT UNSIGNED NOT NULL DEFAULT 0 AFTER `secret`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants
	MODIFY COLUMN `callback_api_key` VARCHAR(32) DEFAULT NULL,
	MODIFY COLUMN `jit_api_key` VARCHAR(32) DEFAULT NULL,
	DROP COLUMN `callback_api_key_version`,
	DROP COLUMN `jit_api_key_version`;
ALTER TABLE merchant_auths DROP COLUMN `secret_version`;
-- +goose StatementEnd
