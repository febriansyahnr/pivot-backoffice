-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchant_fees
    ADD COLUMN settlement_model VARCHAR(20) NULL AFTER settlement_configs;

ALTER TABLE on_behalf_fee_configs
    ADD COLUMN settlement_model VARCHAR(20) NULL AFTER payment_method;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_fees
    DROP COLUMN settlement_model;

ALTER TABLE on_behalf_fee_configs
    DROP COLUMN settlement_model;
-- +goose StatementEnd
