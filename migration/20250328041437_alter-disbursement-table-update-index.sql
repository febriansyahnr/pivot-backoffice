-- +goose Up
-- +goose StatementBegin
ALTER TABLE `disbursements` DROP INDEX `disbursement_unique_reference_per_merchant` ,
DROP INDEX `disbursements_created_at_IDX` ,
DROP INDEX `disbursements_merchant_id_IDX`,
ADD UNIQUE INDEX `disbursement_merchant_reference_uniq_idx` (`merchant_id`, `reference_id`),
ADD INDEX `disbursement_merchant_id_created_at_idx` (`merchant_id`, `created_at`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `disbursements`
DROP INDEX `disbursement_merchant_reference_uniq_idx`,
DROP INDEX `disbursement_merchant_id_created_at_idx`,
ADD UNIQUE INDEX `disbursement_unique_reference_per_merchant` (`reference_id`, `merchant_id`),
ADD INDEX `disbursements_created_at_IDX` (`created_at`),
ADD INDEX `disbursements_merchant_id_IDX` (`merchant_id`);
-- +goose StatementEnd
