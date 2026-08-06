-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles ADD COLUMN merchant_id VARCHAR(100) NULL AFTER uuid;
ALTER TABLE roles ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT 'DEFAULT' AFTER slug;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles DROP COLUMN type;
ALTER TABLE roles DROP COLUMN merchant_id;
-- +goose StatementEnd
