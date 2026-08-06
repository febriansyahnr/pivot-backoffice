-- +goose Up
-- +goose StatementBegin
ALTER TABLE `account_transactions`
    ADD COLUMN `additional_info` JSON NULL AFTER `transaction_timestamp`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `account_transactions`
    DROP COLUMN `additional_info`;
-- +goose StatementEnd
