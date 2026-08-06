-- +goose Up
-- +goose StatementBegin
ALTER TABLE account_transactions ADD INDEX account_transactions_merchant_type_created_comp_idx(merchant_id, type, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE account_transactions DROP INDEX account_transactions_merchant_type_created_comp_idx;
-- +goose StatementEnd
