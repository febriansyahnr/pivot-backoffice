-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts ADD COLUMN pending_transaction_cutoff_at TIMESTAMP NULL AFTER last_update_balance_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts DROP COLUMN pending_transaction_cutoff_at;
-- +goose StatementEnd
