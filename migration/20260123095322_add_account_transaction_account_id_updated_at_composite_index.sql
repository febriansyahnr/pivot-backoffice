-- +goose Up
-- +goose StatementBegin
ALTER TABLE `account_transactions` ADD INDEX `account_transactions_account_id_updated_at_comp_idx`(`account_id`, `updated_at`), ALGORITHM=INPLACE, LOCK=NONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `account_transactions` DROP INDEX `account_transactions_account_id_updated_at_comp_idx`, ALGORITHM=INPLACE, LOCK=NONE; 
-- +goose StatementEnd
