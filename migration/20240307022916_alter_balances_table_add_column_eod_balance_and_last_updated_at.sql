-- +goose Up
-- +goose StatementBegin
ALTER TABLE balances CHANGE COLUMN amount eod_balance DECIMAL(19,4), ADD COLUMN last_update_balance_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER currency;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE balances CHANGE COLUMN eod_balance amount DECIMAL(19,4), DROP COLUMN last_update_balance_at;
-- +goose StatementEnd
