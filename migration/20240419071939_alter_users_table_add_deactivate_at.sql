-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN deactivate_at datetime;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN deactivate_at;
-- +goose StatementEnd