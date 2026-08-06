-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchant_fees ADD COLUMN reference_type varchar(30) NOT NULL DEFAULT '' AFTER reference;
ALTER TABLE on_behalf_fee_configs ADD COLUMN reference_type varchar(30) NOT NULL DEFAULT '' AFTER reference;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_fees DROP COLUMN reference_type;
ALTER TABLE on_behalf_fee_configs DROP COLUMN reference_type;
-- +goose StatementEnd
