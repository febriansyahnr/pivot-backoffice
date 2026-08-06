-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments ADD COLUMN reference_id varchar(100) null AFTER uuid;
ALTER TABLE payments ADD UNIQUE INDEX payments_reference_id_IDX (reference_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments DROP COLUMN reference_id;
-- +goose StatementEnd
