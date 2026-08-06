-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_merchants_parent_id ON merchants(parent_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_merchants_parent_id ON merchants;
-- +goose StatementEnd
