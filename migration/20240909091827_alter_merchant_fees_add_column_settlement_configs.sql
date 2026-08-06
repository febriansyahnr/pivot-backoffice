-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_fees` ADD COLUMN `settlement_configs` JSON NULL AFTER `tax_percentage`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_fees` DROP COLUMN `settlement_configs`;
-- +goose StatementEnd
