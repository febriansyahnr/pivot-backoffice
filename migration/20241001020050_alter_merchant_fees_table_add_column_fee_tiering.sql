-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_fees` 
	ADD COLUMN `tiering_type` VARCHAR(20) NULL AFTER `settlement_configs`,
	ADD COLUMN `tiering_configs` JSON NULL AFTER `tiering_type`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_fees`
	DROP COLUMN `tiering_type`, DROP COLUMN `tiering_configs`;
-- +goose StatementEnd
