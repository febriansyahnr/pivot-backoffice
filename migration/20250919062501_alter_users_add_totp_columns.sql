-- +goose Up
-- +goose StatementBegin
ALTER TABLE `users` 
	ADD COLUMN `totp_wrapped_secret`    VARCHAR(255) NULL AFTER `pin_hash`,
	ADD COLUMN `totp_encrypt_version`   TINYINT UNSIGNED NULL AFTER `totp_wrapped_secret`,
	ADD COLUMN `totp_status`			VARCHAR(30) NOT NULL DEFAULT 'NOT_ENROLLED' AFTER `totp_encrypt_version`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `users`
	DROP COLUMN `totp_wrapped_secret`, DROP COLUMN `totp_encrypt_version`, DROP COLUMN `totp_status`;
-- +goose StatementEnd
