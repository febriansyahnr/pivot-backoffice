-- +goose Up
-- +goose StatementBegin
ALTER TABLE withdrawals ADD COLUMN reference_id VARCHAR(50) NOT NULL DEFAULT '' AFTER merchant_id;
ALTER TABLE withdrawals ADD COLUMN description TEXT NOT NULL AFTER amount;
CREATE INDEX withdrawals_merchant_id_reference_id_idx ON withdrawals (merchant_id, reference_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE withdrawals DROP COLUMN reference_id;
ALTER TABLE withdrawals DROP COLUMN description;
DROP INDEX withdrawals_merchant_id_reference_id_idx ON withdrawals;
-- +goose StatementEnd
