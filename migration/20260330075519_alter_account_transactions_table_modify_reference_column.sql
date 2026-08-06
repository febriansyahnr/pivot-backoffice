-- +goose Up
-- +goose StatementBegin
ALTER TABLE account_transactions MODIFY COLUMN reference VARCHAR(60) DEFAULT NULL, ALGORITHM=INPLACE, LOCK=NONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE account_transactions MODIFY COLUMN reference VARCHAR(20) DEFAULT NULL;
-- +goose StatementEnd
