-- +goose Up
-- +goose StatementBegin
ALTER TABLE disbursements ADD COLUMN account_inquiry_id VARCHAR(36) NULL AFTER sender_name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE disbursements DROP COLUMN account_inquiry_id;
-- +goose StatementEnd
