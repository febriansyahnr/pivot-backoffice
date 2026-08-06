-- +goose Up
-- +goose StatementBegin
ALTER TABLE `inbound` ADD INDEX inbound_created_at_idx(`created_at`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `inbound` DROP INDEX inbound_created_at_idx;
-- +goose StatementEnd
