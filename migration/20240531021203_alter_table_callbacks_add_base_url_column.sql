-- +goose Up
-- +goose StatementBegin
ALTER TABLE callbacks
ADD COLUMN base_url VARCHAR(100) NULL AFTER merchant_id;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE callbacks
DROP COLUMN base_url;

-- +goose StatementEnd