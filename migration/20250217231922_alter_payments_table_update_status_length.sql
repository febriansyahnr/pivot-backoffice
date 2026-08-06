-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments
    MODIFY COLUMN status VARCHAR(40);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments
    MODIFY COLUMN status VARCHAR(20);
-- +goose StatementEnd
