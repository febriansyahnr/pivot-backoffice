-- +goose Up
-- +goose StatementBegin
ALTER TABLE disbursements ADD COLUMN bank_reference_no VARCHAR(60) NULL AFTER processor_reference_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE disbursements DROP COLUMN bank_reference_no;
-- +goose StatementEnd
