-- +goose Up
-- +goose StatementBegin
ALTER TABLE vendors
ADD COLUMN not_archived BOOLEAN
GENERATED ALWAYS AS (IF(deleted_at IS NULL, 1, NULL)) VIRTUAL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE vendors ADD UNIQUE INDEX vendors_merchant_name_uniq_comp_idx(merchant_id, name, not_archived);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vendors DROP INDEX vendors_merchant_name_uniq_comp_idx;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE vendors DROP COLUMN not_archived;
-- +goose StatementEnd
