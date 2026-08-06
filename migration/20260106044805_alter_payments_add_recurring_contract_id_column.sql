-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments ADD COLUMN recurring_contract_id CHAR(36) NULL AFTER processor_reference_number;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments DROP COLUMN recurring_contract_id;
-- +goose StatementEnd
