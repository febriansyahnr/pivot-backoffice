-- +goose Up
-- +goose StatementBegin
ALTER TABLE `disbursements` ADD COLUMN `metadata` JSON NULL AFTER `remark`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `disbursements` DROP COLUMN `metadata`;
-- +goose StatementEnd
