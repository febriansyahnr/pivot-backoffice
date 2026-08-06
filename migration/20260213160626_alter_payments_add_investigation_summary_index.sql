-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_payments_investigation_summary
ON payments (merchant_id, reason_type, investigation_started_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_payments_investigation_summary ON payments;
-- +goose StatementEnd
