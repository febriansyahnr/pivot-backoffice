-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_account_transactions_aggregate_IDX ON `account_transactions` (`merchant_id`, `balance_id`, `status`, `created_at`);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `account_transactions` DROP INDEX idx_account_transactions_aggregate_IDX;
-- +goose StatementEnd
