-- +goose Up
-- +goose StatementBegin
ALTER TABLE payment_methods ADD COLUMN category VARCHAR(100) AFTER type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payment_methods DROP COLUMN category;
-- +goose StatementEnd
