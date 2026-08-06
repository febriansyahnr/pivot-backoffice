-- +goose Up
-- +goose StatementBegin
ALTER TABLE `reconciliations` ADD COLUMN `transaction_type` VARCHAR(32) NOT NULL DEFAULT 'PAYMENT' AFTER `status`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `reconciliations` DROP COLUMN `transaction_type`;
-- +goose StatementEnd
