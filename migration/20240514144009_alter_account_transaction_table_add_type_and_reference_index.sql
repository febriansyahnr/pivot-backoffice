-- +goose Up
-- +goose StatementBegin
ALTER TABLE `account_transactions` ADD INDEX `account_transactions_type_reference_id_IDX`(`type`, `reference_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `account_transactions` DROP INDEX `account_transactions_type_reference_id_IDX`;
-- +goose StatementEnd
