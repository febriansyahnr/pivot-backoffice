-- +goose Up
-- +goose StatementBegin
CREATE TABLE `reconciliations` (
	`uuid` varchar(36) NOT NULL,
	`file_path` varchar(255) NOT NULL,
	`result_file_path` varchar(255) NOT NULL,
	`status` varchar(12) DEFAULT 'PENDING' NOT NULL COMMENT 'PENDING,SUCCESS,FAILED',
	`reasons` text NULL,
	`created_by` varchar(100) NOT NULL,
	`created_at` TIMESTAMP NOT NULL,
	`updated_at` TIMESTAMP NOT NULL,
	CONSTRAINT reconciliations_pk PRIMARY KEY (uuid)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `reconciliations`;
-- +goose StatementEnd
