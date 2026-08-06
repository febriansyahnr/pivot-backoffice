-- +goose Up
-- +goose StatementBegin
ALTER TABLE `payments` ADD COLUMN `type` VARCHAR(12) NOT NULL DEFAULT '' AFTER `status`;
CREATE INDEX payments_merchant_type_status_idx ON payments (merchant_id, type, status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX payments_merchant_type_status_idx ON payments;
ALTER TABLE `payments` DROP COLUMN `type`;
-- +goose StatementEnd
