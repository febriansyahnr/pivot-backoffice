-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchants` ADD COLUMN `status` VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' AFTER `transaction_configs`;
ALTER TABLE `merchants` ADD COLUMN `reason_status` VARCHAR(255) NOT NULL DEFAULT '' AFTER `status`;

ALTER TABLE `merchants` ADD INDEX `merchants_status_idx` (`status`) USING BTREE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchants` DROP INDEX `merchants_status_idx`;
ALTER TABLE `merchants` DROP COLUMN `reason_status`;
ALTER TABLE `merchants` DROP COLUMN `status`;
-- +goose StatementEnd