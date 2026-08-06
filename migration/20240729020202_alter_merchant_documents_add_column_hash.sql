-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_documents` 
	ADD COLUMN `hash` VARCHAR(128) NOT NULL DEFAULT '' AFTER `location`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_documents` DROP COLUMN `hash`;
-- +goose StatementEnd
