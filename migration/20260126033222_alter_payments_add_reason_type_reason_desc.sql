-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments
ADD COLUMN reason_type VARCHAR(50) NULL AFTER status,
ADD COLUMN reason_description TEXT NULL AFTER reason_type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments
DROP COLUMN reason_type,
DROP COLUMN reason_description;
-- +goose StatementEnd
