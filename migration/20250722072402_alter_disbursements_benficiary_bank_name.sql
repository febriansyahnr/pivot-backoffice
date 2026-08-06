-- +goose Up
-- +goose StatementBegin
ALTER TABLE backend_portal.disbursements MODIFY COLUMN beneficiary_bank_name varchar(150);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backend_portal.disbursements MODIFY COLUMN beneficiary_bank_name varchar(60);
-- +goose StatementEnd
