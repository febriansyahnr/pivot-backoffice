-- +goose Up
-- +goose StatementBegin
ALTER TABLE `merchants` ADD COLUMN `third_party_screening_data` JSON NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `merchants` DROP COLUMN `third_party_screening_data`;
-- +goose StatementEnd
