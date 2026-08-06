-- +goose Up
-- +goose StatementBegin
ALTER TABLE `payments` ADD COLUMN `created_by` VARCHAR(100) NULL AFTER `payment_url`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `payments` DROP COLUMN `created_by`;
-- +goose StatementEnd
