-- +goose Up
-- +goose StatementBegin
ALTER TABLE `account_transactions`
    ADD COLUMN `settlement_model` VARCHAR(20) NULL AFTER `settlement_status`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `account_transactions`
    DROP COLUMN `settlement_model`;
-- +goose StatementEnd
