-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT "ACTIVE" AFTER email;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN status;
-- +goose StatementEnd
