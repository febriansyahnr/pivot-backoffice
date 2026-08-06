-- +goose Up
-- +goose StatementBegin
ALTER TABLE `menus`
    ADD COLUMN `allowed_products` JSON NULL DEFAULT NULL AFTER `parent_id`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `menus` DROP COLUMN `allowed_products`;
-- +goose StatementEnd
