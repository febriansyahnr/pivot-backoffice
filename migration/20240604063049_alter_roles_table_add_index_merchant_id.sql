-- +goose Up
-- +goose StatementBegin
ALTER TABLE `roles` ADD INDEX `merchant_id_idx`(`merchant_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `roles` DROP INDEX `merchant_id_idx`;
-- +goose StatementEnd
