-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchants` 
    ADD COLUMN `risk_level` VARCHAR(60) NULL AFTER `status`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchants` DROP COLUMN `risk_level`;
-- +goose StatementEnd