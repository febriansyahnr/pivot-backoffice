-- +goose Up
-- +goose StatementBegin
ALTER TABLE `outbound` MODIFY COLUMN `request_id` VARCHAR(255) AS (client->>'$.request_id') STORED,
    ADD INDEX outbound_date_idx(`date`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `outbound` MODIFY COLUMN `request_id` VARCHAR(36) AS (client->>'$.request_id') STORED,
    DROP INDEX outbound_date_idx;
-- +goose StatementEnd
