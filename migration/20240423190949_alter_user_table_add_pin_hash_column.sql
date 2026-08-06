-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN pin_hash varchar(255) NULL AFTER password;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN pin_hash;
-- +goose StatementEnd
