-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS manual_adjustment_histories (
	`uuid`					VARCHAR(36) NOT NULL PRIMARY KEY,
	`merchant_id`			VARCHAR(36) NOT NULL,
    `transaction_date`      TIMESTAMP NOT NULL,
	`bank_reference_id`		VARCHAR(100) NULL,
	`bank_account`			JSON NOT NULL,
	`type`					VARCHAR(30) NOT NULL,
	`action`				VARCHAR(30) NOT NULL,
	`currency`				VARCHAR(3) NOT NULL,
	`amount`				DECIMAL(19,4) NOT NULL,
	`reference_id`			VARCHAR(36) NOT NULL,
	`proof_of_transfer`		VARCHAR(200) NOT NULL,
	`notes`					VARCHAR(200) NOT NULL,
	`created_by`			VARCHAR(100) NOT NULL,
	`created_at`			TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	`updated_at`			TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
	`deleted_at`			TIMESTAMP NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS manual_adjustment_histories;
-- +goose StatementEnd
