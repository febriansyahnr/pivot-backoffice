-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_account_transactions_reference_id ON account_transactions(reference_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_account_transactions_reference_id ON account_transactions;
-- +goose StatementEnd
