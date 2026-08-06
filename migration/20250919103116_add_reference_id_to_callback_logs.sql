-- +goose Up
-- +goose StatementBegin
ALTER TABLE callback_logs ADD COLUMN reference_id VARCHAR(255) DEFAULT NULL AFTER uuid;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE callback_logs DROP COLUMN reference_id;
-- +goose StatementEnd
