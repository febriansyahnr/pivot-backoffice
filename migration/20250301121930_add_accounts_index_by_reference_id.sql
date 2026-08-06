-- +goose Up
-- +goose StatementBegin
CREATE INDEX accounts_reference_id_idx ON accounts (reference_id) USING BTREE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts DROP INDEX accounts_reference_id_idx;
-- +goose StatementEnd
