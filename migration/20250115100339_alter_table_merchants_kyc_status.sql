-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants ADD COLUMN kyc_status VARCHAR(50) AFTER status;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants DROP COLUMN kyc_status;
-- +goose StatementEnd
