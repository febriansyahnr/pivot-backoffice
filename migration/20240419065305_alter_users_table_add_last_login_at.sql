-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN last_login_at datetime;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN last_login_at;
-- +goose StatementEnd
