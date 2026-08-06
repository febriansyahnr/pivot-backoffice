-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments ADD INDEX payments_merchant_id_reference_id_IDX (merchant_id, reference_id);
ALTER TABLE payments DROP INDEX payments_reference_id_IDX;
ALTER TABLE payments DROP INDEX payments_merchant_id_IDX;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments ADD UNIQUE INDEX payments_reference_id_IDX (reference_id);
ALTER TABLE payments ADD INDEX payments_merchant_id_IDX (merchant_id);
ALTER TABLE payments DROP INDEX payments_merchant_id_reference_id_IDX;
-- +goose StatementEnd
