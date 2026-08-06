-- +goose Up
-- +goose StatementBegin
ALTER TABLE industries ADD COLUMN deleted_at TIMESTAMP(6) NULL DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE industries DROP COLUMN deleted_at;
-- +goose StatementEnd
