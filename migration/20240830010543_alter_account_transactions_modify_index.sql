-- +goose Up
-- +goose StatementBegin
ALTER TABLE `account_transactions` 
	DROP INDEX `idx_account_transactions_aggregate_IDX`,
    DROP INDEX `account_transactions_type_reference_id_IDX`,
	ADD INDEX `account_transactions_merchant_date_comp_idx`(`merchant_id`, `updated_at`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `account_transactions`
	DROP INDEX `account_transactions_merchant_date_comp_idx`,
	ADD INDEX `idx_account_transactions_aggregate_IDX` (`merchant_id`,`account_id`,`status`,`created_at`),
    ADD INDEX `account_transactions_type_reference_id_IDX` (`type`,`reference_id`);
-- +goose StatementEnd
