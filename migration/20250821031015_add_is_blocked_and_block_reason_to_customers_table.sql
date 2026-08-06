-- +goose Up
-- +goose StatementBegin
ALTER TABLE customers 
ADD COLUMN is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN block_reason TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customers 
DROP COLUMN is_blocked,
DROP COLUMN block_reason;
-- +goose StatementEnd
