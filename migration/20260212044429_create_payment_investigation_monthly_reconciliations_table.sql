-- +goose Up
-- +goose StatementBegin
CREATE TABLE payment_investigation_monthly_reconciliations(
	uuid						VARCHAR(36) NOT NULL PRIMARY KEY,
	`date`						TIMESTAMP NOT NULL,
	merchant_id 				VARCHAR(36) NOT NULL,
	payment_ids 				JSON NOT NULL,
	payment_count				INTEGER UNSIGNED NOT NULL,
	gross_amount 				DECIMAL(24, 2) NOT NULL,
	fee_amount					DECIMAL(18, 2) NOT NULL,
	net_amount 					DECIMAL(24, 2) NOT NULL,
	platform_loss_percentage 	DECIMAL(5, 2) NOT NULL,
	platform_max_loss			DECIMAL(12, 2) NOT NULL,
	platform_loss_amount		DECIMAL(24, 2) NOT NULL,
	merchant_loss_amount        DECIMAL(24, 2) NOT NULL,
	created_at					TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY merchant_date_comp_idx (merchant_id, date)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE payment_investigation_monthly_reconciliations;
-- +goose StatementEnd
