-- +goose Up
-- +goose StatementBegin
ALTER TABLE `account_inquiries` ADD UNIQUE INDEX `account_inquiries_code_account_no_comp_uniq_idx` (`beneficiary_bank_code`, `beneficiary_account_no`,`deleted_at`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `account_inquiries` DROP INDEX `account_inquiries_code_account_no_comp_uniq_idx`;
-- +goose StatementEnd
