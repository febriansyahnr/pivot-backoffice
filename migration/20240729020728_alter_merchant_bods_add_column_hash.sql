-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_bods` 
	ADD COLUMN `hash` VARCHAR(128) NOT NULL DEFAULT '' AFTER `identity_file`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_bods` DROP COLUMN `hash`;
-- +goose StatementEnd
