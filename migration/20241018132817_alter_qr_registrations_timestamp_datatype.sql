-- +goose Up
-- +goose StatementBegin
ALTER TABLE `qr_registrations` MODIFY `callback_datetime` timestamp NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `qr_registrations` MODIFY COLUMN `callback_datetime` DATETIME NULL;
-- +goose StatementEnd
