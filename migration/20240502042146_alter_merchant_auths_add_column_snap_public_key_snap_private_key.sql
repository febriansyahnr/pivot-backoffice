-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchant_auths ADD COLUMN snap_public_key TEXT NULL AFTER secret;
ALTER TABLE merchant_auths ADD COLUMN snap_private_key TEXT NULL AFTER snap_public_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_auths DROP COLUMN snap_public_key;
ALTER TABLE merchant_auths DROP COLUMN snap_private_key;
-- +goose StatementEnd
