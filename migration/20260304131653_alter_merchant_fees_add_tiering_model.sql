-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_fees`
  ADD COLUMN `tiering_model` VARCHAR(30) NULL AFTER `settlement_configs`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_fees`
  DROP COLUMN `tiering_model`;
-- +goose StatementEnd
