-- +goose Up
-- +goose StatementBegin
ALTER TABLE callback_masters
    ADD COLUMN visibility VARCHAR(12) NOT NULL DEFAULT 'PUBLIC' COMMENT 'PUBLIC/RESTRICTED' AFTER description,
    ADD COLUMN whitelisted_merchant_ids JSON NULL AFTER visibility;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE callback_masters
    DROP COLUMN visibility,
    DROP COLUMN whitelisted_merchant_ids;
-- +goose StatementEnd
