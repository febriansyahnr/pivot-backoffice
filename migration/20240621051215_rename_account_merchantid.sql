-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts DROP COLUMN balance;
ALTER TABLE accounts RENAME COLUMN merchant_id TO reference_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts ADD COLUMN balance decimal(19,4);
ALTER TABLE accounts RENAME COLUMN reference_id TO merchant_id ;
-- +goose StatementEnd
