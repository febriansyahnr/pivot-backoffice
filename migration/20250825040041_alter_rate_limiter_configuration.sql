-- +goose Up
-- +goose StatementBegin
ALTER TABLE rate_limit_configurations ADD COLUMN http_method VARCHAR(10) DEFAULT '' AFTER `time`;
ALTER TABLE rate_limit_configurations ADD COLUMN burst INT DEFAULT 0 AFTER `limit`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rate_limit_configurations DROP COLUMN http_method;
ALTER TABLE rate_limit_configurations DROP COLUMN burst;
-- +goose StatementEnd
