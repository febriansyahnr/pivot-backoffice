-- +goose Up
-- +goose StatementBegin
ALTER TABLE account_transactions
    ADD COLUMN processor_transaction_id VARCHAR(36) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE account_transactions
    DROP COLUMN processor_transaction_id;
-- +goose StatementEnd
