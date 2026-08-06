-- +goose Up
-- +goose StatementBegin
ALTER TABLE qr_registrations
    ADD COLUMN acquirer_terminal_id VARCHAR(36) AFTER acquirer_merchant_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE qr_registrations
    DROP COLUMN acquirer_terminal_id;
-- +goose StatementEnd
