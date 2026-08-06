-- +goose Up
-- +goose StatementBegin
ALTER TABLE `payments` MODIFY COLUMN `type` VARCHAR(30) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `payments` MODIFY COLUMN `type` VARCHAR(12) NOT NULL DEFAULT '';
-- +goose StatementEnd
