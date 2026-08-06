-- +goose Up
-- +goose StatementBegin
CREATE TABLE recurring_contracts (
	uuid				CHAR(36) NOT NULL PRIMARY KEY,
	merchant_id			CHAR(36) NOT NULL,
	client_reference_id	VARCHAR(100) NOT NULL,
	customer_id			CHAR(36) NOT NULL,
	payment_method_id	CHAR(36) NULL,
	payment_token_id	VARCHAR(100) NULL,
	auth_method 		VARCHAR(30) NOT NULL, 	
	auth_transaction_id	CHAR(36) NULL,
	start_date			TIMESTAMP NULL,
	end_date			TIMESTAMP NOT NULL,
	plan				JSON NOT NULL,				
	trials				JSON NULL,			
	billing				JSON NOT NULL,	
	scheduler_mode		VARCHAR(30) NOT NULL,
	currency			CHAR(3) NOT NULL,
	amount				DECIMAL(18, 2) NOT NULL,
	status				VARCHAR(30) NOT NULL,
	activated_at		TIMESTAMP NULL,
	deactivated_at		TIMESTAMP NULL,
	created_by			CHAR(36) NOT NULL,
	created_at			TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_by			CHAR(36) NOT NULL,
	updated_at			TIMESTAMP NOT NULL,
	deleted_at			TIMESTAMP NULL,
	UNIQUE KEY merchant_client_reference_comp_idx (merchant_id, client_reference_id),
	KEY merchant_created_comp_idx(merchant_id, created_at)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE recurring_contracts;
-- +goose StatementEnd
