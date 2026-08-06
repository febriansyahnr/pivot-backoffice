-- +goose Up
-- +goose StatementBegin
ALTER TABLE customers 
ADD COLUMN phone_country_code VARCHAR(5) DEFAULT NULL AFTER email;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customers
DROP COLUMN phone_country_code;
-- +goose StatementEnd
