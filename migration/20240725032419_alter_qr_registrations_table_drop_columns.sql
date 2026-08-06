-- +goose Up
-- +goose StatementBegin
ALTER TABLE `qr_registrations` 
	DROP COLUMN `bod_info`, DROP COLUMN `boc_info`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `qr_registrations` 
    ADD COLUMN `bod_info` JSON NOT NULL AFTER `business_document`,
	ADD COLUMN `boc_info` JSON NOT NULL AFTER `bod_info`;
-- +goose StatementEnd
