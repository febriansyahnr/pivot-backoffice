-- +goose Up
-- +goose StatementBegin
CREATE INDEX account_transactions_updated_at_idx USING BTREE ON backend_portal.account_transactions (updated_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backend_portal.account_transactions DROP INDEX account_transactions_updated_at_idx;
-- +goose StatementEnd
