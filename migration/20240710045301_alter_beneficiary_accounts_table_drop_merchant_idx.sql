-- +goose Up
-- +goose StatementBegin
ALTER TABLE `beneficiary_accounts` 
	DROP FOREIGN KEY `beneficiary_accounts_ibfk_1`,
	DROP INDEX `beneficiary_accounts_merchant_id_IDX`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `beneficiary_accounts` 
	ADD CONSTRAINT `beneficiary_accounts_ibfk_1` FOREIGN KEY (`merchant_id`) REFERENCES `merchants` (`uuid`),
	ADD INDEX `beneficiary_accounts_merchant_id_IDX`(`merchant_id`);
-- +goose StatementEnd
