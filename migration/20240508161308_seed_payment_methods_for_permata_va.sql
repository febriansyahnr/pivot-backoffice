-- +goose Up
-- +goose StatementBegin
INSERT INTO `payment_methods` (`uuid`,`type`,`category`,`name`,`description`,`logo`,`acquirer`,`bank_name`,`created_at`,`updated_at`,`deleted_at`)
VALUES
    (uuid(), 'VIRTUAL_ACCOUNT', 'PAYMENT', 'VA Permata', '-', 'permata.svg', 'permata', 'Bank Permata', now(), now(), NULL),
    (uuid(), 'VIRTUAL_ACCOUNT', 'DISBURSEMENT_TOP_UP', 'VA Permata', '-', 'permata.svg', 'permata', 'Bank Permata', now(), now(), NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
