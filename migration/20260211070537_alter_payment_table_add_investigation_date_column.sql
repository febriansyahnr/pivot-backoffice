-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments ADD COLUMN `investigation_started_at` TIMESTAMP NULL DEFAULT NULL AFTER `type`,
	ADD COLUMN `investigation_completed_at` TIMESTAMP NULL DEFAULT NULL AFTER `investigation_started_at`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments DROP COLUMN `investigation_started_at`,
	DROP COLUMN `investigation_completed_at`;
-- +goose StatementEnd
