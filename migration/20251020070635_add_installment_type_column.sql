-- +goose Up
-- +goose StatementBegin
ALTER TABLE installment_plans ADD COLUMN installment_type VARCHAR(20) DEFAULT '' AFTER settlement_type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE installment_plans DROP COLUMN installment_type;
-- +goose StatementEnd
