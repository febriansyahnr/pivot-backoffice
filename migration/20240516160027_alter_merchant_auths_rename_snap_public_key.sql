-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchant_auths RENAME COLUMN snap_public_key TO merchant_public_key;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchant_auths RENAME COLUMN merchant_public_key TO snap_public_key;
-- +goose StatementEnd
