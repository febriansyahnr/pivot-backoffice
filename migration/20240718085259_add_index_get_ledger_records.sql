-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_account_id_records_IDX ON account_transactions (account_id,transaction_timestamp, status, type)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_account_id_records_IDX on account_transactions
-- +goose StatementEnd
