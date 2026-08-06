-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchants` 
	ADD COLUMN `external_id` VARCHAR(26) NOT NULL AFTER `uuid`,
	ADD COLUMN `short_name`  VARCHAR(30) NOT NULL AFTER `name`,
	ADD COLUMN `address`  	 VARCHAR(254) NOT NULL AFTER `description`,
	ADD COLUMN `district_id` SMALLINT NOT NULL AFTER `address`,
	ADD COLUMN `postcode` 	 VARCHAR(20) NOT NULL AFTER `district_id`,
	ADD INDEX `merchants_external_id_idx`(`external_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchants`
	DROP COLUMN `external_id`,
	DROP COLUMN `short_name`,
	DROP COLUMN `address`,
	DROP COLUMN `district_id`,
	DROP COLUMN `postcode`,
	DROP INDEX `merchants_external_id_idx`;
-- +goose StatementEnd
