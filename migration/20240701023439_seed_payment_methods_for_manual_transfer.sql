-- +goose Up
-- +goose StatementBegin
INSERT INTO `payment_methods` (`uuid`,`type`,`category`,`name`,`description`,`logo`,`acquirer`,`bank_name`,`created_at`,`updated_at`,`deleted_at`)
VALUES
    (uuid(), 'BANK_TRANSFER', 'DISBURSEMENT_TOP_UP', 'Mandiri (Manual Transfer)', 'Harsya Remitindo\n1340000800606', 'mandiri.svg', 'mandiri', 'Bank Mandiri', now(), now(), NULL),
    (uuid(), 'BANK_TRANSFER', 'DISBURSEMENT_TOP_UP', 'BRI (Manual Transfer)', 'Harsya Remitindo\n111501777777306', 'bri.svg', 'bri', 'Bank BRI', now(), now(), NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
