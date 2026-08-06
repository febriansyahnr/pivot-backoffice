-- +goose Up
-- +goose StatementBegin
-- Remove: NO_ZERO_IN_DATE and NO_ZERO_DATE
SET SESSION sql_mode = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION';
ALTER TABLE accounts MODIFY COLUMN pending_transaction_cutoff_at TIMESTAMP(6) NULL DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts MODIFY COLUMN pending_transaction_cutoff_at TIMESTAMP NULL DEFAULT NULL;
-- +goose StatementEnd
