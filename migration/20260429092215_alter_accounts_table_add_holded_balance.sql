-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts ADD COLUMN holded_balance DECIMAL(19,4) DEFAULT 0 AFTER eod_balance;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts DROP COLUMN holded_balance;
-- +goose StatementEnd
