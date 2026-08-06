-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments ADD COLUMN created_from VARCHAR(30) DEFAULT '' AFTER created_by;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments DROP COLUMN created_from;
-- +goose StatementEnd
