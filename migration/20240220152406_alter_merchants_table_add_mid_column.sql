-- +goose Up
-- +goose StatementBegin
ALTER TABLE merchants ADD COLUMN mid VARCHAR(4) AFTER pic_phone, ADD CONSTRAINT merchant_unique_mid UNIQUE (mid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE merchants DROP INDEX merchant_unique_mid, DROP COLUMN mid;
-- +goose StatementEnd
