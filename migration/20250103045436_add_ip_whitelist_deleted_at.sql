-- +goose Up
-- +goose StatementBegin
ALTER TABLE ip_whitelist_configurations ADD deleted_at datetime NULL;
ALTER TABLE `ip_whitelist_configurations` ADD INDEX `ip_whitelist_merchant_id_IDX` (`merchant_id`) USING BTREE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ip_whitelist_configurations DROP COLUMN deleted_at ;
ALTER TABLE `ip_whitelist_configurations` DROP INDEX `ip_whitelist_merchant_id_IDX`;
-- +goose StatementEnd
