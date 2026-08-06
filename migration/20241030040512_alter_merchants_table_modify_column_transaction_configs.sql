-- +goose Up
-- +goose StatementBegin
UPDATE merchants SET transaction_configs = JSON_OBJECT() WHERE transaction_configs IS NULL; 
UPDATE 
	merchants 
SET transaction_configs = JSON_INSERT(transaction_configs, '$.autoWithdrawal', 'OFF'); 
ALTER TABLE merchants 
	MODIFY COLUMN transaction_configs JSON NOT NULL DEFAULT ('{"autoWithdrawal":"ON"}');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants MODIFY COLUMN transaction_configs JSON NULL;
UPDATE 
	merchants 
SET transaction_configs = JSON_REMOVE(transaction_configs, '$.autoWithdrawal');
UPDATE merchants SET transaction_configs = NULL WHERE transaction_configs = JSON_OBJECT();
-- +goose StatementEnd
