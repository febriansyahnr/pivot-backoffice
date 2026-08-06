-- +goose Up
-- +goose StatementBegin

-- Merchant
ALTER TABLE merchants  ADD COLUMN parent_id varchar(36);

-- Balance
ALTER TABLE balances ADD COLUMN type varchar(20);
ALTER TABLE balances ADD COLUMN user_type varchar(20);
ALTER TABLE balances ADD COLUMN balance decimal(19,4);

RENAME TABLE balances to accounts;

-- Account Transaction
ALTER TABLE account_transactions RENAME COLUMN balance_id TO account_id;
ALTER TABLE account_transactions ADD COLUMN reference varchar(20);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Merchant
ALTER TABLE merchants  DROP COLUMN parent_id;

-- Balance
RENAME TABLE accounts to balances;
ALTER TABLE balances DROP COLUMN type;
ALTER TABLE balances DROP COLUMN user_type;
ALTER TABLE balances DROP COLUMN balance;

-- Account Transaction
ALTER TABLE account_transactions RENAME COLUMN account_id TO balance_id;
ALTER TABLE account_transactions DROP COLUMN reference;
-- +goose StatementEnd
