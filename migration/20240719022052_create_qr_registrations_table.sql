-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `qr_registrations` (
	id							VARCHAR(36) NOT NULL PRIMARY KEY,
	external_id 				VARCHAR(32) NOT NULL,
	acquirer					VARCHAR(30) NOT NULL,
	merchant_type				VARCHAR(50) NOT NULL,
	acquirer_parent_merchant_id VARCHAR(12) NULL,
	merchant_name				VARCHAR(50) NOT NULL,
	merchant_short_name			VARCHAR(50) NOT NULL,
	address						JSON 		NOT NULL,
	business_info				JSON		NOT NULL,
	business_document 			JSON		NOT NULL,
	bod_info					JSON		NOT NULL,
	boc_info					JSON		NOT NULL,
	status						VARCHAR(50) NOT NULL DEFAULT 'FILLING_FORM',
	acquirer_merchant_id 		VARCHAR(12) NULL,
	created_at					TIMESTAMP 	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	created_by					VARCHAR(50) NOT NULL,
	updated_at					TIMESTAMP 	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	KEY `qr_registrations_external_id_idx`(`external_id`)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `qr_registrations`;
-- +goose StatementEnd
