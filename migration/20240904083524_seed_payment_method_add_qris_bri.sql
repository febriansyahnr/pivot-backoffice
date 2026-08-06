-- +goose Up
-- +goose StatementBegin
INSERT INTO `payment_methods` (`uuid`,`type`,`category`,`name`,`description`,`logo`,`acquirer`,`bank_name`,`created_at`,`updated_at`,`deleted_at`) VALUES (uuid(), 'QRIS', 'PAYMENT', 'QRIS BRI', 'QRIS BRI', 'bri.svg', 'bri', 'Bank BRI', now(), now(), NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
