-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants ADD COLUMN fds_configs JSON NULL AFTER transaction_configs;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants DROP COLUMN fds_configs;
-- +goose StatementEnd
