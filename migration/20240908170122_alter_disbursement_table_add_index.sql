-- +goose Up
-- +goose StatementBegin
ALTER TABLE `disbursements` 
ADD INDEX `disbursements_processor_reference_id_IDX` (`processor_reference_id`) USING BTREE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `disbursements` DROP INDEX `disbursements_processor_reference_id_IDX`;
-- +goose StatementEnd
