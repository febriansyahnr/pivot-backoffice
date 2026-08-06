-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants ADD COLUMN metadata JSON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants DROP COLUMN metadata;
-- +goose StatementEnd
