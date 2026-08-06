-- +goose Up
-- +goose StatementBegin
ALTER TABLE backend_portal.request_account_inquiries MODIFY COLUMN beneficiary_bank_code varchar(10)  NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backend_portal.request_account_inquiries MODIFY COLUMN beneficiary_bank_code varchar(5) NOT NULL;
-- +goose StatementEnd
