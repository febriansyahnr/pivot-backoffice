-- +goose Up
-- +goose StatementBegin
CREATE INDEX transfers_comp_merchant_id_created_at_status_idx on transfers (merchant_id, created_at, status);
CREATE INDEX transfers_comp_recipient_id_created_at_status_idx on transfers (recipient_id, created_at, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transfers DROP INDEX transfers_comp_merchant_id_created_at_status_idx;
ALTER TABLE transfers DROP INDEX transfers_comp_recipient_id_created_at_status_idx;
-- +goose StatementEnd
