-- +goose Up
-- +goose StatementBegin
ALTER TABLE account_transactions
DROP PRIMARY KEY;

ALTER TABLE account_transactions
ADD PRIMARY KEY (uuid, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE account_transactions
DROP PRIMARY KEY;

ALTER TABLE account_transactions
ADD PRIMARY KEY (uuid);
-- +goose StatementEnd
