-- +goose Up
-- +goose StatementBegin
ALTER TABLE account_transactions CHANGE COLUMN `type` `type` VARCHAR(20) NOT NULL ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE account_transactions CHANGE COLUMN `type` `type` VARCHAR(13) NOT NULL ;
-- +goose StatementEnd

