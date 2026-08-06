-- +goose Up
-- +goose StatementBegin
ALTER TABLE `qr_registrations`
	ADD COLUMN `callback_detail`   JSON NULL AFTER `acquirer_merchant_id`,
	ADD COLUMN `callback_datetime` DATETIME NULL AFTER `callback_detail`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `qr_registrations`
    DROP COLUMN `callback_detail`,
    DROP COLUMN `callback_datetime`;
-- +goose StatementEnd
