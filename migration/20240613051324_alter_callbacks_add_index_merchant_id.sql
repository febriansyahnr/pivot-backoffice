-- +goose Up
-- +goose StatementBegin
ALTER TABLE `callbacks` ADD INDEX `callbacks_merchant_id_idx`(`merchant_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `callbacks` DROP INDEX `callbacks_merchant_id_idx`;
-- +goose StatementEnd
