-- +goose Up
-- +goose StatementBegin
ALTER TABLE backend_portal.activity_logs MODIFY COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backend_portal.activity_logs MODIFY COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
