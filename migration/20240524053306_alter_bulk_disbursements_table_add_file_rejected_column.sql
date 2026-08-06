-- +goose Up
-- +goose StatementBegin
ALTER TABLE bulk_disbursements ADD COLUMN file_rejected varchar(255) NULL AFTER file_failed;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE bulk_disbursements DROP COLUMN file_rejected;
-- +goose StatementEnd
