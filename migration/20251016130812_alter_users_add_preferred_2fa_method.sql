-- +goose Up
-- +goose StatementBegin
ALTER TABLE `users`
	ADD COLUMN `preferred_2fa_method` VARCHAR(5) AFTER `totp_status`;
-- +goose StatementEnd

-- +goose StatementBegin
-- Set preferred_2fa_method to 'TOTP' for users with ACTIVE TOTP status
UPDATE `users`
SET `preferred_2fa_method` = 'TOTP'
WHERE `totp_status` = 'ACTIVE';
-- +goose StatementEnd

-- +goose StatementBegin
-- Set preferred_2fa_method to 'OTP' for users without ACTIVE TOTP status
UPDATE `users`
SET `preferred_2fa_method` = 'OTP'
WHERE `totp_status` != 'ACTIVE' OR `totp_status` IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `users`
	DROP COLUMN `preferred_2fa_method`;
-- +goose StatementEnd
