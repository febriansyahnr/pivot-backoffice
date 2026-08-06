-- +goose Up
-- +goose StatementBegin
ALTER TABLE `beneficiary_accounts` 
	ADD UNIQUE INDEX `beneficiary_accounts_merchant_bank_code_account_no_comp_uniq_idx`(`merchant_id`,`beneficiary_bank_code`,`beneficiary_account_no`,`deleted_at`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `beneficiary_accounts` 
	DROP INDEX `beneficiary_accounts_merchant_bank_code_account_no_comp_uniq_idx`;
-- +goose StatementEnd
