-- +goose Up
-- +goose StatementBegin
ALTER TABLE payment_methods ADD COLUMN sub_type VARCHAR(255) DEFAULT '' AFTER `type`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payment_methods DROP COLUMN sub_type;
-- +goose StatementEnd
