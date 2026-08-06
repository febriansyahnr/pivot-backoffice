-- +goose Up
-- +goose StatementBegin
ALTER TABLE account_transactions ADD COLUMN reason_type varchar(100) NULL AFTER status;
ALTER TABLE account_transactions ADD COLUMN reason_description varchar(255) NULL AFTER reason_type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE account_transactions DROP COLUMN reason_type;
ALTER TABLE account_transactions DROP COLUMN reason_description;
-- +goose StatementEnd
