-- +goose Up
-- +goose StatementBegin
ALTER TABLE `request_account_inquiries` ADD INDEX `request_account_inquiries_merchant_inquiry_id_comp_idx`(`merchant_id`, `account_inquiry_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `request_account_inquiries` DROP INDEX `request_account_inquiries_merchant_inquiry_id_comp_idx`;
-- +goose StatementEnd
