-- +goose Up
-- +goose StatementBegin
ALTER TABLE backend_portal.beneficiary_accounts MODIFY COLUMN beneficiary_bank_name varchar(150);
ALTER TABLE backend_portal.beneficiary_accounts MODIFY COLUMN beneficiary_account_name varchar(150);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backend_portal.beneficiary_accounts MODIFY COLUMN beneficiary_bank_name varchar(60);
ALTER TABLE backend_portal.beneficiary_accounts MODIFY COLUMN beneficiary_account_name varchar(60);
-- +goose StatementEnd