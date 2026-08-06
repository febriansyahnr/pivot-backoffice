-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants DROP COLUMN active;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants ADD COLUMN active BOOLEAN DEFAULT TRUE;
-- +goose StatementEnd
