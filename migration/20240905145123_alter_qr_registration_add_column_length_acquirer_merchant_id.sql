-- +goose Up
-- +goose StatementBegin
ALTER TABLE qr_registrations MODIFY COLUMN acquirer_merchant_id VARCHAR(20);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE qr_registrations MODIFY COLUMN acquirer_merchant_id VARCHAR(12);
-- +goose StatementEnd
