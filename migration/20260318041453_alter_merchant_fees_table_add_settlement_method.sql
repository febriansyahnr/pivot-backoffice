-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_fees`
    ADD COLUMN `settlement_method` VARCHAR(36) NULL DEFAULT NULL AFTER `settlement_model`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_fees DROP COLUMN `settlement_method`;
-- +goose StatementEnd
