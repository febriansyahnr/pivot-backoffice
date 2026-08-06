-- +goose Up
-- +goose StatementBegin
ALTER TABLE `manual_adjustment_histories` ADD INDEX `manual_adjustment_histories_reference_id_IDX`(`reference_id`);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `manual_adjustment_histories` DROP INDEX `manual_adjustment_histories_reference_id_IDX`;
-- +goose StatementEnd
