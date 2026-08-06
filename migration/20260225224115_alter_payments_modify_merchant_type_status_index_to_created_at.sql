-- +goose Up
-- +goose StatementBegin
DROP INDEX payments_merchant_type_status_idx ON payments;
CREATE INDEX payments_merchant_type_created_at_idx ON payments (merchant_id, type, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX payments_merchant_type_created_at_idx ON payments;
CREATE INDEX payments_merchant_type_status_idx ON payments (merchant_id, type, status);
-- +goose StatementEnd
