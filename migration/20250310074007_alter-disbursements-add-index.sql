-- +goose Up
-- +goose StatementBegin
CREATE INDEX disbursements_bank_code_account_no_updatedat_comp_idx ON disbursements (beneficiary_bank_code, beneficiary_account_no, updated_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX disbursements_bank_code_account_no_updatedat_comp_idx ON disbursements;
-- +goose StatementEnd
