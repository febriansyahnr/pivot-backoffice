-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchant_fees ADD COLUMN amount_type varchar(20);
ALTER TABLE merchant_fees RENAME COLUMN fee_type TO reference;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_fees DROP COLUMN amount_type;
ALTER TABLE merchant_fees RENAME COLUMN reference TO fee_type;
-- +goose StatementEnd
