-- +goose Up
-- +goose StatementBegin
ALTER TABLE customers 
	MODIFY COLUMN updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customers 
	MODIFY COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
