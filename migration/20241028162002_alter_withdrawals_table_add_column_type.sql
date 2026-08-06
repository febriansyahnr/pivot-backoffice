-- +goose Up
-- +goose StatementBegin
ALTER TABLE withdrawals
    ADD COLUMN `type` VARCHAR(20) NOT NULL DEFAULT 'MANUAL' AFTER beneficiary_account_name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE withdrawals DROP COLUMN `type`;
-- +goose StatementEnd
