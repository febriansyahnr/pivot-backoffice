-- +goose Up
-- +goose StatementBegin
DROP INDEX idx_account_id_records_IDX on account_transactions;
CREATE INDEX account_transactions_transaction_timestamp_idx ON account_transactions (reference, type, transaction_timestamp) ALGORITHM=INPLACE LOCK=NONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE INDEX idx_account_id_records_IDX ON account_transactions (account_id,transaction_timestamp, status, type) ALGORITHM=INPLACE LOCK=NONE;
DROP INDEX account_transactions_transaction_timestamp_idx ON account_transactions;
-- +goose StatementEnd
