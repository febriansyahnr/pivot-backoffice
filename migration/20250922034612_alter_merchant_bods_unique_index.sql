-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchant_bods` DROP INDEX `merchant_bods_merchant_position_identity_comp_uniq_idx`,
ADD INDEX `merchant_bods_merchant_position_identity_comp_idx`(`merchant_id`, `position`, `identity_number`, `deleted`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchant_bods` DROP INDEX `merchant_bods_merchant_position_identity_comp_idx`,
ADD UNIQUE INDEX `merchant_bods_merchant_position_identity_comp_uniq_idx`(`merchant_id`, `position`, `identity_number`, `deleted`);
-- +goose StatementEnd
