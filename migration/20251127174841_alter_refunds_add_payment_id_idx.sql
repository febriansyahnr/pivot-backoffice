-- +goose Up
-- +goose StatementBegin
ALTER TABLE `refunds` ADD INDEX `refunds_payment_id_idx`(`payment_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `refunds` DROP INDEX `refunds_payment_id_idx`;
-- +goose StatementEnd
