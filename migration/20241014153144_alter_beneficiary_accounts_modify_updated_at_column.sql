-- +goose Up
-- +goose StatementBegin
ALTER TABLE beneficiary_accounts 
	MODIFY COLUMN updated_at DATETIME(3) NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE beneficiary_accounts 
	MODIFY COLUMN updated_at DATETIME NOT NULL;
-- +goose StatementEnd
