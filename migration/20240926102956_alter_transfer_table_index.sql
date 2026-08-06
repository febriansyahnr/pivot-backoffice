-- +goose Up
-- +goose StatementBegin

DROP INDEX payments_merchant_id_composite_IDX ON transfers;
ALTER TABLE transfers ADD PRIMARY KEY(uuid);
CREATE UNIQUE INDEX transfers_merchant_reference_id_comp_idx ON transfers (`merchant_id`,`reference_id`) USING BTREE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX transfers_merchant_reference_id_comp_idx ON transfers;
ALTER TABLE transfers DROP PRIMARY KEY;
CREATE INDEX payments_merchant_id_composite_IDX ON transfers (`merchant_id`, `uuid`, `reference_id`) USING BTREE;

-- +goose StatementEnd
