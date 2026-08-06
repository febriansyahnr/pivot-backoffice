-- +goose Up
-- +goose StatementBegin
DROP INDEX transfers_merchant_reference_id_comp_idx ON transfers;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE UNIQUE INDEX transfers_merchant_reference_id_comp_idx ON transfers (`merchant_id`,`reference_id`) USING BTREE;
-- +goose StatementEnd
